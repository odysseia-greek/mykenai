package command

import (
	"fmt"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	kubernetes "github.com/odysseia-greek/thales"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Init() *cobra.Command {
	var (
		namespace string
		kubePath  string
	)
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Inits your vault",
		Long: `Allows you to init the vault, it takes
- Namespace
- Filepath`,
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

			config := determineMode(namespace, kubeManager)

			glg.Info("is it secret? Is it safe? Well no longer!")
			glg.Debug("unsealing kube vault")
			initVault(namespace, kubeManager, config)
		},
	}

	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace defaults to odysseia")
	cmd.PersistentFlags().StringVarP(&kubePath, "kubepath", "k", "", "kubeconfig filepath defaults to ~/.kube/config")

	return cmd
}

func initVault(namespace string, kube kubernetes.KubeClient, config *vaultConfig) (*ClusterKeys, error) {
	vaultCommand := []string{"vault", "operator", "init", "-key-shares=1", "-key-threshold=1", "-format=json"}

	vaultInit, err := kube.Workload().ExecNamedPod(namespace, config.PrimaryNode, vaultCommand)
	if err != nil {
		return nil, err
	}
	clusterKeys, err := UnmarshalClusterKeys([]byte(vaultInit))
	if err != nil {
		return nil, err
	}
	glg.Debug(fmt.Sprintf("%s initiated", config.PrimaryNode))

	return &clusterKeys, nil
}
