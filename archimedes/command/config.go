package command

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
)

type CurrentInstallConfig struct {
	ElasticPassword string `yaml:"elastic-password"`
	HarborPassword  string `yaml:"harbor-password"`
	VaultRootToken  string `yaml:"vault-root-token"`
	VaultUnsealKey  string `yaml:"vault-unseal-key"`
}

type BaseConfig struct {
	HelmChartPath  string `yaml:"helm-path"`
	SourceCodePath string `yaml:"source-path"`
}

func GetClusterKeys() (*CurrentInstallConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	fileName := "config.yaml"
	clusterKeyFilePath := filepath.Join(homeDir, ".odysseia", "current", fileName)
	f, err := ioutil.ReadFile(clusterKeyFilePath)
	if err != nil {
		return nil, err
	}

	var currentKeys CurrentInstallConfig
	err = yaml.Unmarshal(f, &currentKeys)
	if err != nil {
		return nil, err
	}

	return &currentKeys, nil
}

func CreateBaseConfig() {

}

func GetBaseConfig(path string) {

}
