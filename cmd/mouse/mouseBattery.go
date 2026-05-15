package mouse

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/markusressel/system-control/internal/upower"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

const (
	mouseBatteryTypeFlag       = "type"
	mouseBatteryTypePercentage = "percentage"
	mouseBatteryTypeNumber     = "number"
)

var getMouseBatteryCmd = &cobra.Command{
	Use:   "battery",
	Short: "Get the current battery level",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		outputType, err := cmd.Flags().GetString(mouseBatteryTypeFlag)
		if err != nil {
			return err
		}

		var battery string
		switch outputType {
		case mouseBatteryTypePercentage:
			battery, err = GetMouseBattery()
		case mouseBatteryTypeNumber:
			battery, err = GetMouseBatteryNumber()
		default:
			return fmt.Errorf("invalid value for --%s: %q (expected %q or %q)", mouseBatteryTypeFlag, outputType, mouseBatteryTypePercentage, mouseBatteryTypeNumber)
		}

		if err != nil {
			return err
		}
		cmd.Println(battery)
		return nil
	},
}

func GetMouseBattery() (string, error) {
	battery, err := util.GetMouseBatteryViaDBus()
	if err == nil && battery != "" {
		return battery, nil
	}

	upowerDevices, err := upower.GetUpowerDevices()
	if err != nil {
		if battery != "" {
			return battery, nil
		}
		return "", err
	}
	for _, device := range upowerDevices {
		if device.Type == "mouse" {
			return device.Percentage, nil
		}
	}
	return "", nil
}

func GetMouseBatteryNumber() (string, error) {
	battery, err := GetMouseBattery()
	if err != nil {
		return "", err
	}

	return extractBatteryNumber(battery), nil
}

func extractBatteryNumber(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimSuffix(trimmed, "%")

	numberRunes := make([]rune, 0, len(trimmed))
	for _, r := range trimmed {
		if unicode.IsDigit(r) || r == '.' {
			numberRunes = append(numberRunes, r)
			continue
		}
		if len(numberRunes) > 0 {
			break
		}
	}

	if len(numberRunes) > 0 {
		return string(numberRunes)
	}

	return trimmed
}

func init() {
	getMouseBatteryCmd.Flags().String(mouseBatteryTypeFlag, mouseBatteryTypePercentage, "output format: percentage or number")
	Command.AddCommand(getMouseBatteryCmd)
}
