package command

import (
	"github.com/spf13/cobra"
)

func SetImageVersion() *cobra.Command {
	var (
		filePath string
	)
	cmd := &cobra.Command{
		Use:   "images",
		Short: "parse image version to helm chart",
		Long: `Allows you to set the versions for the helm chart
- Filepath
`,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	cmd.PersistentFlags().StringVarP(&filePath, "filepath", "f", "", "where to find the image version to set")

	return cmd
}
