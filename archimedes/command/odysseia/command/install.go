package command

import (
	"context"
	"fmt"
	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/odysseia-greek/agora/thales"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ELASTICVERSION  = "2.10.0"
	LONGHORNVERSION = "v1.6.0"
)

var corePodsNames = [...]string{"aristoteles", "perikles", "vault", "solon", "eupalinos"}

func Install() *cobra.Command {
	var (
		namespace              string
		helmFilePath           string
		autoUnsealPath         string
		target                 string
		elasticOperatorVersion string
		longhornVersion        string
		tests                  bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "create everything odysseia",
		Long: `This command lets you create the prereqs for odysseia
`,
		Run: func(cmd *cobra.Command, args []string) {
			argPath := ""
			if len(args) > 0 {
				argPath = args[len(args)-1]
			}

			if argPath == "." {
				currentDir, err := os.Getwd()
				if err != nil {
					return
				}
				argPath = currentDir
			}

			if helmFilePath == "" {
				helmFilePath = argPath
			}

			if helmFilePath == "" {
				currentDir, err := os.Getwd()
				if err != nil {
					return
				}

				logging.Debug(fmt.Sprintf("helmFilePath is empty, defaulting to current dir %s", currentDir))
				helmFilePath = currentDir
			}

			kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
			data, _ := os.ReadFile(kubeconfigPath)
			kube, err := thales.NewFromConfig(data)
			if err != nil {
				logging.Error(errors.Wrap(err, "Failed to create new Kube client").Error())
				return
			}

			err = install(kube, autoUnsealPath, namespace)
			if err != nil {
				logging.Error(errors.Wrap(err, "Failed to install").Error())
				return
			}

			err = createElasticOperator(elasticOperatorVersion)
			if err != nil {
				logging.Error(errors.Wrap(err, "Failed to apply elastic operator").Error())
				return
			}

			if target == "production" {
				if longhornVersion == "" {
					longhornVersion = LONGHORNVERSION
				}

				longHornUrl := fmt.Sprintf("https://raw.githubusercontent.com/longhorn/longhorn/%s/deploy/longhorn.yaml", longhornVersion)
				err := applyManifestFromURL(longHornUrl)
				if err != nil {
					logging.Error(errors.Wrap(err, "Failed to apply longhorn operator").Error())
					return
				}
			}

			err = createApps(helmFilePath, target, namespace, kube, tests)
			if err != nil {
				logging.Error(errors.Wrap(err, "Failed to apply apps").Error())
				return
			}

			logging.System(fmt.Sprintf("Finished creating a fresh install for odysseia for target: %s", target))
		},
	}

	cmd.PersistentFlags().StringVarP(&helmFilePath, "themistokles", "t", "", "Where to find the themistokles and by extension all the helmfiles")
	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "odysseia", "The namespace to use during installs, defaults to 'odysseia' when no value provided.")
	cmd.PersistentFlags().StringVarP(&target, "target", "g", "local", "The target to build for, defaults to 'local' when no value provided.")
	cmd.PersistentFlags().StringVarP(&autoUnsealPath, "unseal", "u", "", "Path to an unseal config to enable vault auto unseal")
	cmd.PersistentFlags().StringVarP(&elasticOperatorVersion, "elasticversion", "e", "", fmt.Sprintf("The elastic version for the operator to use, defaults to '%s' when no value provided.", ELASTICVERSION))
	cmd.PersistentFlags().StringVarP(&longhornVersion, "longhornversion", "l", "", fmt.Sprintf("The longhorn version for the operator to use, defaults to '%s' when no value provided.", LONGHORNVERSION))
	cmd.PersistentFlags().BoolVarP(&tests, "tests", "j", true, "include tests in install, defaults to true")

	return cmd
}

func install(client *thales.KubeClient, autoUnsealPath, ns string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	nsToCreate := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
			Labels: map[string]string{
				"name": ns, // name is set the facilitate the admission webhook ns selector
			},
		},
	}

	_, err := client.CoreV1().Namespaces().Create(ctx, nsToCreate, metav1.CreateOptions{})
	if err != nil {
		if kuberr.IsAlreadyExists(err) {
			logging.Info("Namespace already exists, proceeding.")
		}
		return err
	}

	logging.Info(fmt.Sprintf("Created namespace: %s", ns))

	if autoUnsealPath != "" {
		err := createVaultAutoUnsealConfig(ns, autoUnsealPath, client)
		if err != nil {
			return err
		}
	}

	return nil
}

func createApps(helmfilePath, target, ns string, client *thales.KubeClient, tests bool) error {
	tiers := []string{
		"base",
		"infra",
		"backend",
		"frontend",
	}

	if info, err := os.Stat(helmfilePath); err != nil {
		return err
	} else if !info.IsDir() {
		return errors.New(fmt.Sprintf("helmfilePath is not a directory: %s", helmfilePath))
	}

	helmfileFullPath := filepath.Join(helmfilePath, "helmfile.yaml")
	if _, err := os.Stat(helmfileFullPath); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("helmfile.yaml does not exist in the provided helmfilePath: %s", helmfilePath))
	}

	for _, tier := range tiers {
		err := applyHelmfile(target, tier, helmfilePath)
		if err != nil {
			return err
		}
		if tier == "base" || tier == "infra" {
			err = waitForCorePodsToBeRunning(client, ns)
			if err != nil {
				return err
			}
		}
	}

	if tests {
		err := waitForCorePodsToBeRunning(client, ns)
		if err != nil {
			return err
		}
		err = applyHelmfile(target, "tests", helmfilePath)
		if err != nil {
			return err
		}
	}

	return nil
}

func waitForCorePodsToBeRunning(client *thales.KubeClient, namespace string) error {
	timeout := 5 * time.Minute
	pollInterval := 10 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout reached waiting for all pods to be healthy in namespace %s", namespace)
		case <-time.After(pollInterval):
			pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				return fmt.Errorf("error listing pods: %v", err)
			}

			allHealthy := true
			for _, pod := range pods.Items {
				for _, corePod := range corePodsNames {
					if strings.Contains(pod.Name, corePod) {
						if !isPodHealthy(&pod) {
							allHealthy = false
							logging.Warn(fmt.Sprintf("pod: %s is not healthy", pod.Name))
							break
						}
					}
				}
			}

			if allHealthy {
				logging.Debug("all current core pods are running or have finished")
				return nil
			}
		}
	}
}

func isPodHealthy(pod *corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodRunning && pod.Status.Phase != corev1.PodSucceeded {
		return false
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			if condition.Reason == "PodCompleted" {
				return true
			}
			if condition.Status != corev1.ConditionTrue {
				return false
			}
		}
	}

	return true
}

func createVaultAutoUnsealConfig(ns, vaultSaPath string, client *thales.KubeClient) error {
	data, err := os.ReadFile(vaultSaPath)
	if err != nil {
		return err
	}

	secretName, dataKey := determineSecretAttributes(vaultSaPath)
	if secretName == "" || dataKey == "" {
		return fmt.Errorf("failed to determine secret name or data key")
	}

	secretData := map[string][]byte{
		dataKey: data,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	secretExists := true

	_, err = client.CoreV1().Secrets(ns).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			secretExists = false
		}
	}

	if secretExists {
		err = client.CoreV1().Secrets(ns).Delete(ctx, secretName, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	logging.Info(fmt.Sprintf("secret %s does not exist", secretName))
	immutable := false
	scr := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Immutable: &immutable,
		Data:      secretData,
		Type:      corev1.SecretTypeOpaque,
	}
	creation, err := client.CoreV1().Secrets(ns).Create(ctx, scr, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	logging.Debug(fmt.Sprintf("created new secret: %s for config: %s", creation.Name, dataKey))
	return nil
}

func createElasticOperator(elasticVersion string) error {
	if elasticVersion == "" {
		elasticVersion = ELASTICVERSION
	}
	manifests := []string{
		fmt.Sprintf("https://download.elastic.co/downloads/eck/%s/crds.yaml", elasticVersion),
		fmt.Sprintf("https://download.elastic.co/downloads/eck/%s/operator.yaml", elasticVersion),
	}
	for _, url := range manifests {
		err := applyManifestFromURL(url)
		if err != nil {
			return err
		}
	}

	return nil
}

func determineSecretAttributes(vaultSaPath string) (string, string) {
	fileName := filepath.Base(vaultSaPath)
	secretNames := []string{"gcp", "azure", "aws"}

	for _, name := range secretNames {
		if strings.Contains(fileName, name) {
			return fmt.Sprintf("vaultunseal%s", name), fmt.Sprintf("%sconfig.json", name)
		}
	}

	return "", ""
}

func applyManifestFromURL(url string) error {
	cmd := exec.Command("kubectl", "apply", "-f", url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	logging.Info(fmt.Sprintf("output from apply command: %s", string(output)))
	return nil
}

func applyHelmfile(target, tier, helmfilePath string) error {
	cmd := fmt.Sprintf("helmfile -e %s -l tier=%s apply", target, tier)

	logging.Debug(fmt.Sprintf("creating from helmfile: %s", cmd))
	output, err := util.ExecCommandWithReturn(cmd, helmfilePath)
	if output != "" {
		logging.Debug(output)
	}
	return err
}
