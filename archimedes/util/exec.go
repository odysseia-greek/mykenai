package util

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ExecCommand(command, filePath string) error {
	return ExecCommandStreaming(command, filePath)
}

func ExecCommandStreaming(command, filePath string) error {
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Dir = filePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ExecCommandWithReturn(command, filePath string) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Dir = filePath

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrOutput := strings.TrimSpace(stderr.String())
		if stderrOutput != "" {
			return stdout.String(), fmt.Errorf("%w: %s", err, stderrOutput)
		}
		return stdout.String(), err
	}

	return stdout.String(), nil
}

func ExecCommandWithErrorCode(command, filePath string) error {
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Dir = filePath

	err := cmd.Start()
	if err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
