package command

import (
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/util"
)

func runUnitTests(path string) error {
	cmd := "go test -short -cover ./..."

	glg.Info("****** 🔎 Running Unittests 🔎 ******")
	err := util.ExecCommand(cmd, path)
	if err != nil {
		glg.Error("****** ❌ Unittests Failed ❌ ******")
		return err
	}

	glg.Info("****** ✅ Unittests Passed ✅ ******")

	return nil
}
