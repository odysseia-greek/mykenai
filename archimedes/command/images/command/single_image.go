package command

import (
	"github.com/kpango/glg"
	settings "github.com/odysseia-greek/mykenai/archimedes/command/config/command"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

func CreateSingleImage() *cobra.Command {
	var (
		tag             string
		destinationRepo string
		name            string
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
				gitTag, err := util.ExecCommandWithReturn(`git rev-parse --short HEAD`, odysseiaSettings.OlympiaPath)
				if err != nil {
					glg.Fatal(err)
				}

				tag = gitTag
			}

			if destinationRepo == "" {
				glg.Warnf("destination repo empty, default to %s", defaultRepo)
				destinationRepo = defaultRepo
			}

			glg.Infof("filepath set to: %s", odysseiaSettings.SourcePath)

			BuildImage(odysseiaSettings.SourcePath, name, tag, destinationRepo)
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "image name")
	cmd.PersistentFlags().StringVarP(&tag, "tag", "t", "", "image tag")
	cmd.PersistentFlags().StringVarP(&destinationRepo, "dest", "d", "", "destination repo address")

	return cmd
}

func BuildImage(rootPath, name, tag, destRepo string) {
	directories, err := os.ReadDir(rootPath)
	if err != nil {
		glg.Fatal(err)
	}

	for _, dir := range directories {
		if dir.IsDir() && !strings.HasPrefix(dir.Name(), ".") {
			repoPath := filepath.Join(rootPath, dir.Name())
			innerDirs, err := os.ReadDir(repoPath)
			if err != nil {
				glg.Fatal(err)
			}

			if dir.Name() == name {
				for _, innerFile := range innerDirs {
					if innerFile.Name() == "Dockerfile" {
						glg.Infof("working on project: %s", name)
						err := runImageBuildFlow(repoPath, name, tag, destRepo)
						if err != nil {
							glg.Error(err)
						}
						return
					}
				}
			}

			for _, innerDir := range innerDirs {
				if innerDir.Name() == name {
					if innerDir.IsDir() && !strings.HasPrefix(innerDir.Name(), ".") {
						innerFp := filepath.Join(rootPath, dir.Name(), innerDir.Name())
						innerFiles, err := os.ReadDir(innerFp)
						if err != nil {
							glg.Fatal(err)
						}

						for _, innerFile := range innerFiles {
							if innerFile.Name() == "Dockerfile" {
								glg.Infof("working on project: %s", innerFp)
								err := runImageBuildFlow(innerFp, innerDir.Name(), tag, destRepo)
								if err != nil {
									glg.Error(err)
									return
								}
							}
						}
					}
				}
			}
		}
	}
}
