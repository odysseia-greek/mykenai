package command

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/odysseia-greek/mykenai/archimedes/util"
)

func isDockerRunning() error {
	logging.Info("****** 🐋 Checking if docker is running 🐋 ******")
	command := "docker info"
	err := util.ExecCommandWithErrorCode(command, "/tmp")

	if err == nil {
		logging.Info("****** 🐳 Docker is Running 🐳 ******")
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() == 1 {
			logging.Info("****** 🛑 Docker is NOT Running stopping 🛑 ******")
			return err
		}
	}

	return nil
}

func getContainerFileName(rootPath string) string {
	containerFile := filepath.Join(rootPath, "Containerfile")
	if _, err := os.Stat(containerFile); err == nil {
		return "Containerfile"
	}

	return "Dockerfile"
}

func buildImageMultiArch(rootPath, projectName, tag, dest, target string) error {
	logging.Info("****** 🖊️ Tagging Container Image 🖊️ ******")
	imageName := fmt.Sprintf("%s/%s:%s", dest, projectName, tag)
	logging.Info(fmt.Sprintf("****** 📗 Tagged Image %s 📗 ******", imageName))

	logging.Info("****** 🔨 Building Container Image 🔨 ******")
	if projectName == hippokrates {
		projectName = projectName + ".test"
	}

	containerFile := getContainerFileName(rootPath)

	buildCommand := fmt.Sprintf("docker buildx build --platform=linux/arm64,linux/amd64 -f %s --target=%s --build-arg project_name=%s --build-arg VERSION=%s -t %s . --push", containerFile, target, projectName, tag, imageName)
	logging.Info(buildCommand)

	err := util.ExecCommandStreaming(buildCommand, rootPath)
	if err != nil {
		return err
	}

	logging.Info("****** 🔱 Image Done 🔱 ******")

	return nil
}

func buildImages(rootPath, projectName, tag, dest, target string) error {
	logging.Info("****** 🖊️ Tagging Container Image 🖊️ ******")
	imageName := fmt.Sprintf("%s/%s:%s", dest, projectName, tag)
	logging.Info(fmt.Sprintf("****** 📗 Tagged Image %s 📗 ******", imageName))

	logging.Info("****** 🔨 Building Container Image 🔨 ******")
	if projectName == hippokrates {
		projectName = projectName + ".test"
	}

	containerFile := getContainerFileName(rootPath)

	buildCommand := fmt.Sprintf("docker build -f %s --target=%s --build-arg project_name=%s --build-arg VERSION=%s -t %s . --push", containerFile, target, projectName, tag, imageName)
	logging.Info(buildCommand)

	err := util.ExecCommandStreaming(buildCommand, rootPath)
	if err != nil {
		return err
	}

	logging.Info("****** 🔱 Image Done 🔱 ******")

	return nil
}
