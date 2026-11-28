package battery

import (
	"fmt"
	"strconv"

	"github.com/markusressel/system-control/internal/persistence"
)

// Threshold prints the charging threshold of the battery `name`, or sets it to `args[0]` if given.
func Threshold(name string, args []string) error {
	battery, err := FindBattery(name)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		newValue, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		return battery.SetChargeControlEndThreshold(newValue)
	}

	value, err := battery.ReadChargeControlEndThreshold()
	if err != nil {
		return err
	}
	fmt.Println(value)
	return nil
}

func thresholdKey(battery BatteryInfo) string {
	return battery.Name + "_charge_control_end_threshold"
}

// SaveThreshold saves the current charging threshold of the battery `name` on disk.
func SaveThreshold(name string) error {
	battery, err := FindBattery(name)
	if err != nil {
		return err
	}

	current, err := battery.ReadChargeControlEndThreshold()
	if err != nil {
		return err
	}

	return persistence.SaveInt(thresholdKey(battery), current)
}

// RestoreThreshold sets the charging threshold of the battery `name` to the value SaveThreshold saved.
func RestoreThreshold(name string) error {
	battery, err := FindBattery(name)
	if err != nil {
		return err
	}

	value, err := persistence.ReadInt(thresholdKey(battery))
	if err != nil {
		return err
	}
	return battery.SetChargeControlEndThreshold(int(value))
}
