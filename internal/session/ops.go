package session

import (
	"context"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
)

func init() {
	daemon.Register(daemon.Op{
		Name:  "session.locked",
		Run:   runLocked,
		Watch: daemon.Always,
		// the screen locker is started by anything, there is no event
		Poll: configuration.PollInterval("session", 2*time.Second),
	})
}

func runLocked(ctx context.Context, a daemon.Args) ([]string, error) {
	running, err := isProcessRunning(currentLockProcessName())
	if err != nil {
		return nil, err
	}
	if running {
		return []string{"yes"}, nil
	}
	return []string{"no"}, nil
}
