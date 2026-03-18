package odysseia

import (
	"github.com/odysseia-greek/mykenai/archimedes/command/odysseia/command"
	"github.com/spf13/cobra"
)

func Manager() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "odysseia",
		Short: "meta odysseia commands",
		Long:  `Manage odysseia local cluster workflows and supporting commands.`,
	}

	cmd.AddCommand(
		command.Create(),
		command.Delete(),
		command.Restart(),
		command.Status(),
		command.GenerateDocs(),
		command.Tidy(),
	)

	return cmd
}
