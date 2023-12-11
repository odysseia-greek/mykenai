package command

import (
	"fmt"
	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var swaggerApis = [...]string{"alexandros", "dionysios", "herodotos", "solon", "sokrates"}
var gatewayApis = [...]string{"homeros", "euripides"}
var grpcApis = [...]string{command.Aristophanes, "ptolemaios"}

const (
	swaggerOutputPath string = "docs/swagger.json"
	openapiOutputPath string = "docs/openapi.yaml"
)

func GenerateDocs() *cobra.Command {
	var (
		rootPath string
	)
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "generate docs",
		Long: `Allows you to create documentation for all apis
`,
		Run: func(cmd *cobra.Command, args []string) {

			if rootPath == "" {
				currentDir, err := os.Getwd()
				if err != nil {
					return
				}

				logging.Debug(fmt.Sprintf("rootPath is empty defaulting to current dir %s", currentDir))
				rootPath = currentDir
			}

			generateDocs(rootPath)

		},
	}

	cmd.PersistentFlags().StringVarP(&rootPath, "root", "r", "", "rootpath when not using odysseia-settings")

	return cmd
}

func generateDocs(rootPath string) {
	for _, grpc := range grpcApis {
		var path string
		ploutarchosDocs := filepath.Join(rootPath, command.Olympia, command.Ploutarchos, command.Docs, "grpc")
		if grpc == command.Aristophanes {
			path = filepath.Join(rootPath, command.Attike, command.Aristophanes)
		}
		err := generateGRPC(ploutarchosDocs, command.Aristophanes, path)
		if err != nil {
			logging.Error(err.Error())
		}
	}

	for _, api := range swaggerApis {
		path := filepath.Join(rootPath, command.Olympia)
		if api == "solon" {
			path = filepath.Join(rootPath, command.Delphi)
		}
		ploutarchosDocs := filepath.Join(rootPath, command.Olympia, command.Ploutarchos, command.Docs)
		err := generateSwaggerFiles(path, ploutarchosDocs, api)
		if err != nil {
			logging.Error(err.Error())
		}
	}

	for _, gateway := range gatewayApis {
		apiPath := filepath.Join(rootPath, command.Olympia, gateway, command.Docs)
		err := generateSpectaql(rootPath, apiPath)
		if err != nil {
			logging.Error(err.Error())
		}
	}
}

func generateSwaggerFiles(rootPath, ploutarchosDocs, api string) error {
	logging.Info("****** 🗄️ Generating Swagger Docs 🗄️ ******")

	apiPath := filepath.Join(rootPath, api)
	buildCommand := fmt.Sprintf("docker run -v %s:%s -e SWAGGER_GENERATE_EXTENSION=true --workdir %s quay.io/goswagger/swagger generate spec -o %s -m", rootPath, rootPath, apiPath, swaggerOutputPath)

	err := util.ExecCommand(buildCommand, "/tmp")
	if err != nil {
		return err
	}

	logging.Info("****** 🗄️ Generated Swagger Docs 🗄️ ******")

	swaggerFile := filepath.Join(apiPath, swaggerOutputPath)
	openapiFile := filepath.Join(apiPath, openapiOutputPath)
	generateOpenApi(swaggerFile, openapiFile)
	logging.Info("****** 📋 Generated OpenApi Doc 📋 ******")

	ploutarchosPath := filepath.Join(ploutarchosDocs, "templates", fmt.Sprintf("%s.yaml", api))

	err = util.CopyFileContents(openapiFile, ploutarchosPath)
	if err != nil {
		return err
	}

	return nil
}

func generateOpenApi(swaggerFilePath, openapiFile string) error {
	logging.Info("****** 📋 Transforming OpenApi Doc 📋 ******")

	url := "https://converter.swagger.io/api/convert"
	headers := map[string]string{
		"Accept":       "application/yaml",
		"Content-Type": "application/json",
	}

	fileContent, err := os.ReadFile(swaggerFilePath)
	if err != nil {
		return err
	}

	requestBody := strings.NewReader(string(fileContent))
	client := &http.Client{}
	request, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		return err
	}

	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = os.WriteFile(openapiFile, responseBody, 0644)
	if err != nil {
		return err
	}

	return nil
}

func generateSpectaql(rootPath, spectaqlPath string) error {
	logging.Info("****** 📋 Generating Spectaql Doc 📋 ******")
	buildCommand := "npx spectaql spectaql.yaml"
	err := util.ExecCommand(buildCommand, spectaqlPath)
	if err != nil {
		return err
	}

	logging.Info("****** 📋 Generated Spectaql Doc 📋 ******")

	ploutarchosPath := filepath.Join(rootPath, command.Olympia, command.Ploutarchos, command.Docs, "public")
	spectaqlDir := filepath.Join(spectaqlPath, "public")
	err = util.CopyDir(spectaqlDir, ploutarchosPath)

	return nil
}

func generateGRPC(ploutarchosDocs, api, path string) error {
	logging.Info("****** 🗄️ Generating GRPC Docs 🗄️ ******")

	buildCommand := fmt.Sprintf("docker run -v %s/docs:/out -v %s/proto:/protos pseudomuto/protoc-gen-doc --doc_opt=html,%s.html", path, path, api)

	err := util.ExecCommand(buildCommand, path)
	if err != nil {
		return err
	}

	logging.Info("****** 🗄️ Generated GRPC Docs 🗄️ ******")

	fileName := fmt.Sprintf("%s.html", api)
	grpcFile := filepath.Join(path, command.Docs, fileName)
	ploutarchosPath := filepath.Join(ploutarchosDocs, fileName)

	err = util.CopyFileContents(grpcFile, ploutarchosPath)
	if err != nil {
		return err
	}

	return nil
}
