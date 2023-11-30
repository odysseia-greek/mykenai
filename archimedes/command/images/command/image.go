package command

import (
	"fmt"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"os/exec"
)

func isDockerRunning() error {
	glg.Info("****** 🐋 Checking if docker is running 🐋 ******")
	command := "docker info"
	err := util.ExecCommandWithErrorCode(command, "/tmp")

	if err == nil {
		glg.Info("****** 🐳 Docker is Running 🐳 ******")
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() == 1 {
			glg.Info("****** 🛑 Docker is NOT Running stopping 🛑 ******")
			return err
		}
	}

	return nil
}

func buildImageMultiArch(rootPath, projectName, tag, dest string) error {
	glg.Info("****** 🖊️ Tagging Container Image 🖊️ ******")
	imageName := fmt.Sprintf("%s/%s:%s", dest, projectName, tag)
	glg.Infof("****** 📗 Tagged Image %s 📗 ******", imageName)

	glg.Info("****** 🔨 Building Container Image 🔨 ******")
	if projectName == hippokrates {
		projectName = projectName + ".test"
	}

	buildCommand := fmt.Sprintf("docker buildx build --platform=linux/arm64,linux/amd64 --build-arg project_name=%s -t %s . --push", projectName, imageName)

	output, err := util.ExecCommandWithReturn(buildCommand, rootPath)
	if err != nil {
		return err
	}

	glg.Debug(output)

	glg.Info("****** 🔱 Image Done 🔱 ******")

	return nil
}
