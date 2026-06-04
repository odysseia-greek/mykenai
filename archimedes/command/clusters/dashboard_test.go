package clusters

import (
	"html/template"
	"strings"
	"testing"
)

func TestBuildDashboardView(t *testing.T) {
	opts := &dashboardOptions{
		statusOptions: statusOptions{
			cluster: "hellas",
			timeout: defaultTimeout,
			ssh:     true,
			ping:    true,
		},
		address:  defaultDashboardAddress,
		interval: defaultDashboardInterval,
	}

	view := buildDashboardView([]nodeStatus{
		{
			Cluster: "hellas",
			Node:    node{Name: "sparta.hellas", Address: "192.168.1.121", Arch: "rpi5", Role: "controller"},
			PingOK:  true,
			SSHOK:   true,
			Load:    "0.1 0.2 0.3",
			Memory:  "1024/4096MiB 25.0%",
			Disks:   "nvme0n1:931.5G",
		},
		{
			Cluster:   "hellas",
			Node:      node{Name: "thebai.hellas", Address: "192.168.1.123", Arch: "rpi5", Role: "worker"},
			PingOK:    true,
			SSHOK:     false,
			SSHError:  "permission denied",
			PingError: "",
		},
	}, opts)

	if view.TotalNodes != 2 {
		t.Fatalf("expected two nodes, got %d", view.TotalNodes)
	}
	if view.PingOK != 2 {
		t.Fatalf("expected two ping-ok nodes, got %d", view.PingOK)
	}
	if view.SSHOK != 1 {
		t.Fatalf("expected one ssh-ok node, got %d", view.SSHOK)
	}
	if view.ProblemCount != 1 {
		t.Fatalf("expected one problem, got %d", view.ProblemCount)
	}
	if len(view.Clusters) != 1 || view.Clusters[0].HealthyNodes != 1 {
		t.Fatalf("unexpected cluster health: %+v", view.Clusters)
	}
	if view.Clusters[0].Nodes[1].SSHClass != "bad" {
		t.Fatalf("expected failed ssh class, got %s", view.Clusters[0].Nodes[1].SSHClass)
	}
}

func TestDashboardTemplateRendersLiveShell(t *testing.T) {
	tmpl, err := template.ParseFS(dashboardTemplates, "templates/dashboard.html")
	if err != nil {
		t.Fatal(err)
	}

	var rendered strings.Builder
	err = tmpl.Execute(&rendered, dashboardPage{
		APIPath:    "/api/status",
		Cluster:    "all",
		Timeout:    defaultTimeout.String(),
		IntervalMS: defaultDashboardInterval.Milliseconds(),
	})
	if err != nil {
		t.Fatal(err)
	}

	output := rendered.String()
	for _, expected := range []string{"const apiPath", "fetch(apiPath", "setInterval(refresh, intervalMS)"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected rendered dashboard to contain %q", expected)
		}
	}
}
