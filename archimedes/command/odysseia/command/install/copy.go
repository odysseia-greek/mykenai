package install

import (
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"os"
	"path/filepath"
)

func (a *AppInstaller) copyToCurrentDir() error {
	files, err := os.ReadDir(a.ConfigPath)
	if err != nil {
		glg.Error(err)
		return err
	}

	for _, f := range files {
		srcDir := filepath.Join(a.ConfigPath, f.Name())
		destDir := filepath.Join(a.CurrentPath, f.Name())

		err := util.CopyFileContents(srcDir, destDir)
		if err != nil {
			glg.Error(err)
			return err
		}
	}

	return nil
}
