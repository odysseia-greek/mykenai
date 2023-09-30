package install

import (
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"strings"
)

func (a *AppInstaller) InstallOdysseiaComplete(legacyElastic bool) error {
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
	installHarbor := false
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
		case "harbor":
			installHarbor = true
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
		if legacyElastic {
			err = a.createElastic()
			if err != nil {
				return err
			}
		} else {
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
	}

	//2. install perikles
	if installPerikles {
		err = a.installPerikles()
		if err != nil {
			return err
		}
	}

	if installHarbor {
		//3. install harbor
		err = a.installHarborHelmChart()
		if err != nil {
			return err
		}

		//3.a create harbor project etc.
		err = a.setupHarbor()
		if err != nil {
			return err
		}

		glg.Infof("created harbor project %s at %s", command.DefaultNamespace, command.DefaultHarborUrl)

		//3.b. docker login
		err = a.dockerLogin()
		if err != nil {
			return err
		}

	}

	if installVault {
		//4. install vault
		intstalled, err := a.installVaultHelmChart()
		if err != nil {
			return err
		}

		//4b. provision vault
		err = a.setupVault(intstalled)
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
		err = a.installHelmChart(chartName, a.Charts.Solon)
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
		splitName := strings.Split(a.Charts.Euripides, "/")
		chartName := strings.ToLower(splitName[len(splitName)-1])
		err = a.installHelmChart(chartName, a.Charts.Euripides)
		if err != nil {
			return err
		}
	}

	if installEupalinos {
		err = a.waitForPerikles()
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
		err = a.installHelmChart(chartName, a.Charts.Homeros)
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
		err = a.installHelmChart(chartName, a.Charts.Thermopulai)
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
