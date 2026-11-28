package battery

import (
	"fmt"
	"math"

	"github.com/markusressel/system-control/internal/util"
)

const (
	// NameFlag is the name of the flag of the `battery` commands that selects the battery.
	NameFlag = "name"
	// DefaultName is the battery the `battery` commands use without NameFlag.
	DefaultName = "BAT0"

	remainingTemplate = "%hours%:%minutes%"
)

// Remaining returns the time until the battery `name` is empty, or full while it is charging,
// as "hh:mm" (at most 99 hours), or "∞" if no power flows.
func Remaining(name string) (string, error) {
	batteryInfo, err := FindBattery(name)
	if err != nil {
		return "", err
	}

	charging, err := batteryInfo.IsCharging()
	if err != nil {
		return "", err
	}
	energyTarget, err := batteryInfo.GetEnergyTarget()
	if err != nil {
		return "", err
	}
	energyNow, err := batteryInfo.GetEnergyNow()
	if err != nil {
		return "", err
	}
	powerNow, err := batteryInfo.GetPowerNow()
	if err != nil {
		return "", err
	}

	if powerNow == 0 {
		return "∞", nil
	}

	var remainingTimeInSeconds int64
	if charging {
		remainingTimeInSeconds = CalculateRemainingTime(energyTarget-energyNow, powerNow)
	} else {
		remainingTimeInSeconds = CalculateRemainingTime(energyNow, powerNow)
	}
	remainingTimeInSeconds = max(0, remainingTimeInSeconds)

	remainingHours := int(math.Min(99, float64(remainingTimeInSeconds/60/60)))
	remainingMinutes := (remainingTimeInSeconds / 60) % 60

	return util.ReplacePlaceholders(remainingTemplate, map[string]string{
		"hours":   fmt.Sprintf("%02d", remainingHours),
		"minutes": fmt.Sprintf("%02d", remainingMinutes),
	}), nil
}
