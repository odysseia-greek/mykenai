package topo

import (
	"strings"
	"testing"
)

func TestBuildInventoryJoinsPVCToLogicalVolume(t *testing.T) {
	pvcs := []map[string]any{{"metadata": map[string]any{"name": "vault-data", "namespace": "delphi"}, "spec": map[string]any{"volumeName": "pvc-123", "storageClassName": "pyxis-delete"}, "status": map[string]any{"phase": "Bound"}}}
	pvs := []map[string]any{{"metadata": map[string]any{"name": "pvc-123"}, "spec": map[string]any{"claimRef": map[string]any{"name": "vault-data", "namespace": "delphi"}, "storageClassName": "pyxis-delete", "persistentVolumeReclaimPolicy": "Delete", "capacity": map[string]any{"storage": "30Gi"}, "csi": map[string]any{"volumeHandle": "different-csi-handle"}}, "status": map[string]any{"phase": "Bound"}}}
	classes := []map[string]any{{"metadata": map[string]any{"name": "pyxis-delete"}, "provisioner": "topolvm.io", "parameters": map[string]any{"topolvm.cybozu.com/device-class": "pyxis"}}}
	lvs := []map[string]any{{"metadata": map[string]any{"name": "pvc-123"}, "spec": map[string]any{"nodeName": "thebai-hellas", "size": float64(32212254720)}}}

	rows := buildInventory(pvcs, pvs, classes, lvs).Rows
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	row := rows[0]
	if row.Namespace != "delphi" || row.Claim != "vault-data" || row.PV != "pvc-123" {
		t.Fatalf("unexpected claim chain: %+v", row)
	}
	if row.LogicalVolume != "pvc-123" || row.Node != "thebai-hellas" || row.DeviceClass != "pyxis" {
		t.Fatalf("unexpected TopoLVM chain: %+v", row)
	}
	if row.Status != "Bound" {
		t.Fatalf("unexpected status: %s", row.Status)
	}
}

func TestBuildInventoryReportsBrokenAndOrphanedLinks(t *testing.T) {
	pvs := []map[string]any{{"metadata": map[string]any{"name": "missing-links"}, "spec": map[string]any{"claimRef": map[string]any{"name": "gone", "namespace": "delphi"}, "storageClassName": "pyxis-delete", "csi": map[string]any{"volumeHandle": "gone-lv"}}, "status": map[string]any{"phase": "Released"}}}
	classes := []map[string]any{{"metadata": map[string]any{"name": "pyxis-delete"}, "provisioner": "topolvm.io"}}
	lvs := []map[string]any{{"metadata": map[string]any{"name": "orphan-lv"}, "spec": map[string]any{"nodeName": "athenai-hellas"}}}

	rows := buildInventory(nil, pvs, classes, lvs).Rows
	if len(rows) != 2 {
		t.Fatalf("expected two rows, got %d", len(rows))
	}
	var statuses string
	for _, row := range rows {
		statuses += row.Status + " "
	}
	for _, expected := range []string{"missing-lv", "missing-pvc", "orphan-lv"} {
		if !strings.Contains(statuses, expected) {
			t.Fatalf("expected %q in statuses %q", expected, statuses)
		}
	}
}

func TestFindRowsSupportsEveryIdentifier(t *testing.T) {
	rows := []storageRow{{Namespace: "delphi", Claim: "vault", PV: "pv-1", VolumeHandle: "handle-1", LogicalVolume: "lv-1"}}
	for _, name := range []string{"vault", "pv-1", "handle-1", "lv-1"} {
		if len(findRows(rows, name, "delphi")) != 1 {
			t.Fatalf("expected match for %s", name)
		}
	}
	if len(findRows(rows, "vault", "attike")) != 0 {
		t.Fatal("unexpected cross-namespace match")
	}
}

func TestBuildInventoryExcludesOtherCSIStorage(t *testing.T) {
	pvcs := []map[string]any{{"metadata": map[string]any{"name": "other", "namespace": "default"}, "spec": map[string]any{"storageClassName": "other-csi"}}}
	pvs := []map[string]any{{"metadata": map[string]any{"name": "other-pv"}, "spec": map[string]any{"storageClassName": "other-csi", "csi": map[string]any{"volumeHandle": "other-handle"}}}}
	classes := []map[string]any{{"metadata": map[string]any{"name": "other-csi"}, "provisioner": "example.csi.io"}}
	if rows := buildInventory(pvcs, pvs, classes, nil).Rows; len(rows) != 0 {
		t.Fatalf("expected non-TopoLVM storage to be excluded, got %+v", rows)
	}
}
