package install

import (
	"fmt"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

func (a *AppInstaller) preSteps() error {
	//create config file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(homeDir, command.ConfigFilePath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// create dir
		glg.Infof("config directory does not exist... creating at: %s", configDir)
		os.Mkdir(configDir, 0755)
	}

	t := time.Now()
	formattedDate := fmt.Sprintf("%d%02d%02d",
		t.Year(), t.Month(), t.Day())
	newFullInstallDir := filepath.Join(configDir, formattedDate)

	a.ConfigPath = newFullInstallDir

	if _, err := os.Stat(newFullInstallDir); os.IsNotExist(err) {
		// create dir
		glg.Infof("creating new config at: %s", newFullInstallDir)
		os.Mkdir(newFullInstallDir, 0755)
	} else {
		glg.Infof("directory: %s already exists", newFullInstallDir)
	}

	currentDir := filepath.Join(configDir, "current")

	a.CurrentPath = currentDir

	if _, err := os.Stat(currentDir); os.IsNotExist(err) {
		// create dir
		glg.Infof("creating current config at: %s", currentDir)
		os.Mkdir(currentDir, 0755)
	} else {
		glg.Infof("directory: %s already exists", currentDir)
	}

	return nil
}

func (a *AppInstaller) fillHelmChartPaths() error {
	files, err := ioutil.ReadDir(a.ThemistoklesRoot)
	if err != nil {
		return err
	}

	for _, f := range files {
		elements := reflect.ValueOf(&a.Charts).Elem()
		found := false
		for i := 0; i < elements.NumField(); i++ {
			fieldName := elements.Type().Field(i).Name
			if strings.ToLower(fieldName) == f.Name() {
				path := filepath.Join(a.ThemistoklesRoot, f.Name())
				elements.FieldByName(fieldName).SetString(path)
				found = true
			}
		}

		if !found {
			path := filepath.Join(a.ThemistoklesRoot, f.Name())
			a.Charts.Apis = append(a.Charts.Apis, path)
		}
	}

	return nil
}

func (a *AppInstaller) installPrerequisites() error {
	err := a.Kube.Namespaces().Create(a.Namespace)
	if err != nil {
		return err
	}

	namespaces, err := a.Kube.Namespaces().List()
	if err != nil {
		return err
	}

	installNginx := true
	for _, ns := range namespaces.Items {
		if ns.Name == command.NginxNamespace {
			installNginx = false
			break
		}
	}

	switch a.Profile {
	case "docker-desktop":
		if installNginx {
			rls, err := a.Helm.InstallNamespaced(command.NginxRepoPath, command.NginxNamespace, true)
			if err != nil {
				return err
			}

			glg.Debugf("created nginx release on docker-desktop %v in ns %v", rls.Name, rls.Namespace)
		}

	case "do":
		if installNginx {
			rls, err := a.Helm.InstallNamespaced(command.NginxRepoPath, command.NginxNamespace, true)
			if err != nil {
				return err
			}

			glg.Debugf("created nginx release on a DO k8s %v in ns %v", rls.Name, rls.Namespace)
		}
	case "minikube":
		tmpDir := "/tmp"
		minikubeCommand := fmt.Sprintf("minikube addons enable ingress")
		err = util.ExecCommand(minikubeCommand, tmpDir)
		if err != nil {
			glg.Error(err)
		}
	case "k3s":
		glg.Info("no ingress installed because cluster is running on k3s")
	default:
		rls, err := a.Helm.InstallNamespaced(command.NginxRepoPath, command.NginxNamespace, true)
		if err != nil {
			return err
		}

		glg.Debugf("created nginx release on an unknown k8s %v in ns %v", rls.Name, rls.Namespace)
	}

	return nil
}
