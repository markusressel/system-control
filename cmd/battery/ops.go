package battery

import (
	"context"
	"fmt"
	"math"
	"time"

	batteryutil "github.com/markusressel/system-control/internal/battery"
	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
	"github.com/markusressel/system-control/internal/util"
)

const remainingTemplate = "%hours%:%minutes%"

func init() {
	daemon.Register(daemon.Op{
		Name:  "battery.remaining",
		Run:   runRemaining,
		Watch: func(daemon.Args) bool { return true },
		// the discharge rate is only known by asking for it
		Poll: func() time.Duration {
			if configured := configuration.Get().Daemon.PollIntervals["battery"]; configured > 0 {
				return configured
			}
			return 10 * time.Second
		},
	})
}

func runRemaining(ctx context.Context, a daemon.Args) ([]string, error) {
	batteryInfo, err := batteryutil.FindBattery(a.Flag("name"))
	if err != nil {
		return nil, err
	}

	charging, err := batteryInfo.IsCharging()
	if err != nil {
		return nil, err
	}
	energyTarget, err := batteryInfo.GetEnergyTarget()
	if err != nil {
		return nil, err
	}
	energyNow, err := batteryInfo.GetEnergyNow()
	if err != nil {
		return nil, err
	}
	powerNow, err := batteryInfo.GetPowerNow()
	if err != nil {
		return nil, err
	}

	if powerNow == 0 {
		return []string{"∞"}, nil
	}

	var remainingTimeInSeconds int64
	if charging {
		remainingTimeInSeconds = batteryutil.CalculateRemainingTime(energyTarget-energyNow, powerNow)
	} else {
		remainingTimeInSeconds = batteryutil.CalculateRemainingTime(energyNow, powerNow)
	}
	remainingTimeInSeconds = max(0, remainingTimeInSeconds)

	remainingHours := int(math.Min(99, float64(remainingTimeInSeconds/60/60)))
	remainingMinutes := (remainingTimeInSeconds / 60) % 60

	return []string{util.ReplacePlaceholders(remainingTemplate, map[string]string{
		"hours":   fmt.Sprintf("%02d", remainingHours),
		"minutes": fmt.Sprintf("%02d", remainingMinutes),
	})}, nil
}
