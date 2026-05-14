package session

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func findProcessIDsByName(processName string) ([]int, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	pids := make([]int, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		commPath := filepath.Join("/proc", entry.Name(), "comm")
		commBytes, err := os.ReadFile(commPath)
		if err != nil {
			// Process may have exited between listing /proc and reading comm.
			continue
		}

		comm := strings.TrimSpace(string(commBytes))
		if comm == processName {
			pids = append(pids, pid)
		}
	}

	return pids, nil
}

func terminateProcessesByName(processName string) error {
	pids, err := findProcessIDsByName(processName)
	if err != nil {
		return err
	}

	for _, pid := range pids {
		err = syscall.Kill(pid, syscall.SIGTERM)
		if err != nil && !errors.Is(err, syscall.ESRCH) {
			return fmt.Errorf("failed to terminate process %s (%d): %w", processName, pid, err)
		}
	}

	return nil
}

func isProcessRunning(processName string) (bool, error) {
	pids, err := findProcessIDsByName(processName)
	if err != nil {
		return false, err
	}

	return len(pids) > 0, nil
}
