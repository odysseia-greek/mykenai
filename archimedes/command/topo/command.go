package topo

import "github.com/spf13/cobra"

func Manager() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topo",
		Short: "Inspect TopoLVM storage relationships",
		Long:  "Join Kubernetes PVCs, PVs, StorageClasses, and TopoLVM LogicalVolumes using the active kubectl context.",
	}

	cmd.AddCommand(List(), Get())
	return cmd
}
