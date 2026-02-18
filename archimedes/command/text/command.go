package text

import (
	"github.com/odysseia-greek/mykenai/archimedes/command/text/command"
	"github.com/spf13/cobra"
)

func Manager() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "text",
		Short: "Get texts and parse them",
		Long:  `Crawls texts from the internet and parse them into rhema.json`,
	}

	cmd.AddCommand(
		command.Crawl(),
		command.Fetch(),
	)

	return cmd
}
