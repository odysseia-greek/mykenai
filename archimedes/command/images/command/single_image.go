package command

import (
	"github.com/kpango/glg"
	settings "github.com/odysseia-greek/mykenai/archimedes/command/config/command"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"github.com/spf13/cobra"
	"io/ioutil"
	"path/filepath"
)

func CreateSingleImage() *cobra.Command {
	var (
		tag             string
		destinationRepo string
		name            string
		minikube        bool
	)
	cmd := &cobra.Command{
		Use:   "single",
		Short: "create single image",
		Long: `Allows you to create images for all apis
- Filepath
`,
		Run: func(cmd *cobra.Command, args []string) {
			glg.Green("creating")

			odysseiaSettings, err := settings.ReadOutConfig()
			if err != nil {
				glg.Error(err)
			}

			if tag == "" {
				glg.Warn("no tag set for image, using the git short hash")
				gitTag, err := util.ExecCommandWithReturn(`git rev-parse --short HEAD`, odysseiaSettings.OlymposPath)
				if err != nil {
					glg.Fatal(err)
				}

				tag = gitTag
			}

			if destinationRepo != "" {
				minikube = false
			}

			if destinationRepo == "" && !minikube {
				glg.Warnf("destination repo empty, default to %s", defaultRepo)
				destinationRepo = defaultRepo
			}

			glg.Infof("filepath set to: %s", odysseiaSettings.SourcePath)

			BuildImage(odysseiaSettings.SourcePath, name, tag, destinationRepo, minikube)
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "image name")
	cmd.PersistentFlags().StringVarP(&tag, "tag", "t", "", "image tag")
	cmd.PersistentFlags().StringVarP(&destinationRepo, "dest", "d", "", "destination repo address")
	cmd.PersistentFlags().BoolVarP(&minikube, "minikube", "m", false, "if minikube is used images will be loaded into minikube instead of pushing to a remote repo")

	return cmd
}

func BuildImage(rootPath, name, tag, dest string, minikube bool) {
	directories, err := ioutil.ReadDir(rootPath)
	if err != nil {
		glg.Fatal(err)
	}

	var filePath string
	found := false
	for _, dir := range directories {
		if found {
			break
		}

		if dir.Name() == name {
			filePath = filepath.Join(rootPath, dir.Name())
			found = true
			break
		}

		for _, source := range sourceDirs {
			if found {
				break
			}
			fp := filepath.Join(rootPath, source)
			innerDirs, err := ioutil.ReadDir(fp)
			if err != nil {
				glg.Fatal(err)
			}
			for _, innerDir := range innerDirs {
				if innerDir.Name() == name {
					filePath = filepath.Join(rootPath, source)
					found = true
					break
				}
			}
		}
	}

	for _, flow := range differentFlow {
		if flow == name {
			err = buildImageWithLocalFile(filePath, name, tag, dest)
			if err != nil {
				glg.Error(err)
				return
			}
			return
		}
	}

	apiPath := filePath
	if name != hippokrates && name != eupalinos {
		apiPath = filepath.Join(filePath, name)
	}

	err = runImageBuildFlow(filePath, apiPath, name, tag, dest)
	if err != nil {
		glg.Error(err)
		return
	}
}
