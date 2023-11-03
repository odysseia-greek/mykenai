package install

import (
	"github.com/kpango/glg"
	vaultCommand "github.com/odysseia-greek/mykenai/archimedes/command/vault/command"
)

const vaultName string = "vault"

func (a *AppInstaller) installVaultHelmChart() (bool, error) {
	helmInstalled, _ := a.checkForHelmRelease(vaultName)
	if helmInstalled {
		glg.Debug("skipping install because already installed")
		return false, nil
	}

	secrets, err := a.Kube.Configuration().ListSecrets(a.Namespace)
	if err != nil {
		return false, nil
	}

	vaultSecrets := []string{
		"vault-server-tls",
	}

	for _, secret := range secrets.Items {
		for _, vaultSecret := range vaultSecrets {
			if secret.Name == vaultSecret {
				err = a.Kube.Configuration().DeleteSecret(a.Namespace, secret.Name)
				if err != nil {
					continue
				}
			}
		}
	}

	vaultPvcs := []string{
		"data-vault-0",
	}

	pvcs, err := a.Kube.Storage().ListPvc(a.Namespace)
	if err != nil {
		return false, nil
	}

	for _, pvc := range pvcs.Items {
		for _, vaultPvc := range vaultPvcs {
			if pvc.Name == vaultPvc {
				err = a.Kube.Storage().DeletePvc(a.Namespace, pvc.Name)
				if err != nil {
					continue
				}
			}
		}
	}

	vaultCsr := []string{
		"vault-csr",
	}

	csr, err := a.Kube.Certificate().ListCsr()
	if err != nil {
		return false, nil
	}

	for _, request := range csr.Items {
		for _, v := range vaultCsr {
			if request.Name == v {
				err = a.Kube.Certificate().DeleteCsr(request.Name)
				if err != nil {
					continue
				}
			}
		}
	}

	vaultCommand.EnableTlS(a.Namespace, vaultName, a.Kube)

	values := a.ValueConfig["vault"].(map[string]interface{})

	rls, err := a.Helm.InstallWithValues(a.Charts.Vault, values)

	if err != nil {
		return false, err
	}

	glg.Info(rls.Name)

	return true, nil
}
