package command

import (
	"fmt"
	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
)

func createScaffoldedService(name, indexName, repoName, destination, embed, port, templateType string) (string, error) {
	// Define template files and their destination paths
	files := map[string]string{
		"templates/Dockerfile.tpl": "Dockerfile",
		"templates/.air.toml.tpl":  ".air.toml",
	}

	if templateType == "api" {
		files["templates/api/main.go.tpl"] = "main.go"
		files["templates/api/app/routes.go.tpl"] = "app/routes.go"
		files["templates/api/app/handlers.go.tpl"] = "app/handlers.go"
		files["templates/api/app/config.go.tpl"] = "app/config.go"
		files["templates/api/infra/skaffold.yaml.tpl"] = "infra/skaffold-profile.yaml"
		files["templates/api/infra/helmfile.yaml.tpl"] = "infra/helmfile.yaml"
	} else if templateType == "job" {
		files["templates/job/main.go.tpl"] = "main.go"
		files["templates/job/seeder/config.go.tpl"] = "seeder/config.go"
		files["templates/job/seeder/handler.go.tpl"] = "seeder/handler.go"
		files["templates/job/seeder/index.go.tpl"] = "seeder/index.go"
		files["templates/job/seeder/models.go.tpl"] = "seeder/models.go"
	}

	directoryToCreate := filepath.Join(destination, name)

	// Create the destination directory if it doesn't exist
	err := os.MkdirAll(directoryToCreate, 0755)
	if err != nil {
		return directoryToCreate, errors.Wrap(err, "failed to create destination directory")
	}

	// Process each template file
	for tmplPath, destPath := range files {
		err = processTemplate(tmplPath, filepath.Join(directoryToCreate, destPath), name, indexName, repoName, embed, port)
		if err != nil {
			return directoryToCreate, err
		}
	}

	return directoryToCreate, nil
}

func initGolang(destinationPath, repoName, name string) error {
	commands := []struct {
		Cmd  string
		Args []string
	}{
		{"go", []string{"mod", "init", fmt.Sprintf("github.com/odysseia-greek/%s/%s", repoName, name)}},
		{"go", []string{"mod", "tidy"}},
		{"go", []string{"fmt", "./..."}},
	}

	for _, command := range commands {
		cmd := exec.Command(command.Cmd, command.Args...)
		cmd.Dir = destinationPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			return errors.Wrapf(err, "failed to run command %s %v: %s", command.Cmd, command.Args, string(output))
		}
		logging.Debug(fmt.Sprintf("ran command %s %v: %s", command.Cmd, command.Args, string(output)))
	}

	return nil
}
