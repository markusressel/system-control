package battery

import (
	"context"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
)

func init() {
	daemon.Register(daemon.Op{
		Name:  "battery.remaining",
		Run:   runRemaining,
		Watch: daemon.Always,
		// the discharge rate is only known by asking for it
		Poll: configuration.PollInterval("battery", 10*time.Second),
	})
}

func runRemaining(ctx context.Context, a daemon.Args) ([]string, error) {
	remaining, err := Remaining(a.Flag(NameFlag))
	if err != nil {
		return nil, err
	}
	return []string{remaining}, nil
}
