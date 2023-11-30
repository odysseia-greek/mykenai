package command

type BaseConfig struct {
	HelmChartPath  string `yaml:"helm-path"`
	SourceCodePath string `yaml:"source-path"`
}
