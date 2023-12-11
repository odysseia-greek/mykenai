package util

import (
	"encoding/json"
	"fmt"
	"github.com/odysseia-greek/agora/plato/logging"
	"io"
	"os"
	"path/filepath"
)

func WriteFile(input []byte, outputFile string) error {
	openedFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer openedFile.Close()

	outputFromWrite, err := openedFile.Write(input)
	if err != nil {
		return err
	}

	logging.Info(fmt.Sprintf("finished writing %d bytes", outputFromWrite))
	logging.Info(fmt.Sprintf("file written to %s", outputFile))

	return nil
}

func WriteJSONToFilePrettyPrint(data interface{}, outputFile string) error {
	// Marshal the data into an indented JSON format
	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Create or overwrite the output file
	openedFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer openedFile.Close()

	// Write the indented JSON to the file
	_, err = openedFile.Write(prettyJSON)
	if err != nil {
		return err
	}

	logging.Info(fmt.Sprintf("file written to %s", outputFile))
	return nil
}

func CopyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func CopyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	dirEntries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range dirEntries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
