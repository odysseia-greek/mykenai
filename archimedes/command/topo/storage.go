package topo

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var logicalVolumeGVR = schema.GroupVersionResource{Group: "topolvm.io", Version: "v1", Resource: "logicalvolumes"}

type listOptions struct {
	namespace string
}

type storageRow struct {
	Namespace     string
	Claim         string
	PV            string
	StorageClass  string
	Reclaim       string
	Size          string
	VolumeHandle  string
	LogicalVolume string
	Node          string
	DeviceClass   string
	Status        string
	Finalizers    []string
}

type storageInventory struct {
	Rows []storageRow
}

func List() *cobra.Command {
	opts := &listOptions{}
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List PVC, PV, and TopoLVM LogicalVolume relationships",
		RunE: func(cmd *cobra.Command, args []string) error {
			inventory, err := loadStorageInventory()
			if err != nil {
				return err
			}
			rows := filterNamespace(inventory.Rows, opts.namespace)
			renderRows(cmd, rows)
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.namespace, "namespace", "n", "", "show claims in one namespace")
	cmd.Flags().BoolP("all-namespaces", "A", false, "show claims in all namespaces (default when no namespace is set)")
	return cmd
}

func Get() *cobra.Command {
	opts := &listOptions{}
	cmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Show one PVC, PV, volume handle, or LogicalVolume relationship",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inventory, err := loadStorageInventory()
			if err != nil {
				return err
			}
			rows := findRows(inventory.Rows, args[0], opts.namespace)
			if len(rows) == 0 {
				return fmt.Errorf("no PVC, PV, volume handle, or TopoLVM LogicalVolume found for %q", args[0])
			}
			renderDetails(cmd, rows)
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.namespace, "namespace", "n", "", "limit PVC matches to one namespace")
	return cmd
}

func loadStorageInventory() (storageInventory, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return storageInventory{}, fmt.Errorf("load Kubernetes configuration: %w", err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return storageInventory{}, fmt.Errorf("create Kubernetes client: %w", err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return storageInventory{}, fmt.Errorf("create Kubernetes dynamic client: %w", err)
	}

	ctx := context.Background()
	pvcs, err := client.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return storageInventory{}, fmt.Errorf("list PersistentVolumeClaims: %w", err)
	}
	pvs, err := client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return storageInventory{}, fmt.Errorf("list PersistentVolumes: %w", err)
	}
	storageClasses, err := client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return storageInventory{}, fmt.Errorf("list StorageClasses: %w", err)
	}
	logicalVolumes, err := dynamicClient.Resource(logicalVolumeGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return storageInventory{}, fmt.Errorf("list TopoLVM LogicalVolumes (is TopoLVM installed?): %w", err)
	}

	return buildInventory(
		toUnstructuredSlice(pvcs.Items),
		toUnstructuredSlice(pvs.Items),
		toUnstructuredSlice(storageClasses.Items),
		unstructuredItems(logicalVolumes.Items),
	), nil
}

func toUnstructuredSlice[T any](items []T) []map[string]any {
	objects := make([]map[string]any, 0, len(items))
	for i := range items {
		object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&items[i])
		if err == nil {
			objects = append(objects, object)
		}
	}
	return objects
}

func unstructuredItems(items []unstructured.Unstructured) []map[string]any {
	objects := make([]map[string]any, 0, len(items))
	for _, item := range items {
		objects = append(objects, item.UnstructuredContent())
	}
	return objects
}

func buildInventory(pvcs, pvs, storageClasses, logicalVolumes []map[string]any) storageInventory {
	pvcByKey := map[string]map[string]any{}
	for _, pvc := range pvcs {
		pvcByKey[objectNamespace(pvc)+"/"+objectName(pvc)] = pvc
	}
	scProvisioner := map[string]string{}
	scDeviceClass := map[string]string{}
	for _, sc := range storageClasses {
		scProvisioner[objectName(sc)] = nestedString(sc, "provisioner")
		scDeviceClass[objectName(sc)] = nestedString(sc, "parameters", "topolvm.cybozu.com/device-class")
	}
	lvByName := map[string]map[string]any{}
	for _, lv := range logicalVolumes {
		lvByName[objectName(lv)] = lv
		for _, alias := range []string{nestedString(lv, "spec", "name"), nestedString(lv, "spec", "volumeID")} {
			if alias != "" {
				lvByName[alias] = lv
			}
		}
	}

	rows := make([]storageRow, 0, len(pvs)+len(pvcs)+len(logicalVolumes))
	seenPVC, seenLV := map[string]bool{}, map[string]bool{}
	for _, pv := range pvs {
		ns, claim := nestedString(pv, "spec", "claimRef", "namespace"), nestedString(pv, "spec", "claimRef", "name")
		pvcKey := ns + "/" + claim
		handle := nestedString(pv, "spec", "csi", "volumeHandle")
		lv := lvByName[objectName(pv)]
		if lv == nil {
			lv = lvByName[handle]
		}
		storageClass := first(nestedString(pv, "spec", "storageClassName"), nestedString(pvcByKey[pvcKey], "spec", "storageClassName"))
		if !strings.Contains(scProvisioner[storageClass], "topolvm") {
			continue
		}
		row := storageRow{Namespace: ns, Claim: claim, PV: objectName(pv), StorageClass: storageClass, Reclaim: nestedString(pv, "spec", "persistentVolumeReclaimPolicy"), Size: nestedString(pv, "spec", "capacity", "storage"), VolumeHandle: handle, Status: nestedString(pv, "status", "phase"), Finalizers: stringSlice(nested(pv, "metadata", "finalizers"))}
		if claim != "" {
			seenPVC[pvcKey] = true
		}
		if lv != nil {
			row.LogicalVolume, row.Node, row.DeviceClass = objectName(lv), nestedString(lv, "spec", "nodeName"), nestedString(lv, "spec", "deviceClass")
			if row.DeviceClass == "" {
				row.DeviceClass = scDeviceClass[storageClass]
			}
			if row.DeviceClass == "" {
				row.DeviceClass = "default"
			}
			seenLV[objectName(lv)] = true
		} else {
			row.Status = appendStatus(row.Status, "missing-lv")
		}
		if claim != "" && pvcByKey[pvcKey] == nil {
			row.Status = appendStatus(row.Status, "missing-pvc")
		}
		rows = append(rows, row)
	}
	for key, pvc := range pvcByKey {
		if seenPVC[key] {
			continue
		}
		if !strings.Contains(scProvisioner[nestedString(pvc, "spec", "storageClassName")], "topolvm") {
			continue
		}
		rows = append(rows, storageRow{Namespace: objectNamespace(pvc), Claim: objectName(pvc), PV: nestedString(pvc, "spec", "volumeName"), StorageClass: nestedString(pvc, "spec", "storageClassName"), Size: nestedString(pvc, "spec", "resources", "requests", "storage"), Status: appendStatus(nestedString(pvc, "status", "phase"), "missing-pv"), Finalizers: stringSlice(nested(pvc, "metadata", "finalizers"))})
	}
	for _, lv := range logicalVolumes {
		if seenLV[objectName(lv)] {
			continue
		}
		rows = append(rows, storageRow{LogicalVolume: objectName(lv), VolumeHandle: objectName(lv), Node: nestedString(lv, "spec", "nodeName"), DeviceClass: nestedString(lv, "spec", "deviceClass"), Size: formatSize(nested(lv, "spec", "size")), Status: "orphan-lv", Finalizers: stringSlice(nested(lv, "metadata", "finalizers"))})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Namespace+"/"+rows[i].Claim+rows[i].PV < rows[j].Namespace+"/"+rows[j].Claim+rows[j].PV
	})
	return storageInventory{Rows: rows}
}

func renderRows(cmd *cobra.Command, rows []storageRow) {
	tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAMESPACE\tPVC\tPV\tLOGICALVOLUME\tNODE\tCLASS\tSIZE\tRECLAIM\tSTATUS")
	for _, row := range rows {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", dash(row.Namespace), dash(row.Claim), dash(row.PV), dash(row.LogicalVolume), dash(row.Node), dash(row.DeviceClass), dash(row.Size), dash(row.Reclaim), dash(row.Status))
	}
	_ = tw.Flush()
}

func renderDetails(cmd *cobra.Command, rows []storageRow) {
	for i, row := range rows {
		if i > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "---")
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Namespace: %s\nPVC: %s\nPV: %s\nStorageClass: %s\nReclaimPolicy: %s\nSize: %s\nVolumeHandle: %s\nLogicalVolume: %s\nNode: %s\nDeviceClass: %s\nStatus: %s\nFinalizers: %s\n", dash(row.Namespace), dash(row.Claim), dash(row.PV), dash(row.StorageClass), dash(row.Reclaim), dash(row.Size), dash(row.VolumeHandle), dash(row.LogicalVolume), dash(row.Node), dash(row.DeviceClass), dash(row.Status), dash(strings.Join(row.Finalizers, ",")))
	}
}

func filterNamespace(rows []storageRow, namespace string) []storageRow {
	if namespace == "" {
		return rows
	}
	filtered := []storageRow{}
	for _, row := range rows {
		if row.Namespace == namespace {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func findRows(rows []storageRow, name, namespace string) []storageRow {
	found := []storageRow{}
	for _, row := range rows {
		if namespace != "" && row.Namespace != "" && row.Namespace != namespace {
			continue
		}
		if row.Claim == name || row.PV == name || row.VolumeHandle == name || row.LogicalVolume == name {
			found = append(found, row)
		}
	}
	return found
}

func objectName(object map[string]any) string { return nestedString(object, "metadata", "name") }
func objectNamespace(object map[string]any) string {
	return nestedString(object, "metadata", "namespace")
}
func nested(object map[string]any, path ...string) any {
	var value any = object
	for _, key := range path {
		current, ok := value.(map[string]any)
		if !ok {
			return nil
		}
		value = current[key]
	}
	return value
}
func nestedString(object map[string]any, path ...string) string {
	value := nested(object, path...)
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprint(value)
}
func stringSlice(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, fmt.Sprint(item))
	}
	return result
}
func first(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
func dash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
func appendStatus(status, issue string) string {
	if status == "" {
		return issue
	}
	return status + "," + issue
}
func formatSize(value any) string {
	switch size := value.(type) {
	case float64:
		return fmt.Sprintf("%.0fB", size)
	case string:
		return size
	default:
		if size == nil {
			return ""
		}
		return fmt.Sprint(size)
	}
}
