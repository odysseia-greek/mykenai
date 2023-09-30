package command

import (
	"fmt"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"os"
	"path/filepath"
)

func buildImageWithLocalFile(projectPath, projectName, tag, dest string) error {
	glg.Info("****** 🖊️ Tagging Container Image 🖊️ ******")
	imageName := fmt.Sprintf("%s/%s:%s", dest, projectName, tag)
	glg.Infof("****** 📗 Tagged Image %s 📗 ******", imageName)

	glg.Info("****** 🔨 Building Container Image 🔨 ******")
	buildCommand := fmt.Sprintf("docker buildx build --platform=linux/arm64,linux/amd64 --build-arg project_name=%s -f %s -t %s . --push", projectName, dockerFile, imageName)
	err := util.ExecCommand(buildCommand, projectPath)
	if err != nil {
		return err
	}

	glg.Info("****** 🔱 Image Done 🔱 ******")

	return err
}

func buildLocal(path, projectName string) error {
	for _, arch := range ARCHS {

		fmtCommand := "go fmt ./..."
		err := util.ExecCommand(fmtCommand, path)
		if err != nil {
			return err
		}

		binPath := filepath.Join(path, distDirectory, binDirectory, linux, arch)
		err = os.MkdirAll(binPath, os.ModePerm)
		if err != nil {
			return err
		}

		projectBinPath := filepath.Join(binPath, projectName)
		buildCommand := fmt.Sprintf("GO111MODULE=on GOOS=%s GOARCH=%s CGO_ENABLED=0 go build main.go;mv main %s", linux, arch, projectBinPath)

		glg.Info("****** 🏗️ Building Golang Bin 🏗️ ******")
		err = util.ExecCommand(buildCommand, path)
		if err != nil {
			return err
		}

		glg.Info("****** 🏛️ Building Complete 🏛️ ******")

	}

	return nil
}

func buildLocalTestBin(path, projectName string) error {
	for _, arch := range ARCHS {

		fmtCommand := "go fmt ./..."
		err := util.ExecCommand(fmtCommand, path)
		if err != nil {
			return err
		}

		binPath := filepath.Join(path, distDirectory, binDirectory, linux, arch)
		err = os.MkdirAll(binPath, os.ModePerm)
		if err != nil {
			return err
		}

		testProjectName := projectName + ".test"
		projectBinPath := filepath.Join(binPath, testProjectName)
		buildCommand := fmt.Sprintf("GO111MODULE=on GOOS=%s GOARCH=%s CGO_ENABLED=0 go test -c -o %s;mv %s %s", linux, arch, testProjectName, testProjectName, projectBinPath)

		glg.Info("****** 🏗️ Building Golang Bin 🏗️ ******")
		err = util.ExecCommand(buildCommand, path)
		if err != nil {
			return err
		}

		glg.Info("****** 🏛️ Building Complete 🏛️ ******")

	}

	return nil
}
