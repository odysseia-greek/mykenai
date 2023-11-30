package install

import (
	"encoding/base64"
	"fmt"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/agora/plato/certificates"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	"strings"
	"time"
)

func (a *AppInstaller) installPerikles() error {
	helmInstalled, _ := a.checkForHelmRelease("perikles")
	if helmInstalled {
		glg.Debug("skipping install because already installed")
		return nil
	}

	cert, err := a.setupPerikles()

	values := a.ValueConfig["infra"].(map[string]interface{})
	config := values["config"].(map[string]interface{})
	config["caBundle"] = cert

	rls, err := a.Helm.InstallWithValues(a.Charts.Perikles, values)
	if err != nil {
		return err
	}
	glg.Info(rls.Name)

	//wait for perikles to be healthy
	checkFor := 180 * time.Second
	err = a.checkStatusOfPod(a.Charts.Perikles, checkFor)
	if err != nil {
		return err
	}

	return nil
}

func (a *AppInstaller) waitForPerikles() error {
	pods, err := a.Kube.Workload().List(command.DefaultNamespace)
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, "perikles") {
			timer := 120 * time.Second
			err := a.checkStatusOfNamedPod(pod.Name, timer)
			if err != nil {
				return err
			}

			glg.Debug("perikles running and initiated")
			break
		}
	}

	return nil
}

func (a *AppInstaller) waitForSolon() error {
	pods, err := a.Kube.Workload().List(command.DefaultNamespace)
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, "solon") {
			timer := 360 * time.Second
			err := a.checkStatusOfNamedPod(pod.Name, timer)
			if err != nil {
				return err
			}

			glg.Debug("solon running and initiated")
			break
		}
	}

	return nil
}

func (a *AppInstaller) setupPerikles() (string, error) {
	//create cert pair
	validity := 3650

	orgName := []string{
		command.DefaultNamespace,
	}

	hosts := []string{
		fmt.Sprintf("%s", command.DefaultPerikles),
		fmt.Sprintf("%s.%s", command.DefaultPerikles, command.DefaultNamespace),
		fmt.Sprintf("%s.%s.svc", command.DefaultPerikles, command.DefaultNamespace),
		fmt.Sprintf("%s.%s.svc.cluster.local", command.DefaultPerikles, command.DefaultNamespace),
	}

	certClient, err := certificates.NewCertGeneratorClient(orgName, validity)
	err = certClient.InitCa()
	if err != nil {
		return "", err
	}

	crt, key, _ := certClient.GenerateKeyAndCertSet(hosts, validity)
	certData := make(map[string][]byte)
	certData["tls.key"] = key
	certData["tls.crt"] = crt

	secretName := command.DefaultPeriklesSecretName

	_, err = a.Kube.Configuration().GetSecret(command.DefaultNamespace, command.DefaultPeriklesSecretName)
	if err == nil {
		err = a.Kube.Configuration().DeleteSecret(command.DefaultNamespace, command.DefaultPeriklesSecretName)
		if err != nil {
			return "", err
		}
	}

	//create secret
	err = a.Kube.Configuration().CreateSecret(command.DefaultNamespace, secretName, certData)
	if err != nil {
		return "", err
	}

	//return secret for valueoverwrite
	encodedCert := base64.StdEncoding.EncodeToString(crt)

	return encodedCert, nil
}
