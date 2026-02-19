package battery

import (
	"fmt"
	"strconv"

	"github.com/markusressel/system-control/internal/persistence"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var batteryChargingThresholdCmd = &cobra.Command{
	Use:   "threshold",
	Short: "Get/Set battery charging threshold",
	Long:  ``,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		batteryFlag := cmd.Flag("name")
		battery := batteryFlag.Value.String()

		if len(args) > 0 {
			newValueString := args[0]
			newValue, err := strconv.Atoi(newValueString)
			if err != nil {
				return err
			}
			err = setBatteryThreshold(battery, newValue)
			if err != nil {
				return err
			}
		} else {
			value, err := getBatteryThreshold(battery)
			if err != nil {
				return err
			}
			fmt.Println(value)
		}

		return nil
	},
}

var batteryChargingThresholdSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save the current battery charging threshold on disk",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		batteryFlag := cmd.Flag("name")
		battery := batteryFlag.Value.String()

		current, err := getBatteryThreshold(battery)
		if err != nil {
			return err
		}

		return persistence.SaveInt(battery+"_charge_control_end_threshold", current)
	},
}

var batteryChargingThresholdRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore the last saved battery charging threshold",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		batteryFlag := cmd.Flag("name")
		battery := batteryFlag.Value.String()

		value, err := persistence.ReadInt(battery + "_charge_control_end_threshold")
		if err != nil {
			return err
		}
		return setBatteryThreshold(battery, int(value))
	},
}

func setBatteryThreshold(battery string, value int) error {
	file := "/sys/class/power_supply/" + battery + "/charge_control_end_threshold"
	return util.WriteIntToFile(value, file)
}

func getBatteryThreshold(battery string) (int, error) {
	file := "/sys/class/power_supply/" + battery + "/charge_control_end_threshold"

	value, err := util.ReadIntFromFile(file)
	if err != nil {
		return -1, err
	}
	return int(value), err
}

func init() {
	Command.AddCommand(batteryChargingThresholdCmd)

	batteryChargingThresholdCmd.AddCommand(batteryChargingThresholdSaveCmd)
	batteryChargingThresholdCmd.AddCommand(batteryChargingThresholdRestoreCmd)
}
