package install

import (
	"fmt"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	es "github.com/odysseia-greek/mykenai/archimedes/command/kubernetes/command"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"github.com/odysseia-greek/plato/generator"
	"os"
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

	err := a.Helm.UpdateRepos()
	if err != nil {
		glg.Error(err)
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

	config := map[string]interface{}{
		"environment": "k3s",
	}

	values := map[string]interface{}{
		"config": config,
	}

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

// createElastic is the legacy method to create a cluster using the deprecated helm chart
func (a *AppInstaller) createElastic() error {
	helmInstalled, _ := a.checkForHelmRelease("elasticsearch")
	if helmInstalled {
		glg.Debug("skipping install because already installed")
		return nil
	}

	secrets, err := a.Kube.Configuration().ListSecrets(a.Namespace)
	if err != nil {
		return nil
	}

	elasticSecrets := []string{
		"elastic-credentials",
		"elastic-certificates",
		"elastic-certificate-pem",
		"elastic-certificate-crt",
		"elastic-root-secret",
	}

	for _, secret := range secrets.Items {
		for _, elasticSecret := range elasticSecrets {
			if secret.Name == elasticSecret {
				err = a.Kube.Configuration().DeleteSecret(a.Namespace, secret.Name)
				if err != nil {
					continue
				}
			}
		}
	}

	elasticPvcs := []string{
		"elastic-master-elastic-master-0",
		"elastic-master-elastic-master-1",
		"elastic-master-elastic-master-2",
	}

	pvcs, err := a.Kube.Storage().ListPvc(a.Namespace)
	if err != nil {
		return nil
	}

	for _, pvc := range pvcs.Items {
		for _, elasticPvc := range elasticPvcs {
			if pvc.Name == elasticPvc {
				err = a.Kube.Storage().DeletePvc(a.Namespace, pvc.Name)
				if err != nil {
					continue
				}
			}
		}
	}

	p12Path, err := es.CreateElasticP12(a.Kube, a.Namespace, a.ConfigPath)
	if err != nil {
		return err
	}

	p12File, err := os.ReadFile(p12Path)
	if err != nil {
		return err
	}

	pemDst := filepath.Join(a.ConfigPath, "elastic-certificate.pem")
	cmd := fmt.Sprintf(`openssl pkcs12 -nodes -passin pass:'' -in %s -out %s`, p12Path, pemDst)

	err = util.ExecCommand(cmd, "/")
	if err != nil {
		return err
	}

	pemFile, err := os.ReadFile(pemDst)
	if err != nil {
		return err
	}
	crtDst := filepath.Join(a.ConfigPath, "elastic-certificate.crt")
	crtFile, err := es.GenerateCrtFromPem(pemFile)
	if err != nil {
		return err
	}

	util.WriteFile(crtFile, crtDst)

	//create secrets
	glg.Info("certs for ES tls mode generated applying them as secrets")

	secretNameP12 := "elastic-certificates"
	dataP12 := make(map[string][]byte)
	dataP12["elastic-certificates.p12"] = p12File

	err = a.Kube.Configuration().CreateSecret(a.Namespace, secretNameP12, dataP12)
	if err != nil {
		return err
	}

	secretNamePem := "elastic-certificate-pem"
	dataPem := make(map[string][]byte)
	dataPem["elastic-certificate.pem"] = pemFile

	err = a.Kube.Configuration().CreateSecret(a.Namespace, secretNamePem, dataPem)
	if err != nil {
		return err
	}

	secretNameCrt := "elastic-certificate-crt"
	dataCrt := make(map[string][]byte)
	dataCrt["elastic-certificate.crt"] = crtFile

	err = a.Kube.Configuration().CreateSecret(a.Namespace, secretNameCrt, dataCrt)
	if err != nil {
		return err
	}

	//create elastic login
	password, err := generator.RandomPassword(24)
	if err != nil {
		return err
	}

	a.Config.ElasticPassword = password

	glg.Debug(password)

	data := make(map[string][]byte)
	data["password"] = []byte(password)
	data["username"] = []byte("elastic")

	err = a.Kube.Configuration().CreateSecret(a.Namespace, command.DefaultSecretName, data)
	if err != nil {
		return err
	}

	glg.Infof("created secret with name %s", command.DefaultSecretName)

	values, err := a.parseValueOverwrite("elastic")
	if err != nil {
		return err
	}

	rls, err := a.Helm.InstallWithValues(a.Charts.ElasticSearch, values)
	if err != nil {
		return err
	}
	glg.Info(rls.Name)

	return nil
}
