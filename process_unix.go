//go:build !windows

package main

import (
	"errors"
	"os"
	"syscall"
)

func isProcessAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Signal 0 performs error checking only: nil means the process exists,
	// EPERM means it exists but is owned by another user.
	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true
	}
	return errors.Is(err, syscall.EPERM)
}
