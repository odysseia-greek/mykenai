package command

import (
	"github.com/kpango/glg"
	"github.com/manifoldco/promptui"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Settings struct {
	SourcePath  string `yaml:"sourcePath"`
	OlympiaPath string `yaml:"olympiaPath"`
	DelphiPath  string `yaml:"delphiPath"`
	AgoraPath   string `yaml:"agoraPath"`
	HelmPath    string `yaml:"helmPath"`
}

func Set() *cobra.Command {
	var (
		sourcePath string
		helmPath   string
	)
	cmd := &cobra.Command{
		Use:   "set",
		Short: "sets your environment up for the first time",
		Long: `Sets your environment in a local file for future reference.
If your SourcePath has not been set it will prompt you to provide one. DownloadPath will default to /tmp
- SourcePath
- HelmPath
`,
		Run: func(cmd *cobra.Command, args []string) {
			glg.Green("Creating odysseia settings so all other commands can use these defaults")
			odysseiaSettings := &Settings{}
			var err error

			if sourcePath == "" && odysseiaSettings.SourcePath == "" {
				odysseiaSettings, err = Gather("", "")
				if err != nil {
					glg.Fatal(err)
				}
			}

			if sourcePath != "" {
				odysseiaSettings.SourcePath = sourcePath
			}

			if helmPath == "" && odysseiaSettings.HelmPath == "" {
				var source string
				if odysseiaSettings.SourcePath != "" {
					source = odysseiaSettings.SourcePath
				} else {
					source = sourcePath
				}
				odysseiaSettings, err = Gather(source, "")
				if err != nil {
					glg.Fatal(err)
				}
			}

			if helmPath != "" {
				odysseiaSettings.HelmPath = helmPath
			}

			writeToConfig(*odysseiaSettings)
		},
	}
	cmd.PersistentFlags().StringVarP(&helmPath, "helm", "m", "", "where to find the helm chart")
	cmd.PersistentFlags().StringVarP(&sourcePath, "source", "s", "", "where to find the source code")

	return cmd
}

func writeToConfig(settings Settings) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		glg.Error(err)
	}
	configDir := filepath.Join(homeDir, command.ConfigFilePath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		glg.Infof("config directory does not exist... creating at: %s", configDir)
		os.Mkdir(configDir, 0755)
	}

	marshalledSettings, err := yaml.Marshal(settings)
	if err != nil {
		glg.Error(err)
	}

	glg.Info(string(marshalledSettings))
	settingsPath := filepath.Join(configDir, command.SettingsName)
	util.WriteFile(marshalledSettings, settingsPath)
}

func ReadOutConfig() (*Settings, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		glg.Error(err)
	}

	settingsPath := filepath.Join(homeDir, command.ConfigFilePath, command.SettingsName)
	cfg, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, err
	}

	var s Settings
	err = yaml.Unmarshal(cfg, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func Gather(sourcePath, helmPath string) (*Settings, error) {
	validate := func(input string) error {
		if _, err := os.Stat(input); os.IsNotExist(err) {
			return err
		}

		files, err := os.ReadDir(input)
		if err != nil {
			return err
		}

		for _, file := range files {
			glg.Info(file.Name())
		}

		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	if sourcePath == "" {
		example := filepath.Join(homeDir, command.DefaultExamplePath)

		glg.Infof("SourcePath is empty so you will be prompted to provide one, typically it will something like: %s", example)

		prompt := promptui.Prompt{
			Label:    "Source Path",
			Validate: validate,
		}

		result, err := prompt.Run()

		if err != nil {
			return nil, err
		}

		sourcePath = result
	}

	glg.Infof("SourcePath has been set to: %s", sourcePath)

	if helmPath == "" {
		mykenaiPath := filepath.Join(sourcePath, command.DefaultMykenai)
		files, err := os.ReadDir(mykenaiPath)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			if file.Name() == command.DefaultHelmChartName {
				helmPath = filepath.Join(mykenaiPath, file.Name(), command.DefaultNamespace, "charts")
				glg.Info("HelmPath found in your SourcePath defaulting there")
				break
			}
		}

		if helmPath == "" {
			prompt := promptui.Prompt{
				Label:    "Helm Path",
				Validate: validate,
			}

			result, err := prompt.Run()

			if err != nil {
				return nil, err
			}

			helmPath = result
		}

		glg.Infof("HelmPath has been set to: %s", helmPath)
	}

	olympiaPath := filepath.Join(sourcePath, command.Olympia)
	agoraPath := filepath.Join(sourcePath, command.Agora)
	delphiPath := filepath.Join(sourcePath, command.Delphi)

	return &Settings{
		OlympiaPath: olympiaPath,
		AgoraPath:   agoraPath,
		DelphiPath:  delphiPath,
		SourcePath:  sourcePath,
		HelmPath:    helmPath,
	}, nil

}
