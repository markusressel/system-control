package upower

import (
	"strings"

	"github.com/markusressel/system-control/internal/util"
)

/**
Example output of upower -d:

╰─λ upower -d                                                                                                                                                                                   0 (0.004s) < 20:02:32
Device: /org/freedesktop/UPower/devices/battery_hidpp_battery_0
  native-path:          hidpp_battery_0
  model:                Logitech MX Keys
  serial:               d8-dc-3a-a5
  power supply:         no
  updated:              Mon 05 May 2025 08:02:57 PM CEST (12 seconds ago)
  has history:          yes
  has statistics:       yes
  keyboard
    present:             yes
    rechargeable:        yes
    state:               discharging
    warning-level:       none
    battery-level:       normal
    percentage:          55% (should be ignored)
    icon-name:          'battery-low-symbolic'

Device: /org/freedesktop/UPower/devices/battery_hidpp_battery_1
  native-path:          hidpp_battery_1
  model:                Logitech G604
  serial:               dc-7d-18-8e
  power supply:         no
  updated:              Mon 05 May 2025 08:02:57 PM CEST (12 seconds ago)
  has history:          yes
  has statistics:       yes
  mouse
    present:             yes
    rechargeable:        yes
    state:               discharging
    warning-level:       none
    battery-level:       unknown
    percentage:          50% (should be ignored)
    icon-name:          'battery-caution-symbolic'

Device: /org/freedesktop/UPower/devices/DisplayDevice
  power supply:         no
  updated:              Sun 04 May 2025 04:34:42 PM CEST (98907 seconds ago)
  has history:          no
  has statistics:       no
  unknown
    warning-level:       none
    percentage:          0%
    icon-name:          'battery-missing-symbolic'

Daemon:
  daemon-version:  1.90.9
  on-battery:      no
  lid-is-closed:   no
  lid-is-present:  no
  critical-action: PowerOff

*/

type UpowerDevice struct {
	ID            string
	Type          string
	NativePath    string
	Model         string
	Serial        string
	PowerSupply   bool
	Updated       string
	HasHistory    bool
	HasStatistics bool
	Present       bool
	Rechargeable  bool
	State         string
	WarningLevel  string
	BatteryLevel  string
	Percentage    string
	IconName      string
}

func GetUpowerDevices() ([]UpowerDevice, error) {
	dump, err := util.ExecCommand("upower", "-d")
	if err != nil {
		return nil, err
	}

	devices, err := parseUpowerDump(dump)
	if err != nil {
		return nil, err
	}

	return devices, nil
}

func parseUpowerDump(dump string) ([]UpowerDevice, error) {
	lines := strings.Split(dump, "\n")
	var devices []UpowerDevice
	var currentDevice UpowerDevice

	for _, line := range lines {
		if strings.HasPrefix(line, "Device:") {
			if currentDevice.ID != "" {
				devices = append(devices, currentDevice)
			}
			currentDevice = UpowerDevice{
				ID: line[len("Device: "):],
			}
		} else if strings.HasPrefix(line, "  ") {
			keyValue := strings.SplitN(line[2:], ": ", 2)
			if len(keyValue) == 2 {
				key := strings.TrimSpace(keyValue[0])
				value := strings.TrimSpace(keyValue[1])
				switch key {
				case "native-path":
					currentDevice.NativePath = value
				case "model":
					currentDevice.Model = value
				case "serial":
					currentDevice.Serial = value
				case "power supply":
					currentDevice.PowerSupply = value == "yes"
				case "updated":
					currentDevice.Updated = value
				case "has history":
					currentDevice.HasHistory = value == "yes"
				case "has statistics":
					currentDevice.HasStatistics = value == "yes"
				case "present":
					currentDevice.Present = value == "yes"
				case "rechargeable":
					currentDevice.Rechargeable = value == "yes"
				case "state":
					currentDevice.State = value
				case "warning-level":
					currentDevice.WarningLevel = value
				case "battery-level":
					currentDevice.BatteryLevel = value
				case "percentage":
					currentDevice.Percentage = value
				case "icon-name":
					currentDevice.IconName = value
				default:
					// Ignore unknown keys
				}
			} else {
				// Handle cases where the line doesn't contain a colon
				// but is still relevant (e.g., "keyboard", "mouse", etc.)
				switch strings.TrimSpace(line) {
				case "keyboard":
					currentDevice.Type = "keyboard"
				case "mouse":
					currentDevice.Type = "mouse"
				default:
					currentDevice.Type = "unknown"
				}
			}
		}
	}

	// Add the last device if it exists
	if currentDevice.ID != "" {
		devices = append(devices, currentDevice)
	}

	return devices, nil
}
