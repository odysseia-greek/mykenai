package chart

import (
	"github.com/odysseia-greek/mykenai/archimedes/command/chart/command"
	"github.com/spf13/cobra"
)

func Manager() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chart",
		Short: "setup charts",
		Long:  `Allows you to parse the helm chart for images`,
	}

	cmd.AddCommand(
		command.Parse(),
	)

	return cmd
}
