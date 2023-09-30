package command

import (
	"embed"
	"fmt"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	kubernetes "github.com/odysseia-greek/thales"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const defaultAdminPolicyName = "solon"
const defaultUserPolicyName = "ptolemaios"

var (
	//go:embed hcl/solon-acl.hcl hcl/ptolemaios-acl.hcl
	res embed.FS

	policies = map[string]string{
		defaultAdminPolicyName: "hcl/solon-acl.hcl",
		defaultUserPolicyName:  "hcl/ptolemaios-acl.hcl",
	}
)

func Policy() *cobra.Command {
	var (
		policyName string
		namespace  string
		kubePath   string
	)
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "create policies",
		Long: `Allows you to create policies
- policyName`,
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
			glg.Debug("creating a vault policy")
			createPolicy(policyName, namespace, "", kubeManager)
		},
	}

	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace defaults to odysseia")
	cmd.PersistentFlags().StringVarP(&kubePath, "kubepath", "k", "", "kubeconfig filepath defaults to ~/.kube/config")
	cmd.PersistentFlags().StringVarP(&policyName, "policy", "p", "", "policy to add to kubernetes auth default to solon-admin")

	return cmd
}

func createPolicy(policyName, namespace, rootToken string, kube kubernetes.KubeClient) {
	for key, value := range policies {
		var policyToCreate []byte

		if policyName == "" {
			glg.Debug("no policy set")
			policyName = key
			policy, err := res.ReadFile(value)
			if err != nil {
				glg.Error(err)
			}
			policyToCreate = policy
		} else if key == policyName {
			policy, err := res.ReadFile(value)
			if err != nil {
				glg.Error(err)
			}
			policyToCreate = policy
		} else {
			continue
		}

		vaultSubString := "vault"
		var vaultSelector string
		var podName string

		sets, err := kube.Workload().GetStatefulSets(namespace)
		if err != nil {
			glg.Error(err)
		}

		for _, set := range sets.Items {
			if strings.Contains(set.Name, "vault") {
				for key, element := range set.Spec.Selector.MatchLabels {
					if element == vaultSubString {
						vaultSelector = fmt.Sprintf("%s=%s", key, element)
					}
				}
			}
		}

		pods, err := kube.Workload().GetPodsBySelector(namespace, vaultSelector)
		if err != nil {
			glg.Error(err)
		}
		for _, pod := range pods.Items {
			if strings.Contains(pod.Name, vaultSubString) {
				if pod.Status.Phase == "Running" {
					glg.Debugf(fmt.Sprintf("%s is running in release %s", pods.Items[0].Name, namespace))
					podName = pod.Name
					break
				}
			}
		}

		glg.Debugf("found vault pod running as %s", podName)

		srcPath := fmt.Sprintf("/tmp/%s.hcl", policyName)
		util.WriteFile(policyToCreate, srcPath)

		copy, err := kube.Util().CopyFileToPod(podName, srcPath, srcPath)
		if err != nil {
			glg.Error(err)
		}

		glg.Debug(copy)

		glg.Info("file copied to pod")

		loginCommand := []string{"vault", "login", rootToken}

		login, err := kube.Workload().ExecNamedPod(namespace, podName, loginCommand)
		if err != nil {
			glg.Error(err)
		}

		glg.Info(login)

		command := []string{"vault", "policy", "write", policyName, srcPath}

		vaultCreatePolicy, err := kube.Workload().ExecNamedPod(namespace, podName, command)
		if err != nil {
			glg.Error(err)
		}

		glg.Debug(vaultCreatePolicy)
	}
}
