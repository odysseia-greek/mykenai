package command

type Application struct {
	Deployment string `yaml:"deployment,omitempty"`
	Init       string `yaml:"init,omitempty"`
	Seeder     string `yaml:"seeder,omitempty"`
	Sidecar    string `yaml:"sidecar,omitempty"`
	JobInit    string `yaml:"jobinit,omitempty"`
	Job        string `yaml:"job,omitempty"`
	System     string `yaml:"system,omitempty"`
	Load       string `yaml:"load,omitempty"`
	Stateful   string `yaml:"stateful,omitempty"`
	Tracer     string `yaml:"tracer,omitempty"`
	InitSeeder string `yaml:"initSeeder,omitempty"`
}

type Config struct {
	Alexandros  Application `yaml:"alexandros,omitempty"`
	Dionysios   Application `yaml:"dionysios,omitempty"`
	Euripides   Application `yaml:"euripides,omitempty"`
	Herodotos   Application `yaml:"herodotos,omitempty"`
	Sokrates    Application `yaml:"sokrates,omitempty"`
	Solon       Application `yaml:"solon,omitempty"`
	Homeros     Application `yaml:"homeros,omitempty"`
	Perikles    Application `yaml:"perikles,omitempty"`
	Melissos    Application `yaml:"melissos,omitempty"`
	Pheidias    Application `yaml:"pheidias,omitempty"`
	Ploutarchos Application `yaml:"ploutarchos,omitempty"`
	Hippokrates Application `yaml:"hippokrates,omitempty"`
	Xerxes      Application `yaml:"xerxes,omitempty"`
	Eupalinos   Application `yaml:"eupalinos,omitempty"`
}

type ImageValues struct {
	Images ImagesConfig `yaml:"images"`
}

type ImagesConfig struct {
	ImageRepo   string `yaml:"imageRepo"`
	PullSecret  string `yaml:"pullSecret"`
	Sidecar     Repo   `yaml:"sidecar,omitempty"`
	Init        Repo   `yaml:"init,omitempty"`
	OdysseiaAPI Repo   `yaml:"odysseiaapi,omitempty"`
	Seeder      Repo   `yaml:"seeder,omitempty"`
	Job         Repo   `yaml:"job,omitempty"`
	JobInit     Repo   `yaml:"jobinit,omitempty"`
	System      Repo   `yaml:"system,omitempty"`
	Load        Repo   `yaml:"load,omitempty"`
	Stateful    Repo   `yaml:"stateful,omitempty"`
	Tracer      Repo   `yaml:"tracer,omitempty"`
	InitSeeder  Repo   `yaml:"initSeeder,omitempty"`
}

type Repo struct {
	Repo string `yaml:"repo,omitempty"`
	Tag  string `yaml:"tag,omitempty"`
}
