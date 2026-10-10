package util

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// DefaultCommandTimeout bounds every ExecCommand call, so a hung external tool can't block
// the caller (and the locks it holds) forever.
const DefaultCommandTimeout = 60 * time.Second

// Runner executes an external command and returns its trimmed stdout.
type Runner func(ctx context.Context, command string, args ...string) (string, error)

var runner Runner = runCommand

// SetRunner replaces the function used to execute external commands and returns a function that
// restores the previous one. It exists to test code that shells out.
func SetRunner(r Runner) (restore func()) {
	previous := runner
	runner = r
	return func() { runner = previous }
}

// ExecCommandAndFork executes a shell command with the given arguments
// and disowns the process, resulting in the process continuing to run
// even after the parent process (this golang application) has exited.
func ExecCommandAndFork(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	err := cmd.Start()
	if err != nil {
		return err
	}
	return nil
}

// ExecCommand executes a command with the given arguments and returns its trimmed stdout.
// If the command fails, the returned error contains the content of stderr (and stdout
// holds whatever the command printed before failing).
func ExecCommand(command string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultCommandTimeout)
	defer cancel()
	return runner(ctx, command, args...)
}

// ExecCommandContext is like ExecCommand, but is bound to the given context.
func ExecCommandContext(ctx context.Context, command string, args ...string) (string, error) {
	return runner(ctx, command, args...)
}

func runCommand(ctx context.Context, command string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// children that inherited the pipes must not keep us waiting after a kill
	cmd.WaitDelay = 2 * time.Second

	err := cmd.Run()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			err = fmt.Errorf("timed out: %w", err)
		}
		// stdout is returned too, some tools (e.g. smartctl) report results with a non zero exit code
		return strings.TrimSpace(stdout.String()), fmt.Errorf("command '%s' failed: %w %s", cmd.String(), err, strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String()), nil
}
