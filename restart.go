package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

func RestartQtCreator(config *Config) error {
	PrintStep("Restarting Qt Creator...")

	if runtime.GOOS == "darwin" {
		cmd := exec.Command("open", "-a", config.QtCreatorPath)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to launch Qt Creator via open: %w", err)
		}
		return nil
	}

	execPath, err := GetQtCreatorExecutablePath(config.QtCreatorPath)
	if err != nil {
		return fmt.Errorf("failed to locate Qt Creator executable: %w", err)
	}

	cmd := exec.Command(execPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch Qt Creator: %w", err)
	}
	return nil
}
