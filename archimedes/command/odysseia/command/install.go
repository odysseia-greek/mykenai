package command

import (
	"context"
	"fmt"
	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/odysseia-greek/agora/thales"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Install() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "install",
		Short: "create everything odysseia",
		Long: `Allows you to create documentation for all apis
`,
		Run: func(cmd *cobra.Command, args []string) {
			kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
			data, _ := os.ReadFile(kubeconfigPath)
			kube, err := thales.NewFromConfig(data)
			if err != nil {
				logging.Error(errors.Wrap(err, "Failed to create new Kube client").Error())
				return
			}
			logging.Debug(kube.Host())
		},
	}

	return cmd
}

func install(client thales.KubeClient, ns string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	nsToCreate := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
		},
	}
	_, err := client.CoreV1().Namespaces().Create(ctx, nsToCreate, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func createVaultAutounsealConfig(ns, vaultSaPath string, client thales.KubeClient) error {
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

	logging.Debug(fmt.Sprintf("created new secret: %s", creation.Name))
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
