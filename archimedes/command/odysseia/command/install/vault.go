package install

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"os"
	"path/filepath"
	"strings"
)

const vaultName string = "vault"

type ServiceAccountInfo struct {
	ProjectID string `json:"project_id"`
}

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

	a.enableTlS(vaultName)

	if a.VaultSaPath != "" {
		data, err := os.ReadFile(a.VaultSaPath)
		if err != nil {
			return false, err
		}

		secretName, dataKey := a.determineSecretAttributes()
		if secretName == "" || dataKey == "" {
			return false, fmt.Errorf("failed to determine secret name or data key")
		}

		// Check if the secret already exists; if it does, delete it
		err = a.Kube.Configuration().DeleteSecret(a.Namespace, secretName)
		if err != nil {
			if !strings.Contains(err.Error(), "not found") {
				return false, err
			}
		}

		secret := map[string][]byte{
			dataKey: data,
		}

		err = a.Kube.Configuration().CreateSecret(a.Namespace, secretName, secret)
		if err != nil {
			return false, err
		}

		var saInfo ServiceAccountInfo
		err = json.Unmarshal(data, &saInfo)
		if err != nil {
			return false, err
		}

		a.updateHelmValues(secretName, dataKey, saInfo.ProjectID)
	}

	values := a.ValueConfig["vault"].(map[string]interface{})

	rls, err := a.Helm.InstallWithValues(a.Charts.Vault, values)

	if err != nil {
		return false, err
	}

	glg.Info(rls.Name)

	return true, nil
}

func (a *AppInstaller) determineSecretAttributes() (string, string) {
	fileName := filepath.Base(a.VaultSaPath)
	secretNames := []string{"gcp", "azure", "aws"}

	for _, name := range secretNames {
		if strings.Contains(fileName, name) {
			a.VaultUnsealMethod = name
			return fmt.Sprintf("vaultunseal%s", name), fmt.Sprintf("%sconfig.json", name)
		}
	}

	return "", ""
}

func (a *AppInstaller) updateHelmValues(secretName, dataKey, project string) {
	values := a.ValueConfig["vault"].(map[string]interface{})
	volumeMountName := fmt.Sprintf("userconfig-%s", secretName)

	serverConfig := map[string]interface{}{
		"extraEnvironmentVars": map[string]interface{}{
			"VAULT_CACERT":                   "/vault/userconfig/vault-server-tls/vault.ca",
			"GOOGLE_APPLICATION_CREDENTIALS": fmt.Sprintf("/vault/userconfig/unseal/%s", dataKey),
			"GOOGLE_REGION":                  "global",
			"GOOGLE_PROJECT":                 project,
		},
		"volumes": []map[string]interface{}{
			{
				"name": "userconfig-vault-server-tls",
				"secret": map[string]interface{}{
					"defaultMode": int64(420),
					"secretName":  "vault-server-tls",
				},
			},
			{
				"name": volumeMountName,
				"secret": map[string]interface{}{
					"secretName": secretName,
				},
			},
		},
		"volumeMounts": []map[string]interface{}{
			{
				"mountPath": "/vault/userconfig/vault-server-tls",
				"name":      "userconfig-vault-server-tls",
				"readOnly":  true,
			},
			{
				"mountPath": "/vault/userconfig/unseal",
				"name":      volumeMountName,
			},
		},
	}

	mergeMaps(values["server"].(map[string]interface{}), serverConfig)
	var storageType string
	haMode, exists := values["server"].(map[string]interface{})["ha"].(map[string]interface{})["enabled"].(bool)
	if !exists {
		storageType = "file"
	} else if haMode {
		storageType = "raft"
	} else {
		storageType = "file"
	}

	config := fmt.Sprintf(`
            ui = true
            listener "tcp" {
                address = "[::]:8200"
                cluster_address = "[::]:8201"
                tls_cert_file = "/vault/userconfig/vault-server-tls/vault.crt"
                tls_key_file  = "/vault/userconfig/vault-server-tls/vault.key"
                tls_client_ca_file = "/vault/userconfig/vault-server-tls/vault.crt"
            }

            seal "gcpckms" {
                project     = "%s"
                region      = "global"
                key_ring    = "%s"
                crypto_key  = "%s"
            }

            storage "%s" {
                path = "/vault/data"
            }
        `, project, a.KeyRing, a.CryptoKey, storageType)

	if haMode {
		raftSection, raftExists := values["server"].(map[string]interface{})["ha"].(map[string]interface{})["raft"]
		if !raftExists || raftSection == nil {
			raftSection = make(map[string]interface{})
			values["server"].(map[string]interface{})["ha"].(map[string]interface{})["raft"] = raftSection
		}

		values["server"].(map[string]interface{})["ha"].(map[string]interface{})["raft"].(map[string]interface{})["config"] = config
	} else {
		standaloneSection, standaloneExists := values["server"].(map[string]interface{})["standalone"]
		if !standaloneExists || standaloneSection == nil {
			standaloneSection = make(map[string]interface{})
			values["server"].(map[string]interface{})["standalone"] = standaloneSection
		}

		values["server"].(map[string]interface{})["standalone"].(map[string]interface{})["config"] = config
	}

	a.ValueConfig["vault"] = values
}

func mergeMaps(dest, src map[string]interface{}) {
	for k, v := range src {
		if val, ok := dest[k]; ok {
			destValueMap, destIsMap := val.(map[string]interface{})
			srcValueMap, srcIsMap := v.(map[string]interface{})

			if destIsMap && srcIsMap {
				// If both values are maps, recursively merge them
				mergeMaps(destValueMap, srcValueMap)
			} else {
				// Otherwise, overwrite the destination value with the source value
				dest[k] = v
			}
		} else {
			// If the key doesn't exist in the destination map, add it
			dest[k] = v
		}
	}
}

func (a *AppInstaller) enableTlS(service string) {
	glg.Debug("setting up TLS for vault")

	secretName := "vault-server-tls"
	tmpDir := "/tmp"
	csrName := "vault-csr"

	commandKey := fmt.Sprintf("openssl genrsa -out %s/vault.key 2048", tmpDir)
	err := util.ExecCommand(commandKey, tmpDir)
	if err != nil {
		glg.Error(err)
	}

	keyFromFile, err := os.ReadFile(fmt.Sprintf("%s/vault.key", tmpDir))
	if err != nil {
		glg.Error(err)
	}

	altNames := fmt.Sprintf("DNS.1 = %s", service)
	altNames += fmt.Sprintf("\nDNS.2 = %s.%s", service, a.Namespace)
	altNames += fmt.Sprintf("\nDNS.3 = %s.%s.svc", service, a.Namespace)
	altNames += fmt.Sprintf("\nDNS.4 = %s.%s.svc.cluster.local", service, a.Namespace)
	//vault-0.vault-internal
	altNames += fmt.Sprintf("\nDNS.5 = %s-0.vault-internal", service)
	altNames += fmt.Sprintf("\nDNS.6 = %s-1.vault-internal", service)
	altNames += fmt.Sprintf("\nDNS.7 = %s-2.vault-internal", service)

	csrConf := fmt.Sprintf(`[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
%s
IP.1 = 127.0.0.1
`, altNames)
	outputFileCsr := fmt.Sprintf("%s/csr.conf", tmpDir)

	util.WriteFile([]byte(csrConf), outputFileCsr)

	commandCsr := fmt.Sprintf(`openssl req -new -key %s/vault.key \
    -subj "/O=system:nodes/CN=system:node:%s.%s.svc" \
    -out %s/server.csr \
    -config %s/csr.conf`, tmpDir, service, a.Namespace, tmpDir, tmpDir)
	err = util.ExecCommand(commandCsr, tmpDir)
	if err != nil {
		glg.Error(err)
	}

	serverCsr, err := os.ReadFile(fmt.Sprintf("%s/server.csr", tmpDir))
	if err != nil {
		glg.Error(err)
	}

	encodedCsr := base64.StdEncoding.EncodeToString(serverCsr)

	csrYaml := fmt.Sprintf(`apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: %s
spec:
  groups:
  - system:authenticated
  request: %s
  signerName: kubernetes.io/kubelet-serving
  usages:
  - digital signature
  - key encipherment
  - server auth
`, csrName, encodedCsr)

	outputFileYaml := fmt.Sprintf("%s/csr.yaml", tmpDir)

	util.WriteFile([]byte(csrYaml), outputFileYaml)

	kubeCommand := fmt.Sprintf("kubectl create -f %s/csr.yaml", tmpDir)
	err = util.ExecCommand(kubeCommand, tmpDir)
	if err != nil {
		glg.Error(err)
	}

	kubeCommand = fmt.Sprintf("kubectl certificate approve %s", csrName)
	err = util.ExecCommand(kubeCommand, tmpDir)
	if err != nil {
		glg.Error(err)
	}

	ca, err := a.Kube.Cluster().GetHostCaCert()
	if err != nil {
		glg.Error(err)
	}

	crtCommand := fmt.Sprintf("kubectl get csr %s -o jsonpath='{.status.certificate}'", csrName)
	cert, err := util.ExecCommandWithReturn(crtCommand, tmpDir)
	if err != nil {
		glg.Error(err)
	}

	decodedCert, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		glg.Error(err)
	}

	data := make(map[string][]byte)
	data["vault.crt"] = decodedCert
	data["vault.key"] = keyFromFile

	if strings.Contains(string(ca), "-----BEGIN CERTIFICATE-----") {
		data["vault.ca"] = ca
	} else {
		decodedCa, err := base64.StdEncoding.DecodeString(string(ca))
		if err != nil {
			glg.Error(err)
		}
		data["vault.ca"] = decodedCa
	}

	err = a.Kube.Configuration().CreateSecret(a.Namespace, secretName, data)
	if err != nil {
		glg.Error(err)
	}

	glg.Debug("finished setting up TLS for vault")
}
