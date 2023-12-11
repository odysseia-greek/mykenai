package command

import (
	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/odysseia-greek/agora/thales"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

func Install() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "install",
		Short: "generate docs",
		Long: `Allows you to create documentation for all apis
`,
		Run: func(cmd *cobra.Command, args []string) {
			kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
			data, _ := os.ReadFile(kubeconfigPath)
			kube, err := thales.NewFromConfig(data)
			if err != nil {
				logging.Error(errors.Wrap(err, "Failed to create new Kube client").Error())
				return
			}
			logging.Debug(kube.Host())
		},
	}

	return cmd
}
