package command

import (
	"fmt"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	settings "github.com/odysseia-greek/mykenai/archimedes/command/config/command"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func Parse() *cobra.Command {
	var (
		imagePath       string
		helmPath        string
		destinationRepo string
		version         string
		local           bool
	)
	cmd := &cobra.Command{
		Use:   "parse",
		Short: "parses your",
		Long: `Sets your environment in a local file for future reference.
If your SourcePath has not been set it will prompt you to provide one. DownloadPath will default to /tmp
- SourcePath
- HelmPath
`,
		Run: func(cmd *cobra.Command, args []string) {
			odysseiaSettings, err := settings.ReadOutConfig()
			if err != nil {
				glg.Error(err)
				glg.Warn("warning! You are about to install without a configfile this means archimedes will download everything needed to /tmp. After a reboot you will loose your helm charts. To avoid this from happening please run archimedes config set")

				odysseiaSettings, _ = settings.DownloadRepos("")
			}

			if helmPath == "" && odysseiaSettings.HelmPath == "" {
				odysseiaSettings, err = settings.Gather(odysseiaSettings.SourcePath, "")
				if err != nil {
					glg.Fatal(err)
				}
			}

			if helmPath == "" {
				helmPath = odysseiaSettings.HelmPath
			}

			if imagePath == "" {
				helmDir := filepath.Dir(helmPath)
				dirParts := strings.Split(helmDir, string(os.PathSeparator))
				var grandparentDir []string
				for i, part := range dirParts {
					if strings.Contains(part, command.DefaultHelmChartName) {
						grandparentDir = dirParts[0 : i+1]
					}
				}

				imageBasePath := "/"
				for _, path := range grandparentDir {
					imageBasePath = filepath.Join(imageBasePath, path)
				}

				imagePath = filepath.Join(imageBasePath, command.DefaultImageFile)

			}

			if !local {
				if destinationRepo == "" {
					destinationRepo = command.OdysseiaImageRepo
				}
			}

			err = parseImagesToCharts(helmPath, imagePath, destinationRepo, version)
			if err != nil {
				glg.Error(err)
			}

		},
	}
	cmd.PersistentFlags().StringVarP(&helmPath, "themistokles", "t", "", "where to find the helm charts")
	cmd.PersistentFlags().StringVarP(&imagePath, "images", "i", "", "image file")
	cmd.PersistentFlags().StringVarP(&destinationRepo, "dest", "d", "", "destination repo address")
	cmd.PersistentFlags().StringVarP(&version, "version", "v", "", "version of the helm chart")
	cmd.PersistentFlags().BoolVarP(&local, "local", "l", false, "will not set a destination repo")

	return cmd
}

func parseImagesToCharts(helmPath, imagePath, repo, version string) error {
	yamlData, err := os.ReadFile(imagePath)
	if err != nil {
		return err
	}

	var config Config
	err = yaml.Unmarshal(yamlData, &config)
	if err != nil {
		return err
	}

	dirs, err := os.ReadDir(helmPath)
	if err != nil {
		return err
	}

	configValue := reflect.ValueOf(config)
	for i := 0; i < configValue.NumField(); i++ {
		field := configValue.Field(i)
		fieldName := configValue.Type().Field(i).Name
		api := strings.ToLower(fieldName)

		if field.Interface() != reflect.Zero(field.Type()).Interface() {
			for _, dir := range dirs {
				if api == dir.Name() {
					app, ok := field.Interface().(Application)
					if !ok {
						glg.Errorf("failed to cast field '%s' to Application struct", fieldName)
						continue
					}
					appDir := filepath.Join(helmPath, dir.Name())
					if err := readAndReplaceHelmChart(app, appDir, repo, version); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func updateImageTags(imageValues *ImagesConfig, values Application) {
	if values.Deployment != "" {
		imageValues.OdysseiaAPI.Tag = values.Deployment
	}
	if values.Init != "" {
		imageValues.Init.Tag = values.Init
	}

	if values.Seeder != "" {
		imageValues.Seeder.Tag = values.Seeder
	}

	if values.Sidecar != "" {
		imageValues.Sidecar.Tag = values.Sidecar
	}
	if values.Job != "" {
		imageValues.Job.Tag = values.Job
	}
	if values.JobInit != "" {
		imageValues.JobInit.Tag = values.JobInit
	}
	if values.System != "" {
		imageValues.System.Tag = values.System
	}
	if values.Stateful != "" {
		imageValues.Stateful.Tag = values.Stateful
	}
	if values.Tracer != "" {
		imageValues.Tracer.Tag = values.Tracer
	}
}

func readAndReplaceHelmChart(values Application, path, dest, version string) error {
	valuesPath := filepath.Join(path, command.HelmValuesFile)

	yamlData, err := os.ReadFile(valuesPath)
	if err != nil {
		return fmt.Errorf("error reading YAML file: %w", err)
	}

	var dataStruct map[string]interface{}
	if err := yaml.Unmarshal(yamlData, &dataStruct); err != nil {
		return fmt.Errorf("error unmarshaling YAML: %w", err)
	}

	imagesNode := findNode(dataStruct, "images")
	if imagesNode == nil {
		return fmt.Errorf("error: failed to find 'images' node")
	}

	var imageValues ImageValues
	yamlString, _ := yaml.Marshal(imagesNode)
	if err := yaml.Unmarshal(yamlString, &imageValues.Images); err != nil {
		return fmt.Errorf("error unmarshaling 'images' node: %w", err)
	}

	config, ok := dataStruct["config"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("error: failed to access config field")
	}

	if dest != "" {
		imageValues.Images.ImageRepo = dest
		config["externalRepo"] = true
		config["pullPolicy"] = "Always"
		config["privateImagesInRepo"] = strings.Contains(dest, "harbor")
	} else {
		config["pullPolicy"] = "Never"
		config["privateImagesInRepo"] = false
		config["externalRepo"] = false
	}

	updateImageTags(&imageValues.Images, values)

	dataStruct["images"] = imageValues.Images

	glg.Infof("writing update yaml file to: %s", valuesPath)
	updatedYamlData, err := yaml.Marshal(&dataStruct)
	if err != nil {
		return fmt.Errorf("error marshaling updated YAML: %w", err)
	}

	if err := ioutil.WriteFile(valuesPath, updatedYamlData, 0644); err != nil {
		return fmt.Errorf("error writing updated YAML: %w", err)
	}

	appVersion := getAppVersion(values)
	if err := updateChartFile(path, appVersion, version); err != nil {
		return fmt.Errorf("error updating chart file: %w", err)
	}

	return nil
}

func getAppVersion(values Application) string {
	version := "v0.0.1"
	if values.Deployment != "" {
		version = values.Deployment
	} else if values.Job != "" {
		version = values.Job
	} else if values.Stateful != "" {
		version = values.Stateful
	}

	return version
}

func findNode(data interface{}, key string) interface{} {
	switch node := data.(type) {
	case map[string]interface{}:
		if value, ok := node[key]; ok {
			return value
		}
		for _, v := range node {
			if value := findNode(v, key); value != nil {
				return value
			}
		}
	case []interface{}:
		for _, v := range node {
			if value := findNode(v, key); value != nil {
				return value
			}
		}
	}
	return nil
}

func updateChartFile(path, appVersion, chartVersion string) error {
	chartPath := filepath.Join(path, command.HelmChartFile)
	yamlData, err := os.ReadFile(chartPath)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(yamlData, &data); err != nil {
		return err
	}

	var parsedVersion string
	if strings.Contains(chartVersion, "v") {
		parsedVersion = strings.Split(chartVersion, "v")[1]
	} else {
		parsedVersion = chartVersion
	}

	var parsedAppVerison string
	if strings.Contains(appVersion, "v") {
		parsedAppVerison = strings.Split(appVersion, "v")[1]
	} else {
		parsedAppVerison = appVersion
	}

	data["appVersion"] = parsedAppVerison
	if chartVersion != "" {
		data["version"] = parsedVersion
	}

	updatedYamlData, err := yaml.Marshal(&data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(chartPath, updatedYamlData, 0644)
	if err != nil {
		return err
	}

	return nil
}
