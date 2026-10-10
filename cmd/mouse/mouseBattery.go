package mouse

import (
	"strings"
	"unicode"

	"github.com/markusressel/system-control/internal/commandutil"
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
	RunE:  commandutil.OpRunE("mouse.battery"),
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
	getMouseBatteryCmd.Flags().BoolP("watch", "w", false, "Continuously watch for changes")
	getMouseBatteryCmd.Flags().String(mouseBatteryTypeFlag, mouseBatteryTypePercentage, "output format: percentage or number")
	Command.AddCommand(getMouseBatteryCmd)
}
