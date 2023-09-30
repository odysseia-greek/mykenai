package command

import (
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	kubernetes "github.com/odysseia-greek/thales"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func New() *cobra.Command {
	var (
		namespace string
		kubePath  string
	)
	cmd := &cobra.Command{
		Use:   "new",
		Short: "adds the full flow to vault",
		Long:  `inits, unseals vault and adds both policies and auth method`,
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

			NewVaultFlow(namespace, kubeManager)

		},
	}

	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace defaults to odysseia")
	cmd.PersistentFlags().StringVarP(&kubePath, "kubepath", "k", "", "kubeconfig filepath defaults to ~/.kube/config")

	return cmd
}

func NewVaultFlow(namespace string, kube kubernetes.KubeClient) (*ClusterKeys, error) {
	vaultConfig := determineMode(namespace, kube)

	glg.Info("1. vault init started")
	clusterKeys, err := initVault(namespace, kube, vaultConfig)
	if err != nil {
		return nil, err
	}

	glg.Info("1. vault init completed")
	if vaultConfig.HaEnabled {
		glg.Info("2. vault unseal started for an HA cluster")

	} else {
		glg.Info("2. vault unseal started")
	}

	err = UnsealVault(clusterKeys.UnsealKeysHex[0], namespace, kube, vaultConfig)
	if err != nil {
		return nil, err
	}

	glg.Info("2. vault unseal completed")
	glg.Info("2b. creating secret engine")
	glg.Debug("waiting 2 seconds for sealed status to be removed")
	time.Sleep(2000 * time.Millisecond)
	enableSecrets(namespace, "configs", clusterKeys.RootToken, kube)
	glg.Info("3. adding admin")
	createPolicy(defaultAdminPolicyName, namespace, clusterKeys.RootToken, kube)
	glg.Info("3. finished adding admin")
	glg.Info("4. adding user")
	createPolicy(defaultUserPolicyName, namespace, clusterKeys.RootToken, kube)
	glg.Info("4. finished adding user")
	glg.Info("5. adding kuberentes as auth method")
	enableKubernetesAsAuth(namespace, defaultAdminPolicyName, clusterKeys.RootToken, kube)
	glg.Info("5. finished adding kuberentes as auth method")

	return clusterKeys, nil
}
