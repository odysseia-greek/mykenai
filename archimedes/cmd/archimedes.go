package main

import (
	"strings"

	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/odysseia-greek/mykenai/archimedes/command/images"
	"github.com/odysseia-greek/mykenai/archimedes/command/odysseia"
	"github.com/odysseia-greek/mykenai/archimedes/command/parse"
	"github.com/odysseia-greek/mykenai/archimedes/command/skene"
	"github.com/odysseia-greek/mykenai/archimedes/command/text"
	"github.com/spf13/cobra"
)

var version = "v0.1.0"

var rootCmd *cobra.Command

func main() {
	rootCmd = &cobra.Command{
		Use:   "archimedes",
		Short: "Deploy everything related to odysseia",
		Long: `Create and script everything odysseia related.
Allows you to parse words from a txt file,
build all container images
work with vault and much more is coming`,
		Version: version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Parent() == rootCmd || cmd.Name() == "help" {
				//https://patorjk.com/software/taag/#p=display&f=Crawford2&t=ARCHIMEDES
				logging.System(`
  ____  ____      __  __ __  ____  ___ ___    ___  ___      ___  _____
 /    ||    \    /  ]|  |  ||    ||   |   |  /  _]|   \    /  _]/ ___/
|  o  ||  D  )  /  / |  |  | |  | | _   _ | /  [_ |    \  /  [_(   \_
|     ||    /  /  /  |  _  | |  | |  \_/  ||    _]|  D  ||    _]\__  |
|  _  ||    \ /   \_ |  |  | |  | |   |   ||   [_ |     ||   [_ /  \ |
|  |  ||  .  \\     ||  |  | |  | |   |   ||     ||     ||     |\    |
|__|__||__|\_| \____||__|__||____||___|___||_____||_____||_____| \___|

`)
				logging.System(strings.Repeat("~", 37))
				logging.System("\"εὕρηκα\"")
				logging.System("\"I found it!\"")
				logging.System(strings.Repeat("~", 37))
			}
		},
	}

	rootCmd.AddCommand(
		images.Manager(),
		parse.Manager(),
		odysseia.Manager(),
		skene.Manager(),
		text.Manager(),
	)

	err := rootCmd.Execute()
	if err != nil {
		logging.Error(err.Error())
	}
}
