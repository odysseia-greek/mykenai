package command

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/spf13/cobra"
)

func Tidy() *cobra.Command {
	var rootPath string

	cmd := &cobra.Command{
		Use:   "tidy",
		Short: "run go mod tidy for all modules under a root",
		Long:  `Find every go.mod below a root path and run go mod tidy in each module.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedRoot, err := resolveDocsRoot(rootPath, args)
			if err != nil {
				return err
			}

			return runGoModTidy(resolvedRoot)
		},
	}

	cmd.PersistentFlags().StringVarP(&rootPath, "root", "r", "", "root path to scan for go.mod files")

	return cmd
}

func runGoModTidy(rootPath string) error {
	modules, err := discoverGoModules(rootPath)
	if err != nil {
		return err
	}

	if len(modules) == 0 {
		logging.System(fmt.Sprintf("No go.mod files found under %s", rootPath))
		return nil
	}

	for _, moduleDir := range modules {
		logging.System(fmt.Sprintf("Running go mod tidy in %s", moduleDir))
		if err := runTidyCommand(moduleDir, "go", "mod", "tidy"); err != nil {
			return err
		}
	}

	logging.System(fmt.Sprintf("Ran go mod tidy in %d module(s)", len(modules)))
	return nil
}

func discoverGoModules(rootPath string) ([]string, error) {
	modules := make([]string, 0)
	seen := make(map[string]struct{})

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			switch d.Name() {
			case ".git", ".idea", "node_modules", ".archimedes-tools":
				return filepath.SkipDir
			}
			return nil
		}

		if d.Name() != "go.mod" {
			return nil
		}

		moduleDir := filepath.Dir(path)
		if _, ok := seen[moduleDir]; ok {
			return nil
		}
		seen[moduleDir] = struct{}{}
		modules = append(modules, moduleDir)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(modules)
	return modules, nil
}

func runTidyCommand(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return cmd.Run()
}
