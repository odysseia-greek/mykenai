package install

import (
	"github.com/odysseia-greek/mykenai/archimedes/command"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func (a *AppInstaller) checkConfigForEmpty() error {
	currentConfigPath := filepath.Join(a.CurrentPath, "config.yaml")
	fromFile, err := os.ReadFile(currentConfigPath)
	if err != nil {
		return err
	}

	var currentConfig command.CurrentInstallConfig
	err = yaml.Unmarshal(fromFile, &currentConfig)
	if err != nil {
		return err
	}

	if a.Config.HarborPassword == "" {
		a.Config.HarborPassword = currentConfig.HarborPassword
	}

	if a.Config.ElasticPassword == "" {
		a.Config.ElasticPassword = currentConfig.ElasticPassword
	}

	if a.Config.VaultRootToken == "" {
		a.Config.VaultRootToken = currentConfig.VaultRootToken
	}

	if a.Config.VaultUnsealKey == "" {
		a.Config.VaultUnsealKey = currentConfig.VaultUnsealKey
	}

	return nil
}

func (a *AppInstaller) parseValueOverwrite(service string) (map[string]interface{}, error) {
	var unmarshalledFields map[string]interface{}

	switch service {
	case "harbor":
		harborValues, err := yaml.Marshal(a.ValueConfig.Harbor)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(harborValues, &unmarshalledFields)
	case "elastic":
		elasticValues, err := yaml.Marshal(a.ValueConfig.Elastic)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(elasticValues, &unmarshalledFields)
	}

	return unmarshalledFields, nil
}
