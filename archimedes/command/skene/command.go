package skene

import (
	"github.com/odysseia-greek/mykenai/archimedes/command/skene/command"
	"github.com/spf13/cobra"
)

func Manager() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skene",
		Short: "Create new scaffolding for an api or job",
		Long:  `Allows the creation of an api or a job from a template`,
	}

	cmd.AddCommand(
		command.CreateApi(),
		command.CreateJob(),
	)

	return cmd
}
