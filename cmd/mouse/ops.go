package mouse

import (
	"context"
	"fmt"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
)

func init() {
	daemon.Register(daemon.Op{
		Name:  "mouse.battery",
		Run:   runBattery,
		Watch: func(daemon.Args) bool { return true },
		// a battery level changes slowly, and asking costs a D-Bus round trip
		Poll: func() time.Duration {
			if configured := configuration.Get().Daemon.PollIntervals["mouse"]; configured > 0 {
				return configured
			}
			return time.Minute
		},
	})
}

func runBattery(ctx context.Context, a daemon.Args) ([]string, error) {
	var battery string
	var err error

	switch outputType := a.Flag(mouseBatteryTypeFlag); outputType {
	case mouseBatteryTypePercentage:
		battery, err = GetMouseBattery()
	case mouseBatteryTypeNumber:
		battery, err = GetMouseBatteryNumber()
	default:
		return nil, fmt.Errorf("invalid value for --%s: %q (expected %q or %q)", mouseBatteryTypeFlag, outputType, mouseBatteryTypePercentage, mouseBatteryTypeNumber)
	}
	if err != nil {
		return nil, err
	}

	// no mouse is an empty line, status bars depend on it
	return []string{battery}, nil
}
