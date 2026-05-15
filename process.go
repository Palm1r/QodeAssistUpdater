package main

import (
	"fmt"
	"time"
)

// WaitForProcessExit blocks until the process with the given pid is no longer
// running, or until the timeout elapses. A timeout of 0 waits indefinitely.
func WaitForProcessExit(pid int, timeout time.Duration) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid: %d", pid)
	}

	deadline := time.Now().Add(timeout)
	for {
		if !isProcessAlive(pid) {
			return nil
		}
		if timeout > 0 && time.Now().After(deadline) {
			return fmt.Errorf("timed out after %s waiting for process %d to exit", timeout, pid)
		}
		time.Sleep(200 * time.Millisecond)
	}
}
