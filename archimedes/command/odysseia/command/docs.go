package command

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/spf13/cobra"
)

const (
	protocGenDocModule  = "github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest"
	spectaqlPackage     = "spectaql"
	defaultToolsDirName = ".archimedes-tools"
	bufDocsTemplateName = "buf.gen.docs.yaml"
	spectaqlConfigName  = "spectaql.yaml"
	spectaqlDocsDirName = "spectaql"
)

type docsTarget struct {
	serviceName     string
	servicePath     string
	bufTemplatePath string
	spectaqlDir     string
	spectaqlConfig  string
}

func GenerateDocs() *cobra.Command {
	var rootPath string

	cmd := &cobra.Command{
		Use:   "docs",
		Short: "generate docs",
		Long:  `Discover proto and GraphQL doc configs below a root path and generate the docs without make.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedRoot, err := resolveDocsRoot(rootPath, args)
			if err != nil {
				return err
			}

			return generateDocs(resolvedRoot)
		},
	}

	cmd.PersistentFlags().StringVarP(&rootPath, "root", "r", "", "root path containing the service directories to scan")

	return cmd
}

func resolveDocsRoot(rootPath string, args []string) (string, error) {
	if rootPath == "" && len(args) > 0 {
		rootPath = args[len(args)-1]
	}

	if rootPath == "" || rootPath == "." {
		currentDir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		rootPath = currentDir
	}

	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", absRoot)
	}

	return absRoot, nil
}

func generateDocs(rootPath string) error {
	targets, err := discoverDocsTargets(rootPath)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return fmt.Errorf("found no %s or SpectaQL configs under %s", bufDocsTemplateName, rootPath)
	}

	toolsDir := filepath.Join(rootPath, defaultToolsDirName)
	toolsBinDir := filepath.Join(toolsDir, "bin")

	var protoTargets []docsTarget
	var spectaqlTargets []docsTarget
	for _, target := range targets {
		if target.bufTemplatePath != "" {
			protoTargets = append(protoTargets, target)
		}
		if target.spectaqlConfig != "" {
			spectaqlTargets = append(spectaqlTargets, target)
		}
	}

	if len(protoTargets) > 0 {
		if err := ensureProtocGenDoc(toolsBinDir); err != nil {
			return err
		}
		if _, err := exec.LookPath("buf"); err != nil {
			return fmt.Errorf("buf is required to generate proto docs: %w", err)
		}
	}

	spectaqlExec := "spectaql"
	if len(spectaqlTargets) > 0 {
		spectaqlExec, err = ensureSpectaql(toolsDir)
		if err != nil {
			return err
		}
	}

	for _, target := range protoTargets {
		if err := generateProtoDocs(target, toolsBinDir); err != nil {
			return err
		}
	}

	for _, target := range spectaqlTargets {
		if err := generateSpectaqlDocs(target, spectaqlExec); err != nil {
			return err
		}
	}

	logging.System(fmt.Sprintf("Generated docs for %d service(s)", len(targets)))
	return nil
}

func discoverDocsTargets(rootPath string) ([]docsTarget, error) {
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	var targets []docsTarget
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		servicePath := filepath.Join(rootPath, entry.Name())
		target := docsTarget{
			serviceName: entry.Name(),
			servicePath: servicePath,
		}

		bufTemplatePath := filepath.Join(servicePath, bufDocsTemplateName)
		if fileExists(bufTemplatePath) {
			target.bufTemplatePath = bufTemplatePath
		}

		spectaqlDir, spectaqlConfig := discoverSpectaqlConfig(servicePath)
		target.spectaqlDir = spectaqlDir
		target.spectaqlConfig = spectaqlConfig

		if target.bufTemplatePath != "" || target.spectaqlConfig != "" {
			targets = append(targets, target)
		}
	}

	sort.Slice(targets, func(i, j int) bool {
		return targets[i].serviceName < targets[j].serviceName
	})

	return targets, nil
}

func discoverSpectaqlConfig(servicePath string) (string, string) {
	directConfig := filepath.Join(servicePath, "docs", spectaqlConfigName)
	if fileExists(directConfig) {
		return filepath.Dir(directConfig), directConfig
	}

	nestedDir := filepath.Join(servicePath, "docs", spectaqlDocsDirName)
	nestedConfig := filepath.Join(nestedDir, spectaqlConfigName)
	if fileExists(nestedConfig) {
		return nestedDir, nestedConfig
	}

	return "", ""
}

func ensureProtocGenDoc(toolsBinDir string) error {
	binaryPath := filepath.Join(toolsBinDir, "protoc-gen-doc")
	if fileExists(binaryPath) {
		return nil
	}

	if err := os.MkdirAll(toolsBinDir, 0o755); err != nil {
		return err
	}

	logging.System("Installing protoc-gen-doc into local tools directory")
	return runDocsCommandWithEnv("", []string{"GOBIN=" + toolsBinDir}, "go", "install", protocGenDocModule)
}

func ensureSpectaql(toolsDir string) (string, error) {
	if systemBinary, err := exec.LookPath("spectaql"); err == nil {
		return systemBinary, nil
	}

	if _, err := exec.LookPath("npm"); err != nil {
		return "", fmt.Errorf("spectaql is not installed and npm is unavailable: %w", err)
	}

	installRoot := filepath.Join(toolsDir, "spectaql")
	binPath := filepath.Join(installRoot, "node_modules", ".bin", "spectaql")
	if fileExists(binPath) {
		return binPath, nil
	}

	if err := os.MkdirAll(installRoot, 0o755); err != nil {
		return "", err
	}

	logging.System("Installing spectaql into local tools directory")
	if err := runDocsCommand("", "npm", "install", "--prefix", installRoot, spectaqlPackage); err != nil {
		return "", err
	}

	if !fileExists(binPath) {
		return "", fmt.Errorf("spectaql installation completed but binary was not found at %s", binPath)
	}

	return binPath, nil
}

func generateProtoDocs(target docsTarget, toolsBinDir string) error {
	logging.System(fmt.Sprintf("Generating proto docs for %s", target.serviceName))

	pathEnv := os.Getenv("PATH")
	extraEnv := []string{"PATH=" + toolsBinDir + string(os.PathListSeparator) + pathEnv}

	return runDocsCommandWithEnv(
		filepath.Dir(target.servicePath),
		extraEnv,
		"buf", "generate", "--template", target.bufTemplatePath, target.servicePath,
	)
}

func generateSpectaqlDocs(target docsTarget, spectaqlExec string) error {
	logging.System(fmt.Sprintf("Generating SpectaQL docs for %s", target.serviceName))
	return runDocsCommand(target.spectaqlDir, spectaqlExec, "-c", target.spectaqlConfig)
}

func runDocsCommand(dir string, name string, args ...string) error {
	return runDocsCommandWithEnv(dir, nil, name, args...)
}

func runDocsCommandWithEnv(dir string, extraEnv []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), extraEnv...)

	return cmd.Run()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !info.IsDir()
}
