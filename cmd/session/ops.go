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
		Watch: func(daemon.Args) bool { return true },
		// the screen locker is started by anything, there is no event
		Poll: func() time.Duration {
			if configured := configuration.Get().Daemon.PollIntervals["session"]; configured > 0 {
				return configured
			}
			return 2 * time.Second
		},
	})
}

func runLocked(ctx context.Context, a daemon.Args) ([]string, error) {
	running, err := isProcessRunning("i3lock")
	if err != nil {
		return nil, err
	}
	if running {
		return []string{"yes"}, nil
	}
	return []string{"no"}, nil
}
