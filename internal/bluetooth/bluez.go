package bluetooth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/google/uuid"
)

const (
	bluezService     = "org.bluez"
	adapterInterface = "org.bluez.Adapter1"
	deviceInterface  = "org.bluez.Device1"
	batteryInterface = "org.bluez.Battery1"
)

// managedObjects is the result of ObjectManager.GetManagedObjects: object path -> interface -> property.
type managedObjects map[dbus.ObjectPath]map[string]map[string]dbus.Variant

// ErrNoAdapter is returned if BlueZ is running, but there is no bluetooth adapter.
var ErrNoAdapter = errors.New("no bluetooth adapter found")

// getManagedObjects queries everything BlueZ knows with a single call.
func getManagedObjects(bus *dbus.Conn) (managedObjects, error) {
	var objects managedObjects
	err := bus.Object(bluezService, "/").
		Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).
		Store(&objects)
	if err != nil {
		return nil, fmt.Errorf("could not query bluez: %w", err)
	}
	return objects, nil
}

// defaultAdapterPath returns the path of the first adapter.
func defaultAdapterPath(objects managedObjects) (dbus.ObjectPath, error) {
	var found dbus.ObjectPath
	for path, interfaces := range objects {
		if _, ok := interfaces[adapterInterface]; ok {
			// map iteration is random, pick the lowest path to be stable (hci0 before hci1)
			if found == "" || path < found {
				found = path
			}
		}
	}
	if found == "" {
		return "", ErrNoAdapter
	}
	return found, nil
}

// deviceFromInterfaces converts the BlueZ interfaces of a device object into a BluetoothDevice.
func deviceFromInterfaces(interfaces map[string]map[string]dbus.Variant) (BluetoothDevice, bool) {
	props, ok := interfaces[deviceInterface]
	if !ok {
		return BluetoothDevice{}, false
	}

	dev := BluetoothDevice{UUIDs: map[string]uuid.UUID{}}
	dev.Address = stringProp(props, "Address")
	dev.Name = stringProp(props, "Name")
	dev.Alias = stringProp(props, "Alias")
	if dev.Name == "" {
		dev.Name = dev.Alias
	}
	dev.Icon = stringProp(props, "Icon")
	if class, ok := props["Class"].Value().(uint32); ok {
		dev.Class = fmt.Sprintf("0x%08x", class)
	}
	dev.Paired = boolProp(props, "Paired")
	dev.Bonded = boolProp(props, "Bonded")
	dev.Trusted = boolProp(props, "Trusted")
	dev.Blocked = boolProp(props, "Blocked")
	dev.Connected = boolProp(props, "Connected")
	dev.LegacyPairing = boolProp(props, "LegacyPairing")

	if uuids, ok := props["UUIDs"].Value().([]string); ok {
		for _, s := range uuids {
			if u, err := uuid.Parse(s); err == nil {
				dev.UUIDs[s] = u
			}
		}
	}

	if battery, ok := interfaces[batteryInterface]; ok {
		if p, ok := battery["Percentage"].Value().(byte); ok {
			percentage := int64(p)
			dev.BatteryPercentage = &percentage
		}
	}

	return dev, true
}

func stringProp(props map[string]dbus.Variant, name string) string {
	s, _ := props[name].Value().(string)
	return s
}

func boolProp(props map[string]dbus.Variant, name string) bool {
	b, _ := props[name].Value().(bool)
	return b
}

// devices returns all known devices together with their object paths.
func devices(objects managedObjects) map[dbus.ObjectPath]BluetoothDevice {
	result := map[dbus.ObjectPath]BluetoothDevice{}
	for path, interfaces := range objects {
		if dev, ok := deviceFromInterfaces(interfaces); ok {
			result[path] = dev
		}
	}
	return result
}

// findDevicePath looks up the object path of the device with the given address.
func findDevicePath(objects managedObjects, address string) (dbus.ObjectPath, error) {
	for path, dev := range devices(objects) {
		if strings.EqualFold(dev.Address, address) {
			return path, nil
		}
	}
	return "", fmt.Errorf("device not found: %s", address)
}
