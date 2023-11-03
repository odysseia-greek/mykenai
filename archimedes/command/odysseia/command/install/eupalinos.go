package install

import (
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	"strings"
	"time"
)

func (a *AppInstaller) waitForEupalinos() error {
	pods, err := a.Kube.Workload().List(command.DefaultNamespace)
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, "eupalinos") {
			timer := 120 * time.Second
			err := a.checkStatusOfNamedPod(pod.Name, timer)
			if err != nil {
				return err
			}

			glg.Debug("rabbitmq running and initiated")
			break
		}
	}

	return nil
}

func (a *AppInstaller) InstallEupalinos() error {
	helmInstalled, _ := a.checkForHelmRelease("eupalinos")
	if helmInstalled {
		glg.Debug("skipping install because already installed")
		return nil
	}

	values := a.ValueConfig["infra"].(map[string]interface{})

	rls, err := a.Helm.InstallWithValues(a.Charts.Eupalinos, values)
	if err != nil {
		return err
	}
	glg.Info(rls.Name)

	return nil
}
