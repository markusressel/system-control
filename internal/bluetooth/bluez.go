package bluetooth

import (
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/google/uuid"
)

// getDeviceInfoFromBlueZ queries BlueZ over DBus for a device's properties and
// converts them into a BluetoothDevice struct. If DBus is not available or the
// device object does not exist, an error is returned.
func getDeviceInfoFromBlueZ(address string) (BluetoothDevice, error) {
	bus, err := dbus.SystemBus()
	if err != nil {
		return BluetoothDevice{}, fmt.Errorf("dbus: %w", err)
	}

	// BlueZ object path format: /org/bluez/hci0/dev_XX_XX_XX_XX_XX_XX
	objPath := dbus.ObjectPath("/org/bluez/hci0/dev_" + strings.ReplaceAll(address, ":", "_"))
	obj := bus.Object("org.bluez", objPath)

	var dev BluetoothDevice
	dev.UUIDs = make(map[string]uuid.UUID)

	// helper to read a property; returns false when property missing
	get := func(name string) (dbus.Variant, bool) {
		v, err := obj.GetProperty(name)
		if err != nil {
			return dbus.Variant{}, false
		}
		return v, true
	}

	if v, ok := get("org.bluez.Device1.Name"); ok {
		if s, ok := v.Value().(string); ok {
			dev.Name = s
		}
	}
	if v, ok := get("org.bluez.Device1.Alias"); ok {
		if s, ok := v.Value().(string); ok {
			dev.Alias = s
		}
	}
	if v, ok := get("org.bluez.Device1.Class"); ok {
		// Class is typically a uint32
		switch val := v.Value().(type) {
		case uint32:
			dev.Class = fmt.Sprintf("0x%08x", val)
		case int32:
			dev.Class = fmt.Sprintf("0x%08x", uint32(val))
		case int:
			dev.Class = fmt.Sprintf("0x%08x", uint32(val))
		}
	}
	if v, ok := get("org.bluez.Device1.Icon"); ok {
		if s, ok := v.Value().(string); ok {
			dev.Icon = s
		}
	}
	if v, ok := get("org.bluez.Device1.Paired"); ok {
		if b, ok := v.Value().(bool); ok {
			dev.Paired = b
		}
	}
	if v, ok := get("org.bluez.Device1.Bonded"); ok {
		if b, ok := v.Value().(bool); ok {
			dev.Bonded = b
		}
	}
	if v, ok := get("org.bluez.Device1.Trusted"); ok {
		if b, ok := v.Value().(bool); ok {
			dev.Trusted = b
		}
	}
	if v, ok := get("org.bluez.Device1.Blocked"); ok {
		if b, ok := v.Value().(bool); ok {
			dev.Blocked = b
		}
	}
	if v, ok := get("org.bluez.Device1.Connected"); ok {
		if b, ok := v.Value().(bool); ok {
			dev.Connected = b
		}
	}
	if v, ok := get("org.bluez.Device1.LegacyPairing"); ok {
		if b, ok := v.Value().(bool); ok {
			dev.LegacyPairing = b
		}
	}

	if v, ok := get("org.bluez.Device1.UUIDs"); ok {
		// Expecting []string
		slice, ok := v.Value().([]string)
		if !ok {
			// sometimes DBus may return []interface{}
			if si, ok := v.Value().([]interface{}); ok {
				for _, ii := range si {
					if s, ok := ii.(string); ok {
						if u, err := uuid.Parse(s); err == nil {
							dev.UUIDs[s] = u
						}
					}
				}
			}
		} else {
			for _, s := range slice {
				if u, err := uuid.Parse(s); err == nil {
					dev.UUIDs[s] = u
				}
			}
		}
	}

	// Battery percentage may be exposed via org.bluez.Battery1 on the same object
	if v, ok := get("org.bluez.Battery1.Percentage"); ok {
		switch val := v.Value().(type) {
		case uint8:
			p := int64(val)
			dev.BatteryPercentage = &p
		case uint16:
			p := int64(val)
			dev.BatteryPercentage = &p
		case int16:
			p := int64(val)
			dev.BatteryPercentage = &p
		case int:
			p := int64(val)
			dev.BatteryPercentage = &p
		}
	}

	// Ensure Address field is set
	dev.Address = address

	return dev, nil
}
