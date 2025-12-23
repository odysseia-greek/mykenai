package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"github.com/spf13/cobra"
)

func CreateSingleImage() *cobra.Command {
	var (
		tag             string
		destinationRepo string
		rootPath        string
		target          string
		multi           bool
	)
	cmd := &cobra.Command{
		Use:   "single [ROOT PATH]",
		Short: "Create a single image",
		Long: `This command allows you to create a single image within a specified root path, using a given tag, destination repository, and target.
If no root path is specified through the "-r" flag or positional argument, the command assumes the current working directory as the root path.`,
		Example: `  # build images from current directory
    archimedes images single -t v0.10.0
    # or 
    archimedes images single -t v0.10.0 .
  
    # build images from specified repo path
    archimedes images single -t v0.10.0 -r /path/to/image`,
		Aliases:                    []string{"s", "si"},
		SuggestionsMinimumDistance: 2,

		Run: func(cmd *cobra.Command, args []string) {
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

			if rootPath == "" {
				rootPath = argPath
			}

			if rootPath == "" {
				currentDir, err := os.Getwd()
				if err != nil {
					return
				}

				logging.Debug(fmt.Sprintf("rootPath is empty, defaulting to current dir %s", currentDir))
				rootPath = currentDir
			}

			if tag == "" {
				logging.Warn("no tag set for image, using the git short hash")
				gitTag, err := util.ExecCommandWithReturn(`git rev-parse --short HEAD`, rootPath)
				if err != nil {
					logging.Error(err.Error())
					return
				}

				tag = gitTag
			}

			if destinationRepo == "" {
				logging.Info(fmt.Sprintf("destination repo empty, default to %s", defaultRepo))
				destinationRepo = defaultRepo
			}

			err := isDockerRunning()
			if err != nil {
				return
			}

			logging.Info(fmt.Sprintf("working on repo: %s", rootPath))

			if err := BuildImage(rootPath, tag, destinationRepo, target, multi); err != nil {
				return
			}
		},
		Args: func(cmd *cobra.Command, args []string) error {
			targetFlag := cmd.Flag("target")
			if !targetFlag.Changed {
				return nil
			}
			target := targetFlag.Value.String()
			for _, validArg := range cmd.ValidArgs {
				if validArg == target {
					return nil
				}
			}
			return fmt.Errorf("invalid platform specified: %s", target)
		},
		ValidArgs: []string{"debug", "prod"},
	}

	cmd.PersistentFlags().StringVarP(&rootPath, "root", "r", "", "Root path to start building from. Interprets '.' as current directory. Defaults to current directory when no value or positional argument is provided.")
	cmd.PersistentFlags().StringVarP(&tag, "tag", "t", "", "The tag for the image, if not set it defaults to the current git commit hash.")
	cmd.PersistentFlags().StringVarP(&destinationRepo, "dest", "d", "", "The destination repository address, defaults to predefined DefaultRepo when no value is provided.")
	cmd.PersistentFlags().StringVarP(&target, "target", "g", "prod", "The target to build for, defaults to 'prod' when no value provided.")
	cmd.PersistentFlags().BoolVarP(&multi, "multi", "m", false, "Build multi arch images - default to false")

	return cmd
}

func BuildImage(rootPath, tag, destRepo, target string, multi bool) error {
	innerFiles, err := os.ReadDir(rootPath)
	if err != nil {
		return err
	}

	_, projectName := filepath.Split(rootPath)

	for _, innerFile := range innerFiles {
		if innerFile.Name() == "Dockerfile" || innerFile.Name() == "Containerfile" {
			logging.Info(fmt.Sprintf(fmt.Sprintf("working on project: %s", projectName)))
			if multi {
				if err := buildImageMultiArch(rootPath, projectName, tag, destRepo, target); err != nil {
					return err
				}
			} else {
				if err := buildImages(rootPath, projectName, tag, destRepo, target); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
