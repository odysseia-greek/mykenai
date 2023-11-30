package install

import (
	"fmt"
	"github.com/kpango/glg"
)

func (a *AppInstaller) addElasticOperator() {
	helmInstalled, _ := a.checkForHelmRelease(elasticOperator)
	if helmInstalled {
		glg.Debug("skipping install because already installed")
		return
	}

	found := a.Helm.SearchRepo(elastic)
	if !found {
		a.Helm.AddRepo(elastic, elasticUrl)
	}

	rls, err := a.Helm.InstallNamespacedWithRelease(fmt.Sprintf("elastic/%s", elasticOperator), elasticOperator, elasticNameSpace, true)
	if err != nil {
		glg.Error(err)
	}

	glg.Info(rls.Name)

}

func (a *AppInstaller) installElasticOperator() error {
	helmInstalled, _ := a.checkForHelmRelease("elasticsearch")
	if helmInstalled {
		glg.Debug("skipping install because already installed")
		return nil
	}

	values := a.ValueConfig["elastic"].(map[string]interface{})

	rls, err := a.Helm.InstallWithValues(a.Charts.Elastic, values)
	if err != nil {
		return err
	}
	glg.Info(rls.Name)
	//todo wait for crd to be complete

	return nil
}
