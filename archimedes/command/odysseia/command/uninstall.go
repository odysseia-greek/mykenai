package command

import (
	"github.com/kpango/glg"
	kubernetes "github.com/odysseia-greek/agora/thales"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	"github.com/odysseia-greek/mykenai/archimedes/util/helm"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Uninstall() *cobra.Command {
	var (
		namespace string
		kubePath  string
	)
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "uninstall odysseia",
		Long:  `Allows you to uninstall odysseia`,
		Run: func(cmd *cobra.Command, args []string) {

			if namespace == "" {
				glg.Debugf("defaulting to %s", command.DefaultNamespace)
				namespace = command.DefaultNamespace
			}

			if kubePath == "" {
				glg.Debugf("defaulting to %s", command.DefaultKubeConfig)
				homeDir, err := os.UserHomeDir()
				if err != nil {
					glg.Error(err)
				}

				kubePath = filepath.Join(homeDir, command.DefaultKubeConfig)
			}

			cfg, err := ioutil.ReadFile(kubePath)
			if err != nil {
				glg.Error("error getting kubeconfig")
			}

			kubeManager, err := kubernetes.NewKubeClient(cfg, namespace)
			if err != nil {
				glg.Fatal("error creating kubeclient")
			}

			helmManager, err := helm.NewHelmClient(cfg, namespace)
			if err != nil {
				glg.Fatal("error creating helmclient")
			}

			err = uninstall(helmManager, kubeManager, namespace)
			if err != nil {
				glg.Error(err)
				os.Exit(1)
			}
		},
	}
	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace defaults to odysseia")
	cmd.PersistentFlags().StringVarP(&kubePath, "kubepath", "k", "", "kubeconfig filepath defaults to ~/.kube/config")

	return cmd
}

func uninstall(helm helm.HelmClient, kube kubernetes.KubeClient, ns string) error {
	allApps, err := helm.List()
	if err != nil {
		return err
	}

	for _, app := range allApps {
		if app.Namespace == ns {
			if app.Name == "cloudflare-tunnel" {
				continue
			}
			helm.Uninstall(app.Name)
		}
	}

	err = waitForPodsToTerminate(kube, ns)

	pvcs, err := kube.Storage().ListPvc(ns)
	for _, pvc := range pvcs.Items {
		glg.Debugf("removing pvc: %s", pvc.Name)
		err = kube.Storage().DeletePvc(ns, pvc.Name)
		if err != nil {
			glg.Error(err)
		}
	}

	secrets, err := kube.Configuration().ListSecrets(ns)
	for _, secret := range secrets.Items {
		if strings.Contains(secret.Name, "cloudflare") {
			continue
		}
		glg.Debugf("removing secret: %s", secret.Name)
		err = kube.Configuration().DeleteSecret(ns, secret.Name)
		if err != nil {
			glg.Error(err)
		}
	}

	glg.Info("odysseia install removed without a trace")

	return nil
}

func waitForPodsToTerminate(kube kubernetes.KubeClient, namespace string) error {
	for {
		pods, err := kube.Workload().List(namespace)
		if err != nil {
			return err
		}

		if len(pods.Items) == 1 {
			if strings.Contains(pods.Items[0].Name, "cloudflare") {
				return nil
			}
			continue
		}

		glg.Debug("waiting for pods to shut down")
		time.Sleep(2500 * time.Millisecond)
	}
}
