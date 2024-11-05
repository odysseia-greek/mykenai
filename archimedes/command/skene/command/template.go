package command

import (
	"bytes"
	"embed"
	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/*
var templates embed.FS

func processTemplate(templatePath, destinationPath, name, index, repoName, embed, port string) error {
	// Read the template file from the embedded FS
	tmplContent, err := templates.ReadFile(templatePath)
	if err != nil {
		return errors.Wrapf(err, "failed to read template file: %s", templatePath)
	}

	// Parse the template
	tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(tmplContent))
	if err != nil {
		return errors.Wrapf(err, "failed to parse template: %s", templatePath)
	}

	capitalizedName := cases.Title(language.English, cases.Compact).String(name)

	// Execute the template with the provided data
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		Name            string
		Port            string
		Index           string
		CapitalizedName string
		Embed           string
		RepoName        string
	}{
		Name:            name,
		Port:            port,
		Index:           index,
		CapitalizedName: capitalizedName,
		Embed:           embed,
		RepoName:        repoName,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to execute template: %s", templatePath)
	}

	// Ensure the destination directory exists
	err = os.MkdirAll(filepath.Dir(destinationPath), 0755)
	if err != nil {
		return errors.Wrapf(err, "failed to create destination directory: %s", filepath.Dir(destinationPath))
	}

	// Write the result to the destination file
	err = os.WriteFile(destinationPath, buf.Bytes(), 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to write file: %s", destinationPath)
	}

	return nil
}
