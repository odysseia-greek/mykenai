package install

import (
	"github.com/odysseia-greek/mykenai/archimedes/command"
	"github.com/odysseia-greek/mykenai/archimedes/util/helm"
	"github.com/odysseia-greek/plato/harbor"
	kubernetes "github.com/odysseia-greek/thales"
)

type AppInstaller struct {
	Namespace        string
	ConfigPath       string
	CurrentPath      string
	ThemistoklesRoot string
	OdysseiaRoot     string
	Profile          string
	Build            bool
	Minikube         bool
	AppsToInstall    []string
	ElasticConfig    ElasticOperator
	Charts           Themistokles
	Kube             kubernetes.KubeClient
	Helm             helm.HelmClient
	Harbor           harbor.Client
	Config           command.CurrentInstallConfig
	ValueConfig      ValueOverwrite
}

type Themistokles struct {
	Elastic       string
	ElasticSearch string
	Perikles      string
	Homeros       string
	Vault         string
	Solon         string
	Harbor        string
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

type ValueOverwrite struct {
	Harbor struct {
		HarborAdminPassword string `yaml:"harborAdminPassword"`
		Expose              struct {
			Type string `yaml:"type"`
			TLS  struct {
				Enabled    bool   `yaml:"enabled"`
				CertSource string `yaml:"certSource"`
				Secret     struct {
					SecretName string `yaml:"secretName"`
				} `yaml:"secret"`
			} `yaml:"tls"`
		} `yaml:"expose"`
		ExternalURL string `yaml:"externalURL"`
		NodePort    struct {
			Name  string `yaml:"name"`
			Ports struct {
				HTTP struct {
					Port     int `yaml:"port"`
					NodePort int `yaml:"nodePort"`
				} `yaml:"http"`
				HTTPS struct {
					Port     int `yaml:"port"`
					NodePort int `yaml:"nodePort"`
				} `yaml:"https"`
			} `yaml:"ports"`
		} `yaml:"nodePort"`
	} `yaml:"harbor"`
	Elastic struct {
		VolumeClaimTemplate struct {
			AccessModes      []string `yaml:"accessModes"`
			StorageClassName string   `yaml:"storageClassName"`
			Resources        struct {
				Requests struct {
					Storage string `yaml:"storage"`
				} `yaml:"requests"`
			} `yaml:"resources"`
		} `yaml:"volumeClaimTemplate"`
	} `yaml:"elastic"`
	Vault struct {
		Global struct {
			Enabled    bool `yaml:"enabled"`
			TLSDisable bool `yaml:"tlsDisable"`
		} `yaml:"global"`
		Server struct {
			ExtraEnvironmentVars struct {
				VAULTCACERT string `yaml:"VAULT_CACERT"`
			} `yaml:"extraEnvironmentVars"`
			ExtraVolumes []struct {
				Type string `yaml:"type"`
				Name string `yaml:"name"`
			} `yaml:"extraVolumes"`
			Standalone struct {
				Enabled bool   `yaml:"enabled"`
				Config  string `yaml:"config"`
			} `yaml:"standalone"`
		} `yaml:"server"`
	} `yaml:"vault"`
}
