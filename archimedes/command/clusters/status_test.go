package clusters

import (
	"strings"
	"testing"
)

func TestSelectClusters(t *testing.T) {
	selected := selectClusters("hellas")
	if len(selected) != 1 {
		t.Fatalf("expected one cluster, got %d", len(selected))
	}
	if selected[0].Name != "hellas" {
		t.Fatalf("expected hellas, got %s", selected[0].Name)
	}
	if len(selected[0].Nodes) != 4 {
		t.Fatalf("expected 4 hellas nodes, got %d", len(selected[0].Nodes))
	}

	all := selectClusters("all")
	if len(all) != 2 {
		t.Fatalf("expected two clusters, got %d", len(all))
	}

	if unknown := selectClusters("unknown"); len(unknown) != 0 {
		t.Fatalf("expected unknown cluster to return no results, got %d", len(unknown))
	}
}

func TestParseRemoteStatus(t *testing.T) {
	status := nodeStatus{}
	parseRemoteStatus("load=0.12 0.25 0.31\nmemory=512/8192MiB 6.2%\ndisks=mmcblk0:29.7G(/boot/firmware) sda:465.8G", &status)

	if status.Load != "0.12 0.25 0.31" {
		t.Fatalf("unexpected load: %s", status.Load)
	}
	if status.Memory != "512/8192MiB 6.2%" {
		t.Fatalf("unexpected memory: %s", status.Memory)
	}
	if !strings.Contains(status.Disks, "mmcblk0:29.7G") {
		t.Fatalf("unexpected disks: %s", status.Disks)
	}
}

func TestExpandHome(t *testing.T) {
	expanded := expandHome("~/.ssh/id_raspie")
	if strings.HasPrefix(expanded, "~") {
		t.Fatalf("expected home expansion, got %s", expanded)
	}
	if !strings.HasSuffix(expanded, ".ssh/id_raspie") {
		t.Fatalf("unexpected expanded path: %s", expanded)
	}
}
