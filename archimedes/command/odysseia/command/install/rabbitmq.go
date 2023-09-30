package install

import (
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	"strings"
	"time"
)

func (a *AppInstaller) waitForRabbit() error {
	pods, err := a.Kube.Workload().List(command.DefaultNamespace)
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, "rabbitmq") {
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

func (a *AppInstaller) InstallRabbit() error {
	helmInstalled, _ := a.checkForHelmRelease("rabbitmq")
	if helmInstalled {
		glg.Debug("skipping install because already installed")
		return nil
	}
	rls, err := a.Helm.Install(a.Charts.Rabbitmq)
	if err != nil {
		return err
	}
	glg.Info(rls.Name)

	return nil
}
