package install

import (
	"fmt"
	"github.com/kpango/glg"
	"strings"
	"time"
)

func (a *AppInstaller) installAppsHelmChart() error {
	timer := 360 * time.Second
	err := a.checkStatusOfPod("solon", timer)
	if err != nil {
		return err
	}

	for _, chart := range a.Charts.Apis {
		for _, install := range a.AppsToInstall {
			if strings.Contains(chart, install) {
				splitName := strings.Split(chart, "/")
				chartName := strings.ToLower(splitName[len(splitName)-1])

				helmInstalled, _ := a.checkForHelmRelease(chartName)
				if helmInstalled {
					glg.Debug(fmt.Sprintf("skipping install because already installed: %s", chartName))
					continue
				}

				err := a.installHelmChart(chartName, chart)
				if err != nil {
					if strings.Contains(chartName, "ingress") {
						continue
					}
					return err
				}
			}
		}
	}

	return nil
}

func (a *AppInstaller) installHelmChart(name, chartPath string) error {
	helmInstalled, _ := a.checkForHelmRelease(name)
	if helmInstalled {
		glg.Debugf("skipping install because already installed: %s", name)
		return nil
	}

	rls, err := a.Helm.Install(chartPath)
	if err != nil {
		return err
	}
	glg.Info(rls.Name)

	return nil
}

func (a *AppInstaller) installHelmChartWithValues(name, chartPath string, valuesOverwrite map[string]interface{}) error {
	helmInstalled, _ := a.checkForHelmRelease(name)
	if helmInstalled {
		glg.Debugf("skipping install because already installed: %s", name)
		return nil
	}

	rls, err := a.Helm.InstallWithValues(chartPath, valuesOverwrite)
	if err != nil {
		return err
	}
	glg.Info(rls.Name)

	return nil
}

func (a *AppInstaller) checkForHelmRelease(name string) (bool, error) {
	charts, err := a.Helm.List()
	if err != nil {
		return false, err
	}

	for _, chart := range charts {
		if chart.Name == name {
			return true, nil
		}
	}

	return false, nil
}
