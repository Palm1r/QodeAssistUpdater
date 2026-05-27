package main

import (
	"fmt"
	"time"
)

const (
	waitPidPollInterval = 200 * time.Millisecond
	waitPidDefaultLimit = 30 * time.Second
)

func WaitForProcessExit(pid int, timeout time.Duration) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid: %d", pid)
	}

	if !processAlive(pid) {
		return nil
	}

	PrintStep(fmt.Sprintf("Waiting for process %d to exit...", pid))

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			return nil
		}
		time.Sleep(waitPidPollInterval)
	}
	return fmt.Errorf("process %d did not exit within %s", pid, timeout)
}
