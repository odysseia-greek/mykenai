package command

import (
	"fmt"
	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"os/exec"
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

func buildImageMultiArch(rootPath, projectName, tag, dest, target string) error {
	logging.Info("****** 🖊️ Tagging Container Image 🖊️ ******")
	imageName := fmt.Sprintf("%s/%s:%s", dest, projectName, tag)
	logging.Info(fmt.Sprintf("****** 📗 Tagged Image %s 📗 ******", imageName))

	logging.Info("****** 🔨 Building Container Image 🔨 ******")
	if projectName == hippokrates {
		projectName = projectName + ".test"
	}

	buildCommand := fmt.Sprintf("docker buildx build --platform=linux/arm64,linux/amd64 --target=%s --build-arg project_name=%s -t %s . --push", target, projectName, imageName)
	logging.Info(buildCommand)

	_, err := util.ExecCommandWithReturn(buildCommand, rootPath)
	if err != nil {
		return err
	}

	logging.Info("****** 🔱 Image Done 🔱 ******")

	return nil
}
