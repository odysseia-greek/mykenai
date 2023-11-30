package command

func runImageBuildFlow(projectPath, project, tag, dest string) error {
	err := isDockerRunning()
	if err != nil {
		return err
	}

	if project != hippokrates {
		err = runUnitTests(projectPath)
		if err != nil {
			return err
		}
		err = buildLocal(projectPath, project)
		if err != nil {
			return err
		}
	} else {
		err = buildLocalTestBin(projectPath, project)
		if err != nil {
			return err
		}
	}

	err = buildImageMultiArch(projectPath, project, tag, dest)
	if err != nil {
		return err
	}

	return nil
}
