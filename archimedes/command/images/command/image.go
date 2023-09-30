package command

import (
	"fmt"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"os"
	"os/exec"
	"path/filepath"
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

func buildImage(binPath, rootPath, projectName, tag string, arm bool) error {
	projectPath := filepath.Join(rootPath, projectName)

	dockerSrc := filepath.Join(rootPath, dockerFile)
	dockerDest := filepath.Join(projectPath, dockerFile)
	err := util.CopyFileContents(dockerSrc, dockerDest)
	if err != nil {
		return err
	}

	binDest := filepath.Join(projectPath, projectName)
	err = util.CopyFileContents(binPath, binDest)
	if err != nil {
		return err
	}

	err = os.Chmod(binDest, os.ModePerm)
	if err != nil {
		return err
	}

	glg.Info("****** 🔨 Building Container Image 🔨 ******")
	imageName := fmt.Sprintf("%s:%s", projectName, tag)
	buildCommand := fmt.Sprintf("docker build --build-arg project_name=%s -f %s -t %s . --no-cache", projectName, dockerFile, imageName)
	if arm {
		buildCommand = fmt.Sprintf("docker buildx build --build-arg project_name=%s --platform linux/arm64 -f %s -t ghcr.io/odysseia-greek/%s-%s . --no-cache --push", projectName, dockerFile, imageName, "arm64")
	}
	err = util.ExecCommand(buildCommand, projectPath)
	if err != nil {
		return err
	}

	glg.Info("****** 🔱 Image Done 🔱 ******")

	err = os.Remove(binDest)
	if err != nil {
		return err
	}

	err = os.Remove(dockerDest)
	if err != nil {
		return err
	}

	return nil
}

func buildImageMultiArch(rootPath, projectName, tag, dest string) error {
	projectPath := filepath.Join(rootPath, projectName)

	if projectName == hippokrates || projectName == eupalinos {
		projectPath = rootPath
	}

	var dockerDest string
	if projectName != ploutarchos && projectName != hippokrates && projectName != eupalinos {
		dockerSrc := filepath.Join(rootPath, dockerFile)
		dockerDest = filepath.Join(projectPath, dockerFile)
		err := util.CopyFileContents(dockerSrc, dockerDest)
		if err != nil {
			return err
		}
	}

	glg.Info("****** 🖊️ Tagging Container Image 🖊️ ******")
	imageName := fmt.Sprintf("%s/%s:%s", dest, projectName, tag)
	glg.Infof("****** 📗 Tagged Image %s 📗 ******", imageName)

	glg.Info("****** 🔨 Building Container Image 🔨 ******")
	if projectName == hippokrates {
		projectName = projectName + ".test"
	}

	buildCommand := fmt.Sprintf("docker buildx build --platform=linux/arm64,linux/amd64 --build-arg project_name=%s -f %s -t %s . --push", projectName, dockerFile, imageName)

	output, err := util.ExecCommandWithReturn(buildCommand, projectPath)
	if err != nil {
		return err
	}

	glg.Debug(output)

	glg.Info("****** 🔱 Image Done 🔱 ******")

	if projectName != ploutarchos && projectName != hippokrates && projectName != eupalinos {
		err = os.Remove(dockerDest)
		if err != nil {
			return err
		}
	}

	return nil
}

func pushImage(projectName, tag, destRepo, rootPath string, minikube bool) error {
	if minikube {
		glg.Info("****** 🚢️ Loading Container Image 🚢 ******")
		imageName := fmt.Sprintf("%s:%s", projectName, tag)
		pushCommand := fmt.Sprintf("minikube image load %s", imageName)
		err := util.ExecCommand(pushCommand, rootPath)
		if err != nil {
			return err
		}

		glg.Info("****** 🚀 Container Image Laoded 🚀 ******")

		return nil
	}

	newTag := fmt.Sprintf("%s/%s:%s", destRepo, projectName, tag)

	glg.Info("****** 🖊️ Tagging Container Image 🖊️ ******")
	imageName := fmt.Sprintf("%s:%s", projectName, tag)
	tagCommand := fmt.Sprintf("docker tag %s %s", imageName, newTag)
	err := util.ExecCommand(tagCommand, rootPath)
	if err != nil {
		return err
	}

	glg.Infof("****** 📗 Tagged Image %s 📗 ******", newTag)

	glg.Info("****** 🚢️ Pushing Container Image 🚢 ******")
	pushCommand := fmt.Sprintf("docker push %s", newTag)
	err = util.ExecCommand(pushCommand, rootPath)
	if err != nil {
		return err
	}

	glg.Info("****** 🚀 Pushed Container Image 🚀 ******")
	glg.Info("image can be pulled as:")
	glg.Info(newTag)

	return nil
}
