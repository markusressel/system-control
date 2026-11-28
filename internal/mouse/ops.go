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
		Watch: daemon.Always,
		// a battery level changes slowly, and asking costs a D-Bus round trip
		Poll: configuration.PollInterval("mouse", time.Minute),
	})
}

func runBattery(ctx context.Context, a daemon.Args) ([]string, error) {
	var battery string
	var err error

	switch outputType := a.Flag(BatteryTypeFlag); outputType {
	case BatteryTypePercentage:
		battery, err = GetMouseBattery()
	case BatteryTypeNumber:
		battery, err = GetMouseBatteryNumber()
	default:
		return nil, fmt.Errorf("invalid value for --%s: %q (expected %q or %q)", BatteryTypeFlag, outputType, BatteryTypePercentage, BatteryTypeNumber)
	}
	if err != nil {
		return nil, err
	}

	// no mouse is an empty line, status bars depend on it
	return []string{battery}, nil
}
