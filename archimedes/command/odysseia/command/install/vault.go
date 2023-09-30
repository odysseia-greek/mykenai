package install

import (
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	vaultCommand "github.com/odysseia-greek/mykenai/archimedes/command/vault/command"
	"strings"
	"time"
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

	rls, err := a.Helm.Install(a.Charts.Vault)
	if err != nil {
		return false, err
	}

	glg.Info(rls.Name)

	return true, nil
}

func (a *AppInstaller) setupVault(installed bool) error {
	ready := false
	helmInstalled, _ := a.checkForHelmRelease(vaultName)
	if helmInstalled {
		pods, err := a.Kube.Workload().List(command.DefaultNamespace)
		if err != nil {
			return err
		}

		for _, pod := range pods.Items {
			if strings.Contains(pod.Name, "vault") {

				timer := 60 * time.Second
				err := a.podIsRunning(pod.Name, timer)
				if err != nil {
					glg.Debugf("skipping install because vault is not running after %v seconds", timer)
					return err
				}

				po, err := a.Kube.Workload().GetPodByName(command.DefaultNamespace, pod.Name)
				if err != nil {
					glg.Debug("error getting pod")
					return err
				}

				for _, condition := range po.Status.Conditions {
					if condition.Type == "Ready" && condition.Status == "True" {
						glg.Infof("pod: %s ready: %s", po.Name, condition.Type)
						ready = true
					}
				}
			}
		}
	}

	if ready {
		glg.Debug("vault running and initiated")
		return nil
	}

	if !ready && installed {
		glg.Debug("vault running but not initiated")
		time.Sleep(2000 * time.Millisecond)
		vaultConfig, err := vaultCommand.NewVaultFlow(a.Namespace, a.Kube)
		a.Config.VaultUnsealKey = vaultConfig.UnsealKeysHex[0]
		a.Config.VaultRootToken = vaultConfig.RootToken
		return err
	} else if !ready && !installed {
		glg.Info("vault is not ready and not installed so it might be an older install checking status")
		status, err := vaultCommand.VaultStatus(a.Namespace, a.Kube)
		if err != nil && status == nil {
			return err
		}

		log, _ := status.Marshal()
		glg.Debug(string(log))

		if !status.Initialized {
			time.Sleep(2000 * time.Millisecond)
			vaultConfig, err := vaultCommand.NewVaultFlow(a.Namespace, a.Kube)
			a.Config.VaultUnsealKey = vaultConfig.UnsealKeysHex[0]
			a.Config.VaultRootToken = vaultConfig.RootToken
			return err
		}

		if status.Initialized && status.Sealed {
			glg.Debug("should be deleted and reinstalled")
		}

	}

	return nil
}
