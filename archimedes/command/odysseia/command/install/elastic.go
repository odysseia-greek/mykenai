package install

import (
	"fmt"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"path/filepath"
	"time"
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

func (a *AppInstaller) setElasticSettings() error {
	var data []byte
	ready := false
	for !ready {
		secretName := fmt.Sprintf("%s-%s", a.ElasticConfig.Name, elasticNameAppend)
		secret, err := a.Kube.Configuration().GetSecret(a.Namespace, secretName)
		if err != nil {
			glg.Error(err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		data = secret.Data["elastic"]
		ready = true
	}

	a.Config.ElasticPassword = string(data)

	var crt []byte
	certReady := false
	for !certReady {
		certName := fmt.Sprintf("%s-%s", a.ElasticConfig.Name, certNameAppend)
		certSecret, err := a.Kube.Configuration().GetSecret(a.Namespace, certName)
		if err != nil {
			glg.Error(err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		crt = certSecret.Data["tls.crt"]
		certReady = true
	}

	crtDst := filepath.Join(a.ConfigPath, "elastic-certificate.pem")

	util.WriteFile(crt, crtDst)

	return nil
}
