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
