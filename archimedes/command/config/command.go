package config

import (
	"github.com/odysseia-greek/mykenai/archimedes/command/config/command"
	"github.com/spf13/cobra"
)

func Manager() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "setup archimedes",
		Long:  `Allows you to set paths so archimedes knows there to find code`,
	}

	cmd.AddCommand(
		command.Set(),
	)

	return cmd
}
