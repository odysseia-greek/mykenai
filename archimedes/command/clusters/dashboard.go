package clusters

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

//go:embed templates/dashboard.html
var dashboardTemplates embed.FS

const (
	defaultDashboardAddress  = "127.0.0.1:8181"
	defaultDashboardInterval = 10 * time.Second
)

type dashboardOptions struct {
	statusOptions
	address  string
	interval time.Duration
}

type dashboardPage struct {
	APIPath    string
	Cluster    string
	Timeout    string
	IntervalMS int64
}

type dashboardView struct {
	GeneratedAt  string
	Cluster      string
	Timeout      string
	TotalNodes   int
	PingOK       int
	SSHOK        int
	ProblemCount int
	Clusters     []clusterView
}

type clusterView struct {
	Name         string
	TotalNodes   int
	HealthyNodes int
	Nodes        []nodeView
}

type nodeView struct {
	Name      string
	Address   string
	Arch      string
	Role      string
	PingClass string
	PingLabel string
	SSHClass  string
	SSHLabel  string
	Load      string
	Memory    string
	Disks     string
	Error     string
}

func Dashboard() *cobra.Command {
	opts := &dashboardOptions{
		statusOptions: statusOptions{
			cluster: defaultClusterName,
			timeout: defaultTimeout,
			ssh:     true,
			ping:    true,
		},
		address:  defaultDashboardAddress,
		interval: defaultDashboardInterval,
	}

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "run a live HTML cluster status dashboard",
		Long:  "Run a local HTTP dashboard using an embedded template. The browser polls cluster status at a fixed interval.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDashboard(cmd, opts)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.cluster, "cluster", "c", defaultClusterName, "cluster to inspect: hellas, hellenistike, or all")
	cmd.PersistentFlags().DurationVar(&opts.timeout, "timeout", defaultTimeout, "timeout per node probe")
	cmd.PersistentFlags().BoolVar(&opts.ssh, "ssh", true, "collect load, memory, and disk layout over SSH")
	cmd.PersistentFlags().BoolVar(&opts.ping, "ping", true, "run ICMP ping checks")
	cmd.PersistentFlags().StringVar(&opts.user, "user", "", "override SSH user for all nodes")
	cmd.PersistentFlags().StringVarP(&opts.identity, "identity", "i", "", "override SSH private key for all nodes")
	cmd.PersistentFlags().StringVar(&opts.address, "addr", defaultDashboardAddress, "HTTP listen address")
	cmd.PersistentFlags().DurationVar(&opts.interval, "interval", defaultDashboardInterval, "browser refresh interval")

	return cmd
}

func runDashboard(cmd *cobra.Command, opts *dashboardOptions) error {
	selected := selectClusters(opts.cluster)
	if len(selected) == 0 {
		return fmt.Errorf("unknown cluster %q; expected hellas, hellenistike, or all", opts.cluster)
	}

	tmpl, err := template.ParseFS(dashboardTemplates, "templates/dashboard.html")
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page := dashboardPage{
			APIPath:    "/api/status",
			Cluster:    opts.cluster,
			Timeout:    opts.timeout.String(),
			IntervalMS: opts.interval.Milliseconds(),
		}
		if err := tmpl.Execute(w, page); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		results := probeClusters(selected, &opts.statusOptions)
		sortNodeStatuses(results)
		view := buildDashboardView(results, opts)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(view); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	fmt.Fprintf(cmd.OutOrStdout(), "dashboard listening on http://%s\n", opts.address)
	fmt.Fprintf(cmd.OutOrStdout(), "refresh interval: %s\n", opts.interval)
	return http.ListenAndServe(opts.address, mux)
}

func buildDashboardView(results []nodeStatus, opts *dashboardOptions) dashboardView {
	view := dashboardView{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Cluster:     opts.cluster,
		Timeout:     opts.timeout.String(),
		TotalNodes:  len(results),
	}

	clusterIndex := map[string]int{}
	for _, result := range results {
		if result.PingOK {
			view.PingOK++
		}
		if result.SSHOK {
			view.SSHOK++
		}
		if !nodeFullyHealthy(result) {
			view.ProblemCount++
		}

		index, ok := clusterIndex[result.Cluster]
		if !ok {
			view.Clusters = append(view.Clusters, clusterView{Name: result.Cluster})
			index = len(view.Clusters) - 1
			clusterIndex[result.Cluster] = index
		}

		node := nodeView{
			Name:      result.Node.Name,
			Address:   result.Node.Address,
			Arch:      result.Node.Arch,
			Role:      result.Node.Role,
			PingClass: healthClass(result.PingOK, result.PingError),
			PingLabel: healthLabel(result.PingOK, result.PingError),
			SSHClass:  healthClass(result.SSHOK, result.SSHError),
			SSHLabel:  healthLabel(result.SSHOK, result.SSHError),
			Load:      defaultValue(result.Load, "-"),
			Memory:    defaultValue(result.Memory, "-"),
			Disks:     defaultValue(result.Disks, "-"),
			Error:     combinedError(result),
		}

		view.Clusters[index].Nodes = append(view.Clusters[index].Nodes, node)
		view.Clusters[index].TotalNodes++
		if nodeFullyHealthy(result) {
			view.Clusters[index].HealthyNodes++
		}
	}

	return view
}

func sortNodeStatuses(results []nodeStatus) {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Cluster != results[j].Cluster {
			return results[i].Cluster < results[j].Cluster
		}
		return results[i].Node.Name < results[j].Node.Name
	})
}

func nodeFullyHealthy(result nodeStatus) bool {
	return result.PingOK && result.SSHOK
}

func healthClass(ok bool, err string) string {
	if ok {
		return "ok"
	}
	if err == "" {
		return "unknown"
	}
	return "bad"
}

func healthLabel(ok bool, err string) string {
	if ok {
		return "ok"
	}
	if err == "" {
		return "skipped"
	}
	return "failed"
}

func combinedError(result nodeStatus) string {
	switch {
	case result.PingError != "" && result.SSHError != "":
		return "ping: " + result.PingError + " · ssh: " + result.SSHError
	case result.PingError != "":
		return "ping: " + result.PingError
	case result.SSHError != "":
		return "ssh: " + result.SSHError
	default:
		return ""
	}
}
