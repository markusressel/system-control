package battery

import (
	"fmt"
	"math"

	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var batteryRemainingCmd = &cobra.Command{
	Use:   "remaining",
	Short: "Get the remaining battery life in hours and minutes",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		template := "%hours%:%minutes%"

		batteries, err := util.GetBatteryList()
		if err != nil {
			return err
		}

		if len(batteries) <= 0 {
			return fmt.Errorf("no batteries found")
		}

		var batteryInfo *util.BatteryInfo
		for _, battery := range batteries {
			if util.EqualsIgnoreCase(battery.Name, Name) {
				batteryInfo = &battery
			} else if util.EqualsIgnoreCase(battery.SerialNumber, Name) {
				batteryInfo = &battery
			} else if util.EqualsIgnoreCase(battery.Model, Name) {
				batteryInfo = &battery
			} else if util.EqualsIgnoreCase(battery.Path, Name) {
				batteryInfo = &battery
			}
		}

		if batteryInfo == nil {
			return fmt.Errorf("no battery found matching '%s'", Name)
		}

		batteryInfoNonNull := *batteryInfo

		// get value
		charging, err := batteryInfoNonNull.IsCharging()
		if err != nil {
			return err
		}
		energyTarget, err := batteryInfoNonNull.GetEnergyTarget()
		if err != nil {
			return err
		}
		energyNow, err := batteryInfoNonNull.GetEnergyNow()
		if err != nil {
			return err
		}
		powerNow, err := batteryInfoNonNull.GetPowerNow()
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}

		if powerNow == 0 {
			fmt.Println("∞")
			return nil
		}

		var remainingTimeInSeconds int64
		if charging == true {
			remainingTimeInSeconds = util.CalculateRemainingTime(energyTarget-energyNow, powerNow)
		} else {
			remainingTimeInSeconds = util.CalculateRemainingTime(energyNow, powerNow)
		}

		remainingHours := int(math.Min(99, float64(remainingTimeInSeconds/60/60)))
		remainingMinutes := (remainingTimeInSeconds / 60) % 60

		placeholders := map[string]string{}
		placeholders["hours"] = fmt.Sprintf("%02d", remainingHours)
		placeholders["minutes"] = fmt.Sprintf("%02d", remainingMinutes)

		result := util.ReplacePlaceholders(template, placeholders)
		fmt.Println(result)

		return nil
	},
}

func init() {
	Command.AddCommand(batteryRemainingCmd)
}
