package parse

import (
	"github.com/odysseia-greek/mykenai/archimedes/command/parse/command"
	"github.com/spf13/cobra"
)

func Manager() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parse",
		Short: "parse words",
		Long:  `Allows operations to be done on word lists`,
	}

	cmd.AddCommand(
		command.ReparseList(),
		command.ListToWords(),
		command.SetImageVersion(),
		command.GroupChapters(),
		command.CreateOneFromMany(),
	)

	return cmd
}
