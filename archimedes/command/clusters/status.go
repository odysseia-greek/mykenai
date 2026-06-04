package clusters

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

const (
	defaultClusterName = "all"
	defaultTimeout     = 3 * time.Second
)

type statusOptions struct {
	cluster  string
	timeout  time.Duration
	ssh      bool
	ping     bool
	user     string
	identity string
}

type nodeStatus struct {
	Cluster   string
	Node      node
	PingOK    bool
	PingError string
	SSHOK     bool
	SSHError  string
	Load      string
	Memory    string
	Disks     string
}

func Status() *cobra.Command {
	opts := &statusOptions{
		cluster: defaultClusterName,
		timeout: defaultTimeout,
		ssh:     true,
		ping:    true,
	}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "show cluster node liveness and basic machine health",
		Long: `Show a compact cluster dashboard using ICMP and SSH probes.

The command starts with static inventory data copied from lykourgos/ansible/inventories.
It checks whether nodes answer ping, whether SSH is available, and collects load,
memory, and block-device layout over SSH.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClusterStatus(cmd, opts)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.cluster, "cluster", "c", defaultClusterName, "cluster to inspect: hellas, hellenistike, or all")
	cmd.PersistentFlags().DurationVar(&opts.timeout, "timeout", defaultTimeout, "timeout per node probe")
	cmd.PersistentFlags().BoolVar(&opts.ssh, "ssh", true, "collect load and memory over SSH")
	cmd.PersistentFlags().BoolVar(&opts.ping, "ping", true, "run ICMP ping checks")
	cmd.PersistentFlags().StringVar(&opts.user, "user", "", "override SSH user for all nodes")
	cmd.PersistentFlags().StringVarP(&opts.identity, "identity", "i", "", "override SSH private key for all nodes")

	return cmd
}

func runClusterStatus(cmd *cobra.Command, opts *statusOptions) error {
	selected := selectClusters(opts.cluster)
	if len(selected) == 0 {
		return fmt.Errorf("unknown cluster %q; expected hellas, hellenistike, or all", opts.cluster)
	}

	results := probeClusters(selected, opts)
	sort.Slice(results, func(i, j int) bool {
		if results[i].Cluster != results[j].Cluster {
			return results[i].Cluster < results[j].Cluster
		}
		return results[i].Node.Name < results[j].Node.Name
	})

	renderStatus(cmd.OutOrStdout(), results, opts)
	return nil
}

func probeClusters(clusters []cluster, opts *statusOptions) []nodeStatus {
	var wg sync.WaitGroup
	statuses := make(chan nodeStatus)

	for _, selected := range clusters {
		for _, selectedNode := range selected.Nodes {
			wg.Add(1)
			go func(clusterName string, n node) {
				defer wg.Done()
				statuses <- probeNode(clusterName, n, opts)
			}(selected.Name, selectedNode)
		}
	}

	go func() {
		wg.Wait()
		close(statuses)
	}()

	var results []nodeStatus
	for status := range statuses {
		results = append(results, status)
	}

	return results
}

func probeNode(clusterName string, n node, opts *statusOptions) nodeStatus {
	if opts.user != "" {
		n.User = opts.user
	}
	if opts.identity != "" {
		n.Identity = opts.identity
	}

	result := nodeStatus{
		Cluster: clusterName,
		Node:    n,
	}

	if opts.ping {
		output, err := runWithTimeout(opts.timeout, "ping", "-c", "1", n.Address)
		if err != nil {
			result.PingError = compactError(err, output)
		} else {
			result.PingOK = true
		}
	}

	if opts.ssh {
		args := sshArgs(n, remoteStatusScript())
		output, err := runWithTimeout(opts.timeout, args[0], args[1:]...)
		if err != nil {
			result.SSHError = compactError(err, output)
		} else {
			result.SSHOK = true
			parseRemoteStatus(output, &result)
		}
	}

	return result
}

func sshArgs(n node, remoteCommand string) []string {
	args := []string{
		"ssh",
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=3",
	}

	if n.Identity != "" {
		args = append(args, "-i", expandHome(n.Identity))
	}

	args = append(args, fmt.Sprintf("%s@%s", n.User, n.Address), remoteCommand)
	return args
}

func remoteStatusScript() string {
	return strings.Join([]string{
		`printf 'load=%s\n' "$(cut -d ' ' -f1-3 /proc/loadavg)"`,
		`awk '/MemTotal/{total=$2}/MemAvailable/{avail=$2} END{used=total-avail; printf "memory=%d/%dMiB %.1f%%\n", used/1024, total/1024, used*100/total}' /proc/meminfo`,
		`printf 'disks=%s\n' "$(lsblk -ndo NAME,SIZE,TYPE,MOUNTPOINTS | awk '$3 == "disk" { printf "%s:%s", $1, $2; if ($4 != "") printf "(%s)", $4; printf " " }' | sed 's/ $//')"`,
	}, "; ")
}

func parseRemoteStatus(output string, status *nodeStatus) {
	for _, line := range strings.Split(output, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}
		switch key {
		case "load":
			status.Load = value
		case "memory":
			status.Memory = value
		case "disks":
			status.Disks = value
		}
	}
}

func runWithTimeout(timeout time.Duration, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := strings.TrimSpace(strings.Join([]string{stdout.String(), stderr.String()}, "\n"))
	if ctx.Err() == context.DeadlineExceeded {
		return output, ctx.Err()
	}

	return output, err
}

func renderStatus(out interface{ Write([]byte) (int, error) }, results []nodeStatus, opts *statusOptions) {
	fmt.Fprintln(out, "CLUSTER NODE STATUS")
	fmt.Fprintln(out, "===================")
	fmt.Fprintf(out, "cluster=%s timeout=%s ping=%t ssh=%t\n\n", opts.cluster, opts.timeout, opts.ping, opts.ssh)

	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "CLUSTER\tNODE\tROLE\tADDR\tPING\tSSH\tLOAD\tMEMORY\tDISKS")
	for _, result := range results {
		fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			result.Cluster,
			result.Node.Name,
			result.Node.Role,
			result.Node.Address,
			statusCell(result.PingOK, result.PingError),
			statusCell(result.SSHOK, result.SSHError),
			defaultValue(result.Load, "-"),
			defaultValue(result.Memory, "-"),
			defaultValue(result.Disks, "-"),
		)
	}
	_ = tw.Flush()
}

func statusCell(ok bool, err string) string {
	if ok {
		return "ok"
	}
	if err == "" {
		return "-"
	}
	return "fail: " + err
}

func compactError(err error, output string) string {
	parts := []string{err.Error()}
	if strings.TrimSpace(output) != "" {
		parts = append(parts, strings.TrimSpace(output))
	}
	joined := strings.Join(parts, ": ")
	joined = strings.Join(strings.Fields(joined), " ")
	if len(joined) > 90 {
		return joined[:87] + "..."
	}
	return joined
}

func expandHome(path string) string {
	if path == "" || path[0] != '~' {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}

	return path
}

func defaultValue(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
