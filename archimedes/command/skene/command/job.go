package command

import (
	"fmt"
	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
)

func CreateJob() *cobra.Command {
	var (
		embed       string
		name        string
		repoName    string
		indexName   string
		destination string
	)
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Create a job from templates",
		Long:  `Creates a job from templates.`,
		Run: func(cmd *cobra.Command, args []string) {

			if name == "" {
				logging.Error("cannot create a new api without a name")
				return
			}
			argPath := ""
			if len(args) > 0 {
				argPath = args[len(args)-1]
			}

			if argPath == "." {
				currentDir, err := os.Getwd()
				if err != nil {
					return
				}
				argPath = currentDir
			}

			if destination == "" {
				destination = argPath
			}

			if destination == "" {
				currentDir, err := os.Getwd()
				if err != nil {
					return
				}

				logging.Debug(fmt.Sprintf("destination is empty, defaulting to current dir %s", currentDir))
				destination = currentDir
			}

			destinationPath, err := createScaffoldedService(name, indexName, repoName, destination, embed, "", "job")
			if err != nil {
				logging.Error(errors.Wrap(err, "Failed to create api scaffold").Error())
				return
			}

			err = initGolang(destinationPath, repoName, name)
			if err != nil {
				logging.Error(errors.Wrap(err, "Failed to init golang").Error())
				return
			}

			logging.Info(fmt.Sprintf("Created api with name %s, make sure to rename the app directory to something more on topic and add the quotes at main.go!", name))

		},
	}
	cmd.PersistentFlags().StringVarP(&name, "name", "n", "", "name of the service to be created")
	cmd.PersistentFlags().StringVarP(&repoName, "repo", "r", "", "name of the repo for go mod init")
	cmd.PersistentFlags().StringVarP(&indexName, "index", "i", "", "name of index for elastic")
	cmd.PersistentFlags().StringVarP(&embed, "embed", "e", "", "name of the embed directory")
	cmd.PersistentFlags().StringVarP(&destination, "destination", "d", "", "where to create the service")

	return cmd
}
