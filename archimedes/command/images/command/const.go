package command

const (
	distDirectory string = "dist"
	binDirectory  string = "bin"
	linux         string = "linux"
	defaultRepo   string = "ghcr.io/odysseia-greek"
	docsDest      string = "ploutarchos"
	yamlDest      string = "yaml"
	defaultTests  string = "tests"
	olympos       string = "olympos"
	delphi        string = "delphi"
	knossos       string = "knossos"
	ionia         string = "ionia"
	pheidias      string = "pheidias"
	ploutarchos   string = "ploutarchos"
	dockerFile    string = "Dockerfile"
	hippokrates   string = "hippokrates"
	xerxes        string = "xerxes"
	eupalinos     string = "eupalinos"
	attike        string = "attike"
)

var ARCHS = [...]string{"amd64", "arm64"}
var differentFlow = [...]string{pheidias, xerxes}
var sourceDirs = [...]string{olympos, delphi, knossos, ionia, pheidias, hippokrates, xerxes, eupalinos, attike}
