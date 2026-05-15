package util

import (
	"fmt"
	"math"

	"github.com/godbus/dbus/v5"
)

const (
	uPowerBusName         = "org.freedesktop.UPower"
	uPowerObjectPath      = "/org/freedesktop/UPower"
	uPowerInterface       = "org.freedesktop.UPower"
	uPowerDeviceInterface = "org.freedesktop.UPower.Device"
	uPowerMouseType       = 5
)

func GetMouseBatteryViaDBus() (string, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return "", err
	}

	upowerObject := conn.Object(uPowerBusName, dbus.ObjectPath(uPowerObjectPath))
	var devicePaths []dbus.ObjectPath
	if err := upowerObject.Call(uPowerInterface+".EnumerateDevices", 0).Store(&devicePaths); err != nil {
		return "", err
	}

	for _, devicePath := range devicePaths {
		deviceObject := conn.Object(uPowerBusName, devicePath)

		var deviceTypeVariant dbus.Variant
		typeCall := deviceObject.Call("org.freedesktop.DBus.Properties.Get", 0, uPowerDeviceInterface, "Type")
		if typeCall.Err != nil {
			continue
		}
		if err := typeCall.Store(&deviceTypeVariant); err != nil {
			continue
		}

		if extractUPowerDeviceType(deviceTypeVariant) != uPowerMouseType {
			continue
		}

		var percentageVariant dbus.Variant
		percentageCall := deviceObject.Call("org.freedesktop.DBus.Properties.Get", 0, uPowerDeviceInterface, "Percentage")
		if percentageCall.Err != nil {
			continue
		}
		if err := percentageCall.Store(&percentageVariant); err != nil {
			continue
		}

		percentage, ok := percentageVariant.Value().(float64)
		if !ok {
			continue
		}

		return formatBatteryPercentage(percentage), nil
	}

	return "", nil
}

func extractUPowerDeviceType(deviceTypeVariant dbus.Variant) int {
	switch value := deviceTypeVariant.Value().(type) {
	case uint32:
		return int(value)
	case int32:
		return int(value)
	case uint64:
		return int(value)
	case int64:
		return int(value)
	default:
		return -1
	}
}

func formatBatteryPercentage(percentage float64) string {
	if percentage == math.Trunc(percentage) {
		return fmt.Sprintf("%.0f%%", percentage)
	}

	return fmt.Sprintf("%.1f%%", percentage)
}
