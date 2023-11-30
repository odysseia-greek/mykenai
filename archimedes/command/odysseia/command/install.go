package command

import (
	"embed"
	"fmt"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/agora/thales"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	settings "github.com/odysseia-greek/mykenai/archimedes/command/config/command"
	"github.com/odysseia-greek/mykenai/archimedes/command/odysseia/command/install"
	"github.com/odysseia-greek/mykenai/archimedes/util/helm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

//go:embed "config"
var configPath embed.FS

//go:embed "apps"
var appsPath embed.FS

func Install() *cobra.Command {
	var (
		namespace      string
		kubePath       string
		env            string
		profile        string
		installProfile string
		pathToVaultSa  string
		keyRing        string
		cryptoKey      string
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

			if env == "" {
				env = command.DefaultEnv
			}

			if profile == "" {
				profile = command.DefaultProfile
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
			}

			var vaultUnsealMethod string
			if pathToVaultSa != "" {
				if _, err := os.Stat(pathToVaultSa); os.IsNotExist(err) {
					glg.Fatal("file at path %s does not exist", pathToVaultSa)
				}

				vaultUnsealMethod = "gcp"
			}

			cfg, err := os.ReadFile(kubePath)
			if err != nil {
				glg.Error("error getting kubeconfig")
			}

			kubeManager, err := thales.NewKubeClient(cfg, namespace)
			if err != nil {
				glg.Fatal("error creating kubeclient")
			}

			helmManager, err := helm.NewHelmClient(cfg, namespace)
			if err != nil {
				glg.Fatal("error creating helmclient")
			}

			config, err := configPath.ReadFile(fmt.Sprintf("config/%s.yaml", profile))
			if err != nil {
				glg.Info(err.Error())
			}

			envProfile, err := configPath.ReadFile(fmt.Sprintf("config/%s.yaml", env))
			if err != nil {
				glg.Info(err.Error())
			}

			var configOverwrite map[string]interface{}
			err = yaml.Unmarshal(config, &configOverwrite)
			if err != nil {
				glg.Fatal("error marshalling yaml")
			}

			var envOverwrite map[string]interface{}
			err = yaml.Unmarshal(envProfile, &envOverwrite)
			if err != nil {
				glg.Fatal("error marshalling yaml")
			}

			// Merge k3dConfig into envOverwrite, overwriting values where keys match.
			mergeMaps(envOverwrite, configOverwrite)

			addConfigFields(envOverwrite, env, profile)

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

			odysseia := install.AppInstaller{
				Namespace:         namespace,
				ConfigPath:        "",
				CurrentPath:       "",
				ThemistoklesRoot:  odysseiaSettings.HelmPath,
				OdysseiaRoot:      odysseiaSettings.SourcePath,
				Charts:            install.Themistokles{},
				ValueConfig:       envOverwrite,
				Kube:              kubeManager,
				Helm:              helmManager,
				ElasticConfig:     elasticConfig,
				Profile:           profile,
				AppsToInstall:     ati,
				Build:             build,
				VaultSaPath:       pathToVaultSa,
				CryptoKey:         cryptoKey,
				KeyRing:           keyRing,
				VaultUnsealMethod: vaultUnsealMethod,
			}

			err = odysseia.InstallOdysseiaComplete()
			if err != nil {
				glg.Error(err)
				os.Exit(1)
			}
		},
	}
	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace defaults to odysseia")
	cmd.PersistentFlags().StringVarP(&kubePath, "kubepath", "k", "", "kubeconfig filepath defaults to ~/.kube/config")
	cmd.PersistentFlags().StringVarP(&installProfile, "install-profile", "p", "", "the install profile for apps (tests, infra  etc)")
	cmd.PersistentFlags().StringVarP(&env, "env", "e", "", "the env to use when installing (local, prod)")
	cmd.PersistentFlags().StringVarP(&profile, "profile", "d", "", "the profile to use when installing (k3d, k3s, digital-ocean")
	cmd.PersistentFlags().StringVarP(&pathToVaultSa, "savault", "s", "", "path to vault sa to use for auto unsealing")
	cmd.PersistentFlags().StringVarP(&cryptoKey, "cryptokey", "c", "", "gcp cryptoring")
	cmd.PersistentFlags().StringVarP(&keyRing, "keyring", "r", "", "gcp keyring")
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

func mergeMaps(dest, src map[string]interface{}) {
	for key, srcValue := range src {
		destValue, ok := dest[key]
		if !ok {
			// Key doesn't exist in the destination map, so we add it.
			dest[key] = srcValue
		} else {
			// Key exists in the destination map, we need to merge if it's a nested map.
			destMap, destMapOK := destValue.(map[string]interface{})
			srcMap, srcMapOK := srcValue.(map[string]interface{})
			if srcMapOK && destMapOK {
				mergeMaps(destMap, srcMap)
			} else {
				// Not a map, overwrite the value in the destination.
				dest[key] = srcValue
			}
		}
	}
}

// Traverse the merged map and add environment and kubeVariant within "config" sections.
func addConfigFields(m map[string]interface{}, env, variant string) {
	for _, value := range m {
		subMap, isMap := value.(map[string]interface{})
		if isMap {
			if config, exists := subMap["config"]; exists {
				configMap, isConfigMap := config.(map[string]interface{})
				if isConfigMap {
					// Add "environment" and "kubeVariant" to the "config" section.
					configMap["environment"] = env
					configMap["kubeVariant"] = variant
				}
			}
			// Continue recursively within the sub-map.
			addConfigFields(subMap, env, variant)
		}
	}
}
