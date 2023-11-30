package install

import (
	kubernetes "github.com/odysseia-greek/agora/thales"
	"github.com/odysseia-greek/mykenai/archimedes/util/helm"
)

type AppInstaller struct {
	Namespace         string
	ConfigPath        string
	CurrentPath       string
	ThemistoklesRoot  string
	OdysseiaRoot      string
	Profile           string
	VaultSaPath       string
	VaultUnsealMethod string
	CryptoKey         string
	KeyRing           string
	Build             bool
	Minikube          bool
	AppsToInstall     []string
	ElasticConfig     ElasticOperator
	Charts            Themistokles
	Kube              kubernetes.KubeClient
	Helm              helm.HelmClient
	ValueConfig       map[string]interface{}
}

type Themistokles struct {
	Elastic       string
	ElasticSearch string
	Perikles      string
	Homeros       string
	Vault         string
	Solon         string
	Kibana        string
	Ploutarchos   string
	Xerxes        string
	Hippokrates   string
	Thermopulai   string
	Rabbitmq      string
	Eupalinos     string
	Euripides     string
	Apis          []string
}

type ElasticOperator struct {
	Name string
}

type AppsToInstall struct {
	AppsToInstall []string `yaml:"appsToInstall"`
	Include       string   `yaml:"include"`
}
