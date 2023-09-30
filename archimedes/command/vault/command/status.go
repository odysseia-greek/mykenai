package command

import (
	"fmt"
	"github.com/kpango/glg"
	kubernetes "github.com/odysseia-greek/thales"
	"strings"
)

func VaultStatus(namespace string, kube kubernetes.KubeClient) (*Status, error) {
	vaultSelector := "app.kubernetes.io/name=vault"

	pods, err := kube.Workload().GetPodsBySelector(namespace, vaultSelector)
	if err != nil {
		glg.Error(err)
		return nil, err
	}

	var podName string

	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, "vault") {
			if pod.Status.Phase == "Running" {
				glg.Debugf(fmt.Sprintf("%s is running in release %s", pods.Items[0].Name, namespace))
				if strings.Contains(pod.Name, "-0") {
					podName = pod.Name
				}
			}
		}
	}

	vaultCommand := []string{"vault", "status", "-format=json"}
	vaultStatus, err := kube.Workload().ExecNamedPod(namespace, podName, vaultCommand)
	var status Status
	if vaultStatus != "" {
		status, _ = UnmarshalStatus([]byte(vaultStatus))
	}

	return &status, err
}

func determineMode(namespace string, kube kubernetes.KubeClient) *vaultConfig {
	var config vaultConfig
	vaultSelector := "app.kubernetes.io/name=vault"
	pods, _ := kube.Workload().GetPodsBySelector(namespace, vaultSelector)
	config.HaEnabled = len(pods.Items) > 1

	for _, po := range pods.Items {
		if strings.Contains(po.Name, "-0") {
			config.PrimaryNode = po.Name
		} else {
			config.SecondaryNodes = append(config.SecondaryNodes, po.Name)
		}
	}

	return &config
}
