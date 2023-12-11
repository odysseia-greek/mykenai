package images

import (
	"github.com/odysseia-greek/mykenai/archimedes/command/images/command"
	"github.com/spf13/cobra"
)

func Manager() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images",
		Short: "Work with images",
		Long: `The 'images' command is used to manage images in your project.
It supports a variety of operations such as creating images from a repo or creating a single image.
Use 'archimedes images [command] --help' for more information about a command.`,
	}

	cmd.AddCommand(
		command.CreateImagesFromRepo(),
		command.CreateSingleImage(),
	)

	return cmd
}
