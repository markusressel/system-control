package battery

import (
	"fmt"
	"math"

	batteryutil "github.com/markusressel/system-control/internal/battery"
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

		batteryInfo, err := batteryutil.FindBattery(Name)
		if err != nil {
			return err
		}

		// get value
		charging, err := batteryInfo.IsCharging()
		if err != nil {
			return err
		}
		energyTarget, err := batteryInfo.GetEnergyTarget()
		if err != nil {
			return err
		}
		energyNow, err := batteryInfo.GetEnergyNow()
		if err != nil {
			return err
		}
		powerNow, err := batteryInfo.GetPowerNow()
		if err != nil {
			return err
		}

		if powerNow == 0 {
			fmt.Println("∞")
			return nil
		}

		var remainingTimeInSeconds int64
		if charging == true {
			remainingTimeInSeconds = batteryutil.CalculateRemainingTime(energyTarget-energyNow, powerNow)
		} else {
			remainingTimeInSeconds = batteryutil.CalculateRemainingTime(energyNow, powerNow)
		}

		remainingTimeInSeconds = max(0, remainingTimeInSeconds)

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
