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

func CreateImagesFromRepo() *cobra.Command {
	var (
		tag             string
		destinationRepo string
		repoName        string
		minikube        bool
	)
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "create all images from a repo",
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

			if repoName == "" {
				glg.Fatal("no repo set cannot continue")
			}

			glg.Infof("filepath set to: %s", odysseiaSettings.SourcePath)
			glg.Infof("working on repo: %s", repoName)

			BuildImagesFromRepo(odysseiaSettings.SourcePath, repoName, tag, destinationRepo)
		},
	}

	cmd.PersistentFlags().StringVarP(&repoName, "repo", "r", "", "repo name")
	cmd.PersistentFlags().StringVarP(&tag, "tag", "t", "", "image tag")
	cmd.PersistentFlags().StringVarP(&destinationRepo, "dest", "d", "", "destination image repo address")
	cmd.PersistentFlags().BoolVarP(&minikube, "minikube", "m", false, "if minikube is used images will be loaded into minikube instead of pushing to a remote repo")

	return cmd
}

func BuildImagesFromRepo(rootPath, repoName, tag, destRepo string) {
	directories, err := ioutil.ReadDir(rootPath)
	if err != nil {
		glg.Fatal(err)
	}

	for _, dir := range directories {
		if dir.Name() == repoName {

			repoPath := filepath.Join(rootPath, dir.Name())
			innerDirs, err := ioutil.ReadDir(repoPath)
			if err != nil {
				glg.Fatal(err)
			}

			for _, innerDir := range innerDirs {
				if innerDir.IsDir() && !strings.HasPrefix(innerDir.Name(), ".") {
					innerFp := filepath.Join(rootPath, dir.Name(), innerDir.Name())
					innerFiles, err := ioutil.ReadDir(innerFp)
					if err != nil {
						glg.Fatal(err)
					}

					for _, innerFile := range innerFiles {
						if innerFile.Name() == "main.go" {
							glg.Infof("working on project: %s", innerFp)
							runImageBuildFlow(repoPath, innerFp, innerDir.Name(), tag, destRepo)
						}
					}
				}
			}
		}
	}
}
