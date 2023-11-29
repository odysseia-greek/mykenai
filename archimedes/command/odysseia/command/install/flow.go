package install

import (
	"fmt"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"strings"
)

func (a *AppInstaller) InstallOdysseiaComplete() error {
	err := a.preSteps()
	if err != nil {
		glg.Error(err)
		return err
	}

	err = a.fillHelmChartPaths()
	if err != nil {
		glg.Error(err)
		return err
	}

	err = a.installPrerequisites()
	if err != nil {
		return err
	}

	defer func() {
		//save config to configpath
		err = a.checkConfigForEmpty()
		if err != nil {
			glg.Error(err)
		}

		currentConfig, err := yaml.Marshal(a.Config)
		if err != nil {
			glg.Error(err)
		}

		glg.Info(string(currentConfig))
		currentConfigPath := filepath.Join(a.ConfigPath, "config.yaml")
		util.WriteFile(currentConfig, currentConfigPath)
		// copy everything to currentdir
		err = a.copyToCurrentDir()
		if err != nil {
			glg.Error(err)
		}

		return
	}()

	//1. loop over install candidates
	installElastic := false
	installPerikles := false
	installVault := false
	installSolon := false
	installIngress := false
	installHomeros := false
	installEupalinos := false
	installEuripides := false

	for _, install := range a.AppsToInstall {
		switch install {
		case elasticsearch:
			installElastic = true
		case elastic:
			installElastic = true
		case "perikles":
			installPerikles = true
		case "vault":
			installVault = true
		case "solon":
			installSolon = true
		case "ingress":
			installIngress = true
		case "homeros":
			installHomeros = true
		case "eupalinos":
			installEupalinos = true
		case "euripides":
			installEuripides = true
		}
	}

	if installElastic {
		a.addElasticOperator()
		err = a.installElasticOperator()
		if err != nil {
			return err
		}
		err := a.setElasticSettings()
		if err != nil {
			return err
		}

	}

	//2. install perikles
	if installPerikles {
		err = a.installPerikles()
		if err != nil {
			return err
		}
	}

	if installVault {
		//4. install vault
		_, err := a.installVaultHelmChart()
		if err != nil {
			return err
		}
	}

	if installSolon {
		//6. install solon
		err = a.waitForPerikles()
		if err != nil {
			return err
		}

		splitName := strings.Split(a.Charts.Solon, "/")
		chartName := strings.ToLower(splitName[len(splitName)-1])
		values := a.ValueConfig["infra"].(map[string]interface{})
		if a.VaultUnsealMethod != "" {
			if values["envVariables"] == nil {
				values["envVariables"] = make(map[string]interface{})
			}

			envVariables, ok := values["envVariables"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("peisistratos config could not be created")
			}

			if envVariables["peisistratos"] == nil {
				envVariables["peisistratos"] = make(map[string]interface{})
			}

			peisistratos, ok := envVariables["peisistratos"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("peisistratos config could not be created")
			}

			peisistratos["unsealProvider"] = a.VaultUnsealMethod
		}
		err = a.installHelmChartWithValues(chartName, a.Charts.Solon, values)
		if err != nil {
			return err
		}

		err = a.waitForSolon()
		if err != nil {
			return err
		}
	}

	if installEuripides {
		//7. install euripides
		err = a.waitForSolon()

		splitName := strings.Split(a.Charts.Euripides, "/")
		chartName := strings.ToLower(splitName[len(splitName)-1])
		values := a.ValueConfig["apis"].(map[string]interface{})
		err = a.installHelmChartWithValues(chartName, a.Charts.Euripides, values)
		if err != nil {
			return err
		}
	}

	if installEupalinos {
		err = a.waitForPerikles()
		if err != nil {
			return err
		}

		err = a.waitForSolon()
		if err != nil {
			return err
		}

		//8. install eupalinos
		err = a.InstallEupalinos()
		if err != nil {
			return err
		}
	}

	if installHomeros {
		//9. install homeros
		splitName := strings.Split(a.Charts.Homeros, "/")
		chartName := strings.ToLower(splitName[len(splitName)-1])
		values := a.ValueConfig["apis"].(map[string]interface{})
		err = a.installHelmChartWithValues(chartName, a.Charts.Homeros, values)
		if err != nil {
			return err
		}
	}

	//7. install app
	err = a.waitForEupalinos()
	if err != nil {
		return err
	}
	err = a.installAppsHelmChart()
	if err != nil {
		return err
	}

	installSystemTests := false
	for _, install := range a.AppsToInstall {
		if install == "hippokrates" {
			installSystemTests = true
			break
		}
	}

	if installIngress {
		//6. install ingress
		splitName := strings.Split(a.Charts.Thermopulai, "/")
		chartName := strings.ToLower(splitName[len(splitName)-1])
		values := a.ValueConfig["ingress"].(map[string]interface{})
		err = a.installHelmChartWithValues(chartName, a.Charts.Thermopulai, values)
		if err != nil {
			return err
		}
	}

	//8. install tests
	if installSystemTests {
		splitName := strings.Split(a.Charts.Hippokrates, "/")
		chartName := strings.ToLower(splitName[len(splitName)-1])
		err = a.installHelmChart(chartName, a.Charts.Hippokrates)
		if err != nil {
			return err
		}
	}

	installDocs := false
	for _, install := range a.AppsToInstall {
		if install == "ploutarchos" {
			installDocs = true
			break
		}
	}

	//9. inst
	//all docs
	if installDocs {
		splitName := strings.Split(a.Charts.Ploutarchos, "/")
		chartName := strings.ToLower(splitName[len(splitName)-1])
		err = a.installHelmChart(chartName, a.Charts.Ploutarchos)
		if err != nil {
			return err
		}
	}

	return nil
}
