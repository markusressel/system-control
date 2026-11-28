package media

import (
	"bufio"
	"context"
	"os/exec"
	"syscall"

	"github.com/markusressel/system-control/internal/supervise"
)

// followFormat is the format of a line of the follow command, see parseFollowLine.
const followFormat = "{{playerInstance}}\t{{status}}"

// followCommand returns the command printing a line whenever the status of any player changes.
// A variable for tests.
var followCommand = func(ctx context.Context) *exec.Cmd {
	return exec.CommandContext(ctx, "playerctl", "--all-players", "--follow", "--format", followFormat, "status")
}

// follower keeps the follow command running.
var follower = &supervise.Loop{Name: "media: following players"}

// FollowStatus describes the health of the follow command for `daemon status`.
func FollowStatus() []string {
	return follower.Status()
}

// Following reports whether changes are currently being followed, so that nothing has to poll.
func Following() bool {
	return follower.Active()
}

// Follow calls onChange whenever the status of a player changes (or one appears or goes
// away), until ctx is done. It records the status for LastPlayed. A follow command that ends
// is started again with a backoff.
func Follow(ctx context.Context, onChange func()) {
	follower.Run(ctx, func(ctx context.Context, ready func()) error {
		return followOnce(ctx, onChange, ready)
	}, onChange)
}

func followOnce(ctx context.Context, onChange func(), ready func()) error {
	cmd := followCommand(ctx)
	// its own process group: canceling kills children too, which would otherwise keep stdout
	// open and the scanner blocked
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGTERM, Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
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

	ready()
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if player, status, ok := parseFollowLine(scanner.Text()); ok {
			recordStatus(player, status)
		}
		onChange()
	}
	return scanner.Err()
}
