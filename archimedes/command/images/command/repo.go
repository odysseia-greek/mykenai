package command

import (
	"fmt"
	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

func CreateImagesFromRepo() *cobra.Command {
	var (
		tag             string
		destinationRepo string
		repoPath        string
		target          string
		multi           bool
	)
	cmd := &cobra.Command{
		Use:   "repo [REPO PATH]",
		Short: "Create all images from a repo",
		Long: `This command allows you to create images for all APIs within a specified repository, 
using a given tag, destination repository, and target.
If no path to the repository is specified through the "-r" flag or positional argument, 
the command assumes the current working directory as the repository path`,
		Example: `  # build images from current directory
    archimedes images repo -t v0.10.0
    # or 
    archimedes images repo -t v0.10.0 .
  
    # build images from specified repo path
    archimedes images repo -t v0.10.0 -r /path/to/repo`,
		Aliases:                    []string{"r", "re"},
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

			if repoPath == "" {
				repoPath = argPath
			}

			if repoPath == "" {
				currentDir, err := os.Getwd()
				if err != nil {
					return
				}

				logging.Debug(fmt.Sprintf("rootPath is empty, defaulting to current dir %s", currentDir))
				repoPath = currentDir
			}

			if tag == "" {
				logging.Warn("no tag set for image, using the git short hash")
				gitTag, err := util.ExecCommandWithReturn(`git rev-parse --short HEAD`, repoPath)
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

			logging.Info(fmt.Sprintf("working on repo: %s", repoPath))

			err = BuildImagesFromRepo(repoPath, tag, destinationRepo, target, multi)
			if err != nil {
				logging.Error(err.Error())
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

	cmd.PersistentFlags().StringVarP(&repoPath, "repo", "r", "", "Path to the repo build. Interprets '.' as current directory. Defaults to current directory when no value or positional argument is provided.")
	cmd.PersistentFlags().StringVarP(&tag, "tag", "t", "", "The tag for the image, if not set it defaults to the current git commit hash.")
	cmd.PersistentFlags().StringVarP(&destinationRepo, "dest", "d", "", "The destination image repo address, defaults to predefined DefaultRepo when no value is provided.")
	cmd.PersistentFlags().StringVarP(&target, "target", "g", "prod", "The target to build for, defaults to 'prod' when no value provided.")
	cmd.PersistentFlags().BoolVarP(&multi, "multi", "m", false, "Build multi arch images - default to false")

	return cmd
}

// BuildImagesFromRepo takes a repository path, tag, destination repository, and platform as input parameters and builds container images from the repository's Dockerfiles.
// It iterates through the directories in the repository and finds Dockerfiles recursively. For each Dockerfile found, it builds a multi-architecture container image using the build
func BuildImagesFromRepo(repoPath, tag, destRepo, target string, multi bool) error {
	directories, err := os.ReadDir(repoPath)
	if err != nil {
		return err
	}

	for _, dir := range directories {
		if dir.IsDir() && !strings.HasPrefix(dir.Name(), ".") {
			innerFp := filepath.Join(repoPath, dir.Name())
			innerFiles, err := os.ReadDir(innerFp)
			if err != nil {
				return err
			}

			for _, innerFile := range innerFiles {
				if innerFile.Name() == "Dockerfile" {
					logging.Info(fmt.Sprintf("working on project: %s", innerFp))
					if multi {
						if err := buildImageMultiArch(innerFp, dir.Name(), tag, destRepo, target); err != nil {
							return err
						}
					} else {
						if err := buildImages(innerFp, dir.Name(), tag, destRepo, target); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}
