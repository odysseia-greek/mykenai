package install

import (
	"encoding/json"
	"fmt"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"github.com/odysseia-greek/plato/certificates"
	"github.com/odysseia-greek/plato/generator"
	"github.com/odysseia-greek/plato/harbor"
	"time"
)

func (a *AppInstaller) installHarborHelmChart() error {
	//create harbor login
	password, err := generator.RandomPassword(24)
	if err != nil {
		return err
	}

	a.Config.HarborPassword = password

	glg.Debug(password)

	data := make(map[string]string)
	data["docker-server"] = "harbor"
	data["docker-username"] = command.DefaultAdmin
	data["docker-password"] = password
	data["docker-email"] = "odysseia@example.com"

	secret, _ := json.Marshal(data)
	secretData := map[string]string{".dockerconfigjson": string(secret)}

	err = a.Kube.Configuration().CreateDockerSecret(a.Namespace, command.DefaultDockerRegistrySecret, secretData)
	if err != nil {
		return err
	}

	hosts := []string{
		command.DefaultHarbor,
	}
	org := []string{
		command.DefaultNamespace,
	}

	validity := 3650

	certClient, err := certificates.NewCertGeneratorClient(org, validity)
	err = certClient.InitCa()
	if err != nil {
		return err
	}

	crt, key, _ := certClient.GenerateKeyAndCertSet(hosts, validity)
	certData := make(map[string][]byte)
	certData["tls.key"] = key
	certData["tls.crt"] = crt

	secretName := command.DefaultHarborCertSecretName

	err = a.Kube.Configuration().CreateSecret(command.DefaultNamespace, secretName, certData)
	if err != nil {
		return err
	}

	a.ValueConfig.Harbor.HarborAdminPassword = password
	a.ValueConfig.Harbor.Expose.TLS.Secret.SecretName = secretName

	values, err := a.parseValueOverwrite("harbor")
	if err != nil {
		return err
	}

	rls, err := a.Helm.InstallWithValues(a.Charts.Harbor, values)
	if err != nil {
		return err
	}
	glg.Info(rls.Name)

	harborManager, _ := harbor.NewHarborClient(command.DefaultHarborUrl, command.DefaultAdmin, password, crt)
	a.Harbor = harborManager

	return nil
}

func (a *AppInstaller) setupHarbor() error {
	//wait for harbor to install
	timer := 120 * time.Second
	err := a.checkStatusOfPod("harbor-core", timer)
	if err != nil {
		return err
	}

	err = a.Harbor.CreateProject(command.DefaultNamespace, true)
	return err
}

func (a *AppInstaller) dockerLogin() error {
	dockerCommand := fmt.Sprintf("docker login %s --username %s --password %s", a.ValueConfig.Harbor.ExternalURL, command.DefaultAdmin, a.ValueConfig.Harbor.HarborAdminPassword)
	err := util.ExecCommand(dockerCommand, "/")
	if err != nil {
		return err
	}
	return nil
}
