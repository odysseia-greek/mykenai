package command

import (
	"embed"
	"fmt"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	settings "github.com/odysseia-greek/mykenai/archimedes/command/config/command"
	"github.com/odysseia-greek/mykenai/archimedes/command/odysseia/command/install"
	"github.com/odysseia-greek/mykenai/archimedes/util/helm"
	kubernetes "github.com/odysseia-greek/thales"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

//go:embed "config"
var configPath embed.FS

//go:embed "apps"
var appsPath embed.FS

func Install() *cobra.Command {
	var (
		namespace      string
		kubePath       string
		installProfile string
		legacy         bool
		build          bool
	)
	cmd := &cobra.Command{
		Use:   "install",
		Short: "installs odysseia using your settings",
		Long: `Allows you to install odysseia
- profile: different setups are available that add on top of one another
	- base (just the basics to run local development)
	- infra (the whole needed infra part)
	- apps (all the apps that actually run the apis)
	- full (also include the frontend and docs)
	- tests (adds tests)
`,
		Run: func(cmd *cobra.Command, args []string) {

			if namespace == "" {
				glg.Debugf("defaulting to %s", command.DefaultNamespace)
				namespace = command.DefaultNamespace
			}

			if kubePath == "" {
				glg.Debugf("defaulting to %s", command.DefaultKubeConfig)
				homeDir, err := os.UserHomeDir()
				if err != nil {
					glg.Error(err)
				}

				kubePath = filepath.Join(homeDir, command.DefaultKubeConfig)
			}

			odysseiaSettings, err := settings.ReadOutConfig()
			if err != nil {
				glg.Error(err)
				glg.Warn("warning! You are about to install without a configfile this means archimedes will download everything needed to /tmp. After a reboot you will loose your helm charts. To avoid this from happening please run archimedes config set")

				odysseiaSettings, _ = settings.DownloadRepos("")
			}

			cfg, err := ioutil.ReadFile(kubePath)
			if err != nil {
				glg.Error("error getting kubeconfig")
			}

			kubeManager, err := kubernetes.NewKubeClient(cfg, namespace)
			if err != nil {
				glg.Fatal("error creating kubeclient")
			}

			helmManager, err := helm.NewHelmClient(cfg, namespace)
			if err != nil {
				glg.Fatal("error creating helmclient")
			}

			profile, err := kubeManager.Cluster().GetCurrentContext()
			if err != nil {
				glg.Fatal("error getting current context")
			}

			glg.Debugf("current profile is: %s", profile)

			glg.Info("getting config from yaml files")

			config, err := configPath.ReadFile(fmt.Sprintf("config/%s.yaml", profile))
			if err != nil {
				glg.Info(err.Error())
				if strings.Contains(err.Error(), "file does not exist") {
					config, _ = configPath.ReadFile("config/default.yaml")
				}
			}

			var valueOverwrite install.ValueOverwrite
			err = yaml.Unmarshal(config, &valueOverwrite)
			if err != nil {
				glg.Fatal("error marshalling yaml")
			}

			glg.Info("creating a new install for odysseia")

			elasticConfig := install.ElasticOperator{
				Name: command.ElasticOperatorName,
			}

			if installProfile == "" {
				glg.Debugf("defaulting to %s", command.DefaultAppProfile)
				installProfile = command.DefaultAppProfile
			}

			ati, err := readAppsYaml(installProfile)
			if err != nil {
				glg.Fatal(err)
			}

			newConfig := command.CurrentInstallConfig{
				ElasticPassword: "",
				HarborPassword:  "",
				VaultRootToken:  "",
				VaultUnsealKey:  "",
			}

			odysseia := install.AppInstaller{
				Namespace:        namespace,
				ConfigPath:       "",
				CurrentPath:      "",
				ThemistoklesRoot: odysseiaSettings.HelmPath,
				OdysseiaRoot:     odysseiaSettings.SourcePath,
				Charts:           install.Themistokles{},
				Config:           newConfig,
				ValueConfig:      valueOverwrite,
				Kube:             kubeManager,
				Helm:             helmManager,
				ElasticConfig:    elasticConfig,
				Profile:          profile,
				Harbor:           nil,
				AppsToInstall:    ati,
				Build:            build,
			}

			err = odysseia.InstallOdysseiaComplete(legacy)
			if err != nil {
				glg.Error(err)
				os.Exit(1)
			}
		},
	}
	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace defaults to odysseia")
	cmd.PersistentFlags().StringVarP(&kubePath, "kubepath", "k", "", "kubeconfig filepath defaults to ~/.kube/config")
	cmd.PersistentFlags().StringVarP(&installProfile, "profile", "p", "", "the profile to use when installing")
	cmd.PersistentFlags().BoolVarP(&legacy, "legacy", "l", false, "install legacy elastic with helm chart")
	cmd.PersistentFlags().BoolVarP(&build, "build", "b", true, "whether to build images")

	return cmd
}

func readAppsYaml(profile string) ([]string, error) {
	appsToInstall, err := appsPath.ReadFile(fmt.Sprintf("apps/%s.yaml", profile))
	if err != nil {
		return nil, err
	}

	var ati []string

	var a install.AppsToInstall
	err = yaml.Unmarshal(appsToInstall, &a)
	if err != nil {
		return nil, err
	}

	ati = append(ati, a.AppsToInstall...)

	if a.Include != "" {
		b, err := readAppsYaml(a.Include)
		if err != nil {
			return nil, err
		}

		ati = append(ati, b...)
	}

	return ati, nil
}
