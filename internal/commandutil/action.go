package commandutil

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"github.com/spf13/cobra"
)

// LockedRunE wraps a Cobra RunE function with a cross-process file lock.
// This guarantees that rapid executions (e.g., from scrolling a widget)
// will process sequentially instead of racing and reading stale hardware state.
func LockedRunE(lockName string, run func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		queuedAt := time.Now()

		tmpDir := filepath.Join(os.TempDir(), "system-control")
		if err := os.MkdirAll(tmpDir, 0755); err != nil {
			log.Fatalf("Could not create temp directory: %v", err)
		}

		fileLock := flock.New(filepath.Join(tmpDir, lockName+".lock"))
		if err := fileLock.Lock(); err != nil {
			log.Fatalf("Could not acquire lock %s: %v", lockName, err)
		}
		defer fileLock.Unlock()

		if time.Since(queuedAt) > 200*time.Millisecond {
			log.Printf("Dropping queued command %s: took too long to acquire lock", cmd.Name())
			return nil
		}

		return run(cmd, args)
	}
}

// MutatorResult represents the outcome of a mutator command that should be output to stdout and pushed to the daemon
type MutatorResult struct {
	CommandId string
	Flags     map[string]string
	Value     string
}

// MutatorRunE wraps a command that mutates state. It automatically handles printing the resulting state
// and pushing it to the daemon IPC if running.
func MutatorRunE(run func(cmd *cobra.Command, args []string) ([]MutatorResult, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		results, err := run(cmd, args)
		if err != nil {
			return err
		}

		for _, res := range results {
			fmt.Println(res.Value)
			PushState(res.CommandId, res.Flags, res.Value)
		}

		return nil
	}
}
