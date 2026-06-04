package clusters

import "github.com/spf13/cobra"

func Manager() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clusters",
		Short: "Inspect Odysseia cluster nodes",
		Long:  "Inspect Odysseia cluster nodes using lightweight network and SSH probes.",
	}

	cmd.AddCommand(Status())
	cmd.AddCommand(Dashboard())

	return cmd
}
