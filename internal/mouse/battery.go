package mouse

import (
	"strings"
	"unicode"

	"github.com/markusressel/system-control/internal/upower"
	"github.com/markusressel/system-control/internal/util"
)

const (
	// BatteryTypeFlag is the name of the flag that selects the output format of `mouse battery`.
	BatteryTypeFlag = "type"
	// BatteryTypePercentage prints the battery level as reported, e.g. "85%".
	BatteryTypePercentage = "percentage"
	// BatteryTypeNumber prints only the number of the battery level.
	BatteryTypeNumber = "number"
)

// GetMouseBattery returns the battery level of the mouse, an empty string when there is no mouse.
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

// GetMouseBatteryNumber returns the battery level of the mouse as a plain number.
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
