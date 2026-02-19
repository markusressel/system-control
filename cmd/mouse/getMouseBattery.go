package mouse

import (
	"github.com/markusressel/system-control/internal/upower"
	"github.com/spf13/cobra"
)

var getMouseBatteryCmd = &cobra.Command{
	Use:   "battery",
	Short: "Get the current battery level",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		battery, err := GetMouseBattery()
		if err != nil {
			return err
		}
		cmd.Println(battery)
		return nil
	},
}

func GetMouseBattery() (string, error) {
	upowerDevices, err := upower.GetUpowerDevices()
	if err != nil {
		return "", err
	}
	for _, device := range upowerDevices {
		if device.Type == "mouse" {
			return device.Percentage, nil
		}
	}
	return "", nil
}

func init() {
	Command.AddCommand(getMouseBatteryCmd)
}
