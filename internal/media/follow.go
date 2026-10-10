package media

import (
	"bufio"
	"context"
	"log"
	"os/exec"
	"sync/atomic"
	"syscall"
	"time"
)

// followCommand returns the command printing a line whenever the status of a player changes.
// A variable for tests.
var followCommand = func(ctx context.Context) *exec.Cmd {
	return exec.CommandContext(ctx, "playerctl", "--follow", "status")
}

// restart delays of a follow command that ended, variables for tests
var (
	followRestartDelay = time.Second
	followMaxDelay     = 30 * time.Second
	followHealthyAfter = time.Minute
)

var following atomic.Bool

// Following reports whether changes are currently being followed, so that nothing has to poll.
func Following() bool {
	return following.Load()
}

// Follow calls onChange whenever the status of a player changes (or one appears or goes
// away), until ctx is done. A follow command that ends is started again with a backoff.
func Follow(ctx context.Context, onChange func()) {
	delay := followRestartDelay
	loggedStartFailure := false

	for ctx.Err() == nil {
		started := time.Now()
		err := followOnce(ctx, onChange)
		following.Store(false)
		if ctx.Err() != nil {
			return
		}

		if err != nil && !loggedStartFailure {
			log.Printf("media: can not follow players, status is polled: %v", err)
			loggedStartFailure = true
		}
		if time.Since(started) > followHealthyAfter {
			delay = followRestartDelay
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
		delay = min(delay*2, followMaxDelay)
	}
}

func followOnce(ctx context.Context, onChange func()) error {
	cmd := followCommand(ctx)
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGTERM}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	following.Store(true)
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		onChange()
	}
	return scanner.Err()
}
