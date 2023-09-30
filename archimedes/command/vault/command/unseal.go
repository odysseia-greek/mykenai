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

func Unseal() *cobra.Command {
	var (
		key       string
		namespace string
		kubePath  string
	)
	cmd := &cobra.Command{
		Use:   "unseal",
		Short: "Unseal your vault",
		Long: `Allows you unseal the vault, it takes
- Key`,
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

			glg.Info("is it secret? Is it safe? Well no longer!")
			glg.Debug("unsealing kube vault")

			config := determineMode(namespace, kubeManager)
			err = UnsealVault(key, namespace, kubeManager, config)
			if err != nil {
				glg.Error(err)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&key, "key", "u", "", "unseal key, if not set cluster-keys will be used")
	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace defaults to odysseia")
	cmd.PersistentFlags().StringVarP(&kubePath, "kubepath", "k", "", "kubeconfig filepath defaults to ~/.kube/config")

	return cmd
}

func UnsealVault(key, namespace string, kube kubernetes.KubeClient, config *vaultConfig) error {
	if key == "" {
		glg.Info("key was not given, trying to get key from cluster-keys.json")
		clusterKeys, err := command.GetClusterKeys()
		if err != nil {
			glg.Fatal("could not get cluster keys")
		}
		key = clusterKeys.VaultUnsealKey
		glg.Info("key found")
	}

	vaultOperator := []string{"vault", "operator", "unseal", key}

	vaultUnsealed, err := kube.Workload().ExecNamedPod(namespace, config.PrimaryNode, vaultOperator)
	if err != nil {
		return err
	}

	glg.Info(vaultUnsealed)

	if config.HaEnabled {
		err := unsealVaultHA(vaultOperator, namespace, kube, config)
		if err != nil {
			return err
		}
	}

	return nil
}

func unsealVaultHA(initCommand []string, namespace string, kube kubernetes.KubeClient, config *vaultConfig) error {
	raftJoinCommand := []string{"vault", "operator", "raft", "join", "-leader-ca-cert=@/vault/userconfig/vault-server-tls/vault.ca", fmt.Sprintf("https://%s.vault-internal:8200", config.PrimaryNode)}

	for _, node := range config.SecondaryNodes {
		joinedRaft, err := kube.Workload().ExecNamedPod(namespace, node, raftJoinCommand)
		if err != nil {
			return err
		}
		glg.Info(joinedRaft)

		vaultUnsealed, err := kube.Workload().ExecNamedPod(namespace, node, initCommand)
		if err != nil {
			return err
		}
		glg.Info(vaultUnsealed)
	}

	return nil
}
