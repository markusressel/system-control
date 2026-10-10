package commandutil

import (
	"fmt"
	"time"

	"github.com/markusressel/system-control/internal/daemon"
	"github.com/spf13/cobra"
)

// LockedRunE wraps a Cobra RunE function with a cross-process lock.
// This guarantees that rapid executions (e.g., from scrolling a widget)
// will process sequentially instead of racing and reading stale hardware state.
// Commands that waited too long for the lock are dropped, see daemon.MaxQueueWait.
//
// Only needed for commands that are not implemented as a daemon.Op, those get
// their lock from Op.Lock.
func LockedRunE(lockName string, run func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		queuedAt := time.Now()
		_, err := daemon.WithLock(lockName, queuedAt, func() error {
			return run(cmd, args)
		})
		return err
	}
}

// PrintRunE wraps a command that mutates state and returns the lines to print.
func PrintRunE(run func(cmd *cobra.Command, args []string) ([]string, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		lines, err := run(cmd, args)
		if err != nil {
			return err
		}
		for _, line := range lines {
			fmt.Println(line)
		}
		return nil
	}
}
