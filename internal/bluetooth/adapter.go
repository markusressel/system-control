package bluetooth

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/godbus/dbus/v5"
)

func connect() (*dbus.Conn, managedObjects, error) {
	bus, err := dbus.SystemBus()
	if err != nil {
		return nil, nil, fmt.Errorf("dbus: %w", err)
	}
	objects, err := getManagedObjects(bus)
	return bus, objects, err
}

func setAdapterPowered(powered bool) error {
	bus, objects, err := connect()
	if err != nil {
		return err
	}
	path, err := defaultAdapterPath(objects)
	if err != nil {
		return err
	}
	err = bus.Object(bluezService, path).
		Call("org.freedesktop.DBus.Properties.Set", 0, adapterInterface, "Powered", dbus.MakeVariant(powered)).Err
	if err != nil {
		return fmt.Errorf("failed to set Powered: %w", err)
	}
	return nil
}

func TurnOnBluetoothAdapter() error {
	return setAdapterPowered(true)
}

func TurnOffBluetoothAdapter() error {
	return setAdapterPowered(false)
}

// SetBluetoothScan starts or stops discovery of nearby devices. BlueZ stops discovery
// when the process that started it exits, so a scan only lasts as long as the caller runs.
func SetBluetoothScan(enable bool) error {
	bus, objects, err := connect()
	if err != nil {
		return err
	}
	path, err := defaultAdapterPath(objects)
	if err != nil {
		return err
	}

	method := adapterInterface + ".StopDiscovery"
	if enable {
		method = adapterInterface + ".StartDiscovery"
	}
	err = bus.Object(bluezService, path).Call(method, 0).Err
	var dbusErr dbus.Error
	if errors.As(err, &dbusErr) &&
		(dbusErr.Name == "org.bluez.Error.InProgress" || dbusErr.Name == "org.bluez.Error.Failed" && !enable) {
		// already in the requested state
		return nil
	}
	return err
}

// GetBluetoothDevices returns all devices known to BlueZ, sorted by name.
func GetBluetoothDevices() (BluetoothDeviceList, error) {
	_, objects, err := connect()
	if err != nil {
		return nil, err
	}
	result := BluetoothDeviceList{}
	for _, dev := range devices(objects) {
		result = append(result, dev)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name != result[j].Name {
			return result[i].Name < result[j].Name
		}
		return result[i].Address < result[j].Address
	})
	return result, nil
}

func callDeviceMethod(device BluetoothDevice, method string) error {
	bus, objects, err := connect()
	if err != nil {
		return err
	}
	path, err := findDevicePath(objects, device.Address)
	if err != nil {
		return err
	}
	if err := bus.Object(bluezService, path).Call(deviceInterface+"."+method, 0).Err; err != nil {
		return fmt.Errorf("%s failed: %w", method, err)
	}
	return nil
}

func PairBluetoothDevice(device BluetoothDevice) error {
	return callDeviceMethod(device, "Pair")
}

func ConnectToBluetoothDevice(device BluetoothDevice) error {
	return callDeviceMethod(device, "Connect")
}

func DisconnectBluetoothDevice(device BluetoothDevice) error {
	return callDeviceMethod(device, "Disconnect")
}

// DisconnectAllBluetoothDevices disconnects all connected devices. A failure to disconnect
// one device does not prevent disconnecting the others, the errors are combined.
func DisconnectAllBluetoothDevices() error {
	bus, objects, err := connect()
	if err != nil {
		return err
	}
	var failed []string
	for path, dev := range devices(objects) {
		if !dev.Connected {
			continue
		}
		if err := bus.Object(bluezService, path).Call(deviceInterface+".Disconnect", 0).Err; err != nil {
			failed = append(failed, fmt.Sprintf("%s (%s): %v", dev.Name, dev.Address, err))
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("failed to disconnect: %s", strings.Join(failed, "; "))
	}
	return nil
}

// RemoveBluetoothDevice unpairs the device and removes it from BlueZ.
func RemoveBluetoothDevice(device BluetoothDevice) error {
	bus, objects, err := connect()
	if err != nil {
		return err
	}
	path, err := findDevicePath(objects, device.Address)
	if err != nil {
		return err
	}
	adapterPath, err := defaultAdapterPath(objects)
	if err != nil {
		return err
	}
	err = bus.Object(bluezService, adapterPath).Call(adapterInterface+".RemoveDevice", 0, path).Err
	if err != nil {
		return fmt.Errorf("RemoveDevice failed: %w", err)
	}
	return nil
}
