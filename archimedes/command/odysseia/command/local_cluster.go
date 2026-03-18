package command

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/spf13/cobra"
)

const (
	defaultFluxEnv     = "romaioi"
	defaultFluxNS      = "flux-system"
	defaultFluxSource  = "themistokles"
	defaultFluxRepoURL = "https://github.com/odysseia-greek/mykenai.git"
	defaultFluxBranch  = "main"
	defaultSOPSSecret  = "sops-gpg"
	defaultSOPSKeyFP   = "3C8B5BB6281C34C5E80C473086FCFB28CF0EC482"
)

type localClusterOptions struct {
	ha   bool
	env  string
	root string
}

type commandStep struct {
	description string
	dir         string
	env         []string
	name        string
	args        []string
}

func Create() *cobra.Command {
	opts := &localClusterOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a local odysseia cluster",
		Long: `Create a local odysseia cluster using the existing Lima, Ansible,
k0s bootstrap, and Flux workflow.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(opts)
		},
	}

	addCreateFlags(cmd, opts)

	return cmd
}

func Delete() *cobra.Command {
	opts := &localClusterOptions{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete a local odysseia cluster",
		Long:  `Delete the local odysseia cluster and the Lima disks used for it.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(opts)
		},
	}

	addClusterFlags(cmd, opts)

	return cmd
}

func Restart() *cobra.Command {
	opts := &localClusterOptions{}

	cmd := &cobra.Command{
		Use:   "restart",
		Short: "recreate a local odysseia cluster",
		Long:  `Delete the local odysseia cluster and create it again.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runDelete(opts); err != nil {
				return err
			}

			return runCreate(opts)
		},
	}

	addCreateFlags(cmd, opts)

	return cmd
}

func addClusterFlags(cmd *cobra.Command, opts *localClusterOptions) {
	cmd.PersistentFlags().BoolVar(&opts.ha, "ha", false, "use the Lima HA topology instead of the default single-node topology")
	cmd.PersistentFlags().StringVar(&opts.root, "root", "", "repo root path; defaults to the current directory or one of its parents")
}

func addCreateFlags(cmd *cobra.Command, opts *localClusterOptions) {
	addClusterFlags(cmd, opts)
	cmd.PersistentFlags().StringVar(&opts.env, "env", defaultFluxEnv, "Flux environment to bootstrap")
}

func runCreate(opts *localClusterOptions) error {
	repoRoot, err := resolveRepoRoot(opts.root)
	if err != nil {
		return err
	}

	logging.System(fmt.Sprintf("Creating local odysseia cluster (ha=%t, env=%s)", opts.ha, opts.env))

	limaDir := filepath.Join(repoRoot, "lykourgos", "lima")
	ansibleDir := filepath.Join(repoRoot, "lykourgos", "ansible")
	k0sDir := filepath.Join(repoRoot, "lykourgos", "k0s")
	themistoklesDir := filepath.Join(repoRoot, "themistokles")

	if err := ensureLimaDisks(opts.ha, limaDir); err != nil {
		return err
	}

	for _, step := range limaCreateSteps(opts.ha, limaDir) {
		if err := runStep(step); err != nil {
			return err
		}
	}

	for _, step := range ansibleSteps(opts.ha, ansibleDir) {
		if err := runStep(step); err != nil {
			return err
		}
	}

	for _, step := range bootstrapSteps(opts.ha, ansibleDir, k0sDir) {
		if err := runStep(step); err != nil {
			return err
		}
	}

	for _, step := range fluxBootstrapSteps(opts.ha, opts.env, ansibleDir, themistoklesDir) {
		if err := runStep(step); err != nil {
			return err
		}
	}

	logging.System("Local odysseia cluster is ready")

	return nil
}

func runDelete(opts *localClusterOptions) error {
	repoRoot, err := resolveRepoRoot(opts.root)
	if err != nil {
		return err
	}

	logging.System(fmt.Sprintf("Deleting local odysseia cluster (ha=%t)", opts.ha))

	limaDir := filepath.Join(repoRoot, "lykourgos", "lima")
	for _, step := range limaDeleteSteps(opts.ha, limaDir) {
		if err := runStepAllowFailure(step); err != nil {
			return err
		}
	}

	logging.System("Local odysseia cluster deleted")

	return nil
}

func resolveRepoRoot(root string) (string, error) {
	if root != "" {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			return "", err
		}
		if err := validateRepoRoot(absRoot); err != nil {
			return "", err
		}

		return absRoot, nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := currentDir
	for {
		if err := validateRepoRoot(dir); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("failed to locate repo root from %s", currentDir)
}

func validateRepoRoot(root string) error {
	required := []string{
		filepath.Join(root, "archimedes"),
		filepath.Join(root, "lykourgos"),
		filepath.Join(root, "themistokles"),
	}

	for _, path := range required {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", path)
		}
	}

	return nil
}

func limaCreateSteps(ha bool, limaDir string) []commandStep {
	if ha {
		return []commandStep{
			{
				description: "start Lima HA controller",
				dir:         limaDir,
				name:        "limactl",
				args:        []string{"start", "--name=k0s-controller", "k0s-ha.yaml"},
			},
			{
				description: "wait for Lima controller token",
				dir:         limaDir,
				name:        "sleep",
				args:        []string{"30"},
			},
			{
				description: "start Lima HA worker 1",
				dir:         limaDir,
				name:        "limactl",
				args:        []string{"start", "--name=k0s-worker1", "--set=.networks[0].macAddress=52:55:55:12:34:61", "--set=.additionalDisks[0].name=pyxis-worker1", "k0s-ha-worker.yaml"},
			},
			{
				description: "start Lima HA worker 2",
				dir:         limaDir,
				name:        "limactl",
				args:        []string{"start", "--name=k0s-worker2", "--set=.networks[0].macAddress=52:55:55:12:34:62", "--set=.additionalDisks[0].name=pyxis-worker2", "k0s-ha-worker.yaml"},
			},
			{
				description: "join Lima HA worker 1",
				dir:         limaDir,
				name:        "/bin/sh",
				args: []string{
					"-c",
					`TOKEN="$(limactl shell k0s-controller cat /tmp/worker-token.txt)" && printf "%s\n" "$TOKEN" | limactl shell k0s-worker1 sudo tee /tmp/worker-token.txt >/dev/null && limactl shell k0s-worker1 sudo k0s install worker --token-file /tmp/worker-token.txt && limactl shell k0s-worker1 sudo systemctl start k0sworker && limactl shell k0s-worker1 sudo systemctl enable k0sworker`,
				},
			},
			{
				description: "join Lima HA worker 2",
				dir:         limaDir,
				name:        "/bin/sh",
				args: []string{
					"-c",
					`TOKEN="$(limactl shell k0s-controller cat /tmp/worker-token.txt)" && printf "%s\n" "$TOKEN" | limactl shell k0s-worker2 sudo tee /tmp/worker-token.txt >/dev/null && limactl shell k0s-worker2 sudo k0s install worker --token-file /tmp/worker-token.txt && limactl shell k0s-worker2 sudo systemctl start k0sworker && limactl shell k0s-worker2 sudo systemctl enable k0sworker`,
				},
			},
		}
	}

	return []commandStep{
		{
			description: "start Lima single-node cluster",
			dir:         limaDir,
			name:        "limactl",
			args:        []string{"start", "--yes", "--name=k0s-byzantium", "k0s-single.yaml"},
		},
	}
}

func ansibleSteps(ha bool, ansibleDir string) []commandStep {
	playbook := "playbooks/k0s-lima-single.yaml"
	if ha {
		playbook = "playbooks/k0s-lima-ha.yaml"
	}

	return []commandStep{
		{
			description: "run Ansible on Lima cluster",
			dir:         ansibleDir,
			name:        "ansible-playbook",
			args:        []string{"-i", "inventories/lima/hosts.yaml", playbook},
		},
	}
}

func bootstrapSteps(ha bool, ansibleDir, k0sDir string) []commandStep {
	env := []string{}
	if ha {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", filepath.Join(ansibleDir, "playbooks", "k0s-kubeconfig-k0s-controller")))
	}

	return []commandStep{
		{
			description: "bootstrap k0s cluster",
			dir:         k0sDir,
			env:         env,
			name:        "./bootstrap.sh",
		},
	}
}

func fluxBootstrapSteps(ha bool, envName, ansibleDir, themistoklesDir string) []commandStep {
	pathInRepo := filepath.ToSlash(filepath.Join(".", "themistokles", "aer", envName))
	env := []string{}
	if ha {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", filepath.Join(ansibleDir, "playbooks", "k0s-kubeconfig-k0s-controller")))
	}

	return []commandStep{
		{
			description: "ensure flux namespace exists",
			dir:         themistoklesDir,
			env:         env,
			name:        "/bin/sh",
			args:        []string{"-c", fmt.Sprintf(`kubectl get ns "%s" >/dev/null 2>&1 || kubectl create ns "%s"`, defaultFluxNS, defaultFluxNS)},
		},
		{
			description: "create or update SOPS GPG secret",
			dir:         themistoklesDir,
			env:         env,
			name:        "/bin/sh",
			args: []string{
				"-c",
				fmt.Sprintf(`gpg --batch --yes --pinentry-mode loopback --export-secret-keys --armor "%s" | kubectl -n "%s" create secret generic "%s" --from-file=sops.asc=/dev/stdin --dry-run=client -o yaml | kubectl apply -f -`, defaultSOPSKeyFP, defaultFluxNS, defaultSOPSSecret),
			},
		},
		{
			description: "create Flux git source",
			dir:         themistoklesDir,
			env:         env,
			name:        "/bin/sh",
			args: []string{
				"-c",
				fmt.Sprintf(`flux create source git %s --url="%s" --branch="%s" --interval=1m --export | kubectl apply -f -`, defaultFluxSource, defaultFluxRepoURL, defaultFluxBranch),
			},
		},
		{
			description: "create Flux kustomization",
			dir:         themistoklesDir,
			env:         env,
			name:        "/bin/sh",
			args: []string{
				"-c",
				fmt.Sprintf(`flux create kustomization %s --namespace="%s" --source="%s" --path="%s" --prune=true --interval=10m --export | kubectl apply -f -`, defaultFluxSource, defaultFluxNS, defaultFluxSource, pathInRepo),
			},
		},
	}
}

func limaDeleteSteps(ha bool, limaDir string) []commandStep {
	if ha {
		return []commandStep{
			{
				description: "stop Lima HA cluster",
				dir:         limaDir,
				name:        "limactl",
				args:        []string{"stop", "k0s-controller", "k0s-worker1", "k0s-worker2"},
			},
			{
				description: "delete Lima HA VMs",
				dir:         limaDir,
				name:        "limactl",
				args:        []string{"delete", "k0s-controller", "k0s-worker1", "k0s-worker2"},
			},
			{
				description: "delete Lima HA disks",
				dir:         limaDir,
				name:        "/bin/sh",
				args:        []string{"-c", "limactl disk delete pyxis-controller pyxis-worker1 pyxis-worker2"},
			},
		}
	}

	return []commandStep{
		{
			description: "stop Lima single-node cluster",
			dir:         limaDir,
			name:        "limactl",
			args:        []string{"stop", "k0s-byzantium"},
		},
		{
			description: "delete Lima single-node VM",
			dir:         limaDir,
			name:        "limactl",
			args:        []string{"delete", "k0s-byzantium"},
		},
		{
			description: "delete Lima single-node disk",
			dir:         limaDir,
			name:        "limactl",
			args:        []string{"disk", "delete", "pyxis"},
		},
	}
}

func ensureLimaDisks(ha bool, limaDir string) error {
	disks := []string{"pyxis"}
	if ha {
		disks = []string{"pyxis-controller", "pyxis-worker1", "pyxis-worker2"}
	}

	output, err := runCommandCapture(limaDir, nil, "limactl", "disk", "ls")
	if err != nil {
		return err
	}

	for _, disk := range disks {
		if strings.Contains(output, disk) {
			logging.Info(fmt.Sprintf("Lima disk %s already exists", disk))
			continue
		}

		if err := runStep(commandStep{
			description: "create Lima disk " + disk,
			dir:         limaDir,
			name:        "limactl",
			args:        []string{"disk", "create", disk, "--size", "30G"},
		}); err != nil {
			return err
		}
	}

	return nil
}

func runStep(step commandStep) error {
	logging.System(step.description)
	return runCommand(step.dir, step.env, step.name, step.args...)
}

func runStepAllowFailure(step commandStep) error {
	logging.System(step.description)
	err := runCommand(step.dir, step.env, step.name, step.args...)
	if err != nil {
		logging.Info(fmt.Sprintf("Ignoring failure for %q: %v", step.description, err))
	}

	return nil
}

func runCommand(dir string, extraEnv []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), extraEnv...)

	return cmd.Run()
}

func runCommandCapture(dir string, extraEnv []string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), extraEnv...)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
