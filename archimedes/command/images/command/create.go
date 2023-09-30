package command

import (
	"github.com/kpango/glg"
	settings "github.com/odysseia-greek/mykenai/archimedes/command/config/command"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"github.com/spf13/cobra"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func CreateImages() *cobra.Command {
	var (
		tag             string
		destinationRepo string
		minikube        bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create images for all apis",
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

			LoopAndCreateImages(odysseiaSettings.SourcePath, tag, destinationRepo, minikube)
		},
	}

	cmd.PersistentFlags().StringVarP(&tag, "tag", "t", "", "image tag")
	cmd.PersistentFlags().StringVarP(&destinationRepo, "dest", "d", "", "destination repo address")
	cmd.PersistentFlags().BoolVarP(&minikube, "minikube", "m", false, "if minikube is used images will be loaded into minikube instead of pushing to a remote repo")

	return cmd
}

func LoopAndCreateImages(filePath, tag, destRepo string, minikube bool) {
	directories, err := ioutil.ReadDir(filePath)
	if err != nil {
		glg.Fatal(err)
	}

	for _, dir := range directories {
		for _, source := range sourceDirs {
			if source == dir.Name() {
				fp := filepath.Join(filePath, dir.Name())
				innerDirs, err := ioutil.ReadDir(fp)
				if err != nil {
					glg.Fatal(err)
				}

				for _, innerDir := range innerDirs {
					if innerDir.Name() == "package.json" {
						if destRepo == defaultRepo {
							continue
						}
						err := buildImageWithLocalFile(fp, dir.Name(), tag, destRepo)
						if err != nil {
							glg.Fatal(err)
						}

						continue
					}
					if innerDir.IsDir() && !strings.HasPrefix(innerDir.Name(), ".") {
						innerFp := filepath.Join(filePath, dir.Name(), innerDir.Name())
						innerFiles, err := ioutil.ReadDir(innerFp)
						if err != nil {
							glg.Fatal(err)
						}

						for _, innerFile := range innerFiles {
							if innerFile.Name() == "main.go" {
								runImageBuildFlow(fp, innerFp, innerDir.Name(), tag, destRepo)
							}
						}
					}

				}

			}
		}
	}
}
