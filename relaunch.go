package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// RelaunchApplication starts the given application detached from this process,
// so the updater can exit without terminating it.
func RelaunchApplication(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("relaunch target not found: %w", err)
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" && strings.HasSuffix(path, ".app") {
		cmd = exec.Command("open", path)
	} else {
		cmd = exec.Command(path)
	}

	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", path, err)
	}

	return cmd.Process.Release()
}
