package util

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

func ExecCommand(command, filePath string) error {
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Dir = filePath

	stdOut, _ := cmd.StdoutPipe()
	err := cmd.Start()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdOut)
	for scanner.Scan() {
		text := scanner.Text()
		fmt.Println(text)
	}
	cmd.Wait()

	return nil
}

func ExecCommandWithReturn(command, filePath string) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Dir = filePath

	stdOut, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	err = cmd.Start()
	if err != nil {
		return "", err
	}

	var textBuilder strings.Builder // Use strings.Builder to efficiently accumulate strings

	scanner := bufio.NewScanner(stdOut)
	for scanner.Scan() {
		textBuilder.WriteString(scanner.Text() + "\n") // Append each line to the builder
	}

	cmd.Wait()

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return textBuilder.String(), nil // Return the accumulated output as a single string
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
