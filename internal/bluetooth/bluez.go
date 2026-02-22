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

// listAllDevicesFromBlueZ returns a list of all Device1 objects known to BlueZ,
// enriched via getDeviceInfoFromBlueZ. It uses the ObjectManager to discover
// device object paths and then collects their Address property.
func listAllDevicesFromBlueZ() ([]BluetoothDevice, error) {
	bus, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("dbus: %w", err)
	}
	obj := bus.Object("org.bluez", "/")

	var managedObjects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	if err := obj.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&managedObjects); err != nil {
		return nil, fmt.Errorf("GetManagedObjects: %w", err)
	}

	devices := make([]BluetoothDevice, 0)
	for _, ifaces := range managedObjects {
		if props, ok := ifaces["org.bluez.Device1"]; ok {
			// Extract Address property
			if v, ok := props["Address"]; ok {
				if addr, ok := v.Value().(string); ok {
					// get full device info
					dev, err := getDeviceInfoFromBlueZ(addr)
					if err != nil {
						// If enrichment fails, build minimal device
						dev = BluetoothDevice{Address: addr}
					}
					// ensure Name when missing, try to build from object path or Alias
					if dev.Name == "" {
						// try Alias
						if v2, ok2 := props["Alias"]; ok2 {
							if s, ok3 := v2.Value().(string); ok3 {
								dev.Name = s
							}
						}
					}
					devices = append(devices, dev)
				}
			}
		}
	}

	return devices, nil
}

// getDefaultAdapterPath finds the first Adapter1 object path via ObjectManager.
func getDefaultAdapterPath() (dbus.ObjectPath, error) {
	bus, err := dbus.SystemBus()
	if err != nil {
		return "", fmt.Errorf("dbus: %w", err)
	}
	obj := bus.Object("org.bluez", "/")
	var managedObjects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	if err := obj.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&managedObjects); err != nil {
		return "", fmt.Errorf("GetManagedObjects: %w", err)
	}
	for path, ifaces := range managedObjects {
		if _, ok := ifaces["org.bluez.Adapter1"]; ok {
			return path, nil
		}
	}
	return "", fmt.Errorf("no bluez adapter found")
}

// setAdapterPowered sets the Powered property on the default adapter.
func setAdapterPowered(powered bool) error {
	path, err := getDefaultAdapterPath()
	if err != nil {
		return err
	}
	bus, err := dbus.SystemBus()
	if err != nil {
		return fmt.Errorf("dbus: %w", err)
	}
	obj := bus.Object("org.bluez", path)
	// Use org.freedesktop.DBus.Properties.Set
	if err := obj.Call("org.freedesktop.DBus.Properties.Set", 0, "org.bluez.Adapter1", "Powered", dbus.MakeVariant(powered)).Err; err != nil {
		return fmt.Errorf("failed to set Powered: %w", err)
	}
	return nil
}

// removeDeviceByPath removes a device by its object path using Adapter1.RemoveDevice
func removeDeviceByPath(devicePath dbus.ObjectPath) error {
	adapterPath, err := getDefaultAdapterPath()
	if err != nil {
		return err
	}
	bus, err := dbus.SystemBus()
	if err != nil {
		return fmt.Errorf("dbus: %w", err)
	}
	adapterObj := bus.Object("org.bluez", adapterPath)
	if err := adapterObj.Call("org.bluez.Adapter1.RemoveDevice", 0, devicePath).Err; err != nil {
		return fmt.Errorf("RemoveDevice failed: %w", err)
	}
	return nil
}

// findDevicePath looks up the DBus object path for a given device address using ObjectManager.
func findDevicePath(address string) (dbus.ObjectPath, error) {
	bus, err := dbus.SystemBus()
	if err != nil {
		return "", fmt.Errorf("dbus: %w", err)
	}
	obj := bus.Object("org.bluez", "/")

	var managedObjects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	if err := obj.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&managedObjects); err != nil {
		return "", fmt.Errorf("GetManagedObjects: %w", err)
	}

	for path, ifaces := range managedObjects {
		if props, ok := ifaces["org.bluez.Device1"]; ok {
			if v, ok := props["Address"]; ok {
				if addr, ok := v.Value().(string); ok {
					if strings.EqualFold(addr, address) {
						return path, nil
					}
				}
			}
		}
	}

	// As a fallback, try constructing the common path using hci0 and probe it
	constructed := dbus.ObjectPath("/org/bluez/hci0/dev_" + strings.ReplaceAll(address, ":", "_"))
	constructedObj := bus.Object("org.bluez", constructed)
	// Probe by attempting to read the Address property
	if _, err := constructedObj.GetProperty("org.bluez.Device1.Address"); err == nil {
		return constructed, nil
	}

	return "", fmt.Errorf("device not found: %s", address)
}

// callDeviceMethod invokes a method on the Device1 interface for the given device path.
func callDeviceMethod(devicePath dbus.ObjectPath, method string) error {
	bus, err := dbus.SystemBus()
	if err != nil {
		return fmt.Errorf("dbus: %w", err)
	}
	devObj := bus.Object("org.bluez", devicePath)
	if err := devObj.Call("org.bluez.Device1."+method, 0).Err; err != nil {
		return fmt.Errorf("Device1.%s failed: %w", method, err)
	}
	return nil
}

// listDevicesMatching allows filtering devices from ObjectManager using a predicate.
func listDevicesMatching(pred func(BluetoothDevice) bool) ([]BluetoothDevice, error) {
	all, err := listAllDevicesFromBlueZ()
	if err != nil {
		return nil, err
	}
	filtered := make([]BluetoothDevice, 0)
	for _, d := range all {
		if pred(d) {
			filtered = append(filtered, d)
		}
	}
	return filtered, nil
}
