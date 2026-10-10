package battery

import (
	"fmt"
	"strconv"

	batteryutil "github.com/markusressel/system-control/internal/battery"
	"github.com/markusressel/system-control/internal/persistence"
	"github.com/spf13/cobra"
)

var batteryChargingThresholdCmd = &cobra.Command{
	Use:   "threshold",
	Short: "Get/Set battery charging threshold",
	Long:  ``,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		battery, err := batteryutil.FindBattery(Name)
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
	},
}

func thresholdKey(battery batteryutil.BatteryInfo) string {
	return battery.Name + "_charge_control_end_threshold"
}

var batteryChargingThresholdSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save the current battery charging threshold on disk",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		battery, err := batteryutil.FindBattery(Name)
		if err != nil {
			return err
		}

		current, err := battery.ReadChargeControlEndThreshold()
		if err != nil {
			return err
		}

		return persistence.SaveInt(thresholdKey(battery), current)
	},
}

var batteryChargingThresholdRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore the last saved battery charging threshold",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		battery, err := batteryutil.FindBattery(Name)
		if err != nil {
			return err
		}

		value, err := persistence.ReadInt(thresholdKey(battery))
		if err != nil {
			return err
		}
		return battery.SetChargeControlEndThreshold(int(value))
	},
}

func init() {
	Command.AddCommand(batteryChargingThresholdCmd)

	batteryChargingThresholdCmd.AddCommand(batteryChargingThresholdSaveCmd)
	batteryChargingThresholdCmd.AddCommand(batteryChargingThresholdRestoreCmd)
}
