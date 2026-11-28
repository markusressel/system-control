package bluetooth

import (
	"fmt"
	"sort"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/markusressel/system-control/internal/util"
)

func findBluetoothDeviceFuzzy(name string, devices []BluetoothDevice) []BluetoothDevice {
	// check exact address matches first
	for _, device := range devices {
		if util.EqualsIgnoreCase(device.Address, name) {
			return []BluetoothDevice{device}
		}
	}

	// then check fuzzy name matches
	deviceNames := make([]string, len(devices))
	for i, device := range devices {
		deviceNames[i] = device.Name
	}

	fuzzyMatches := fuzzy.RankFindNormalizedFold(name, deviceNames)
	sort.Sort(fuzzyMatches)

	result := make([]BluetoothDevice, 0)
	for _, match := range fuzzyMatches {
		for _, device := range devices {
			if device.Name == match.Target {
				result = append(result, device)
			}
		}
	}

	return result
}

func createDeviceNameList(devices []BluetoothDevice) []string {
	return util.MapFunc(devices, func(device BluetoothDevice) string {
		return fmt.Sprintf("%s (%s)", device.Name, device.Address)
	})
}

// resolveDevice finds the one device matching the given address or (fuzzy) name.
func resolveDevice(name string, devices []BluetoothDevice) (BluetoothDevice, error) {
	matchingDevices := findBluetoothDeviceFuzzy(name, devices)
	switch len(matchingDevices) {
	case 0:
		return BluetoothDevice{}, fmt.Errorf("device not found: %v", name)
	case 1:
		return matchingDevices[0], nil
	default:
		return BluetoothDevice{}, multipleDevicesError(matchingDevices)
	}
}

func multipleDevicesError(devices []BluetoothDevice) error {
	return fmt.Errorf("multiple matching devices found: %v", createDeviceNameList(devices))
}

// findDeviceByName finds the one known device matching the given address or (fuzzy) name.
func findDeviceByName(name string) (BluetoothDevice, error) {
	devices, err := GetBluetoothDevices()
	if err != nil {
		return BluetoothDevice{}, err
	}
	return resolveDevice(name, devices)
}

// ConnectByName connects to the one known device matching the given address or (fuzzy) name.
func ConnectByName(name string) error {
	device, err := findDeviceByName(name)
	if err != nil {
		return err
	}
	return ConnectToBluetoothDevice(device)
}

// DisconnectByName disconnects the one known device matching the given address or (fuzzy)
// name, or every connected device if name is empty.
func DisconnectByName(name string) error {
	if name == "" {
		return DisconnectAllBluetoothDevices()
	}

	device, err := findDeviceByName(name)
	if err != nil {
		return err
	}
	return DisconnectBluetoothDevice(device)
}

// RemoveByName removes the one known device matching the given address or (fuzzy) name.
func RemoveByName(name string) error {
	device, err := findDeviceByName(name)
	if err != nil {
		return err
	}
	return RemoveBluetoothDevice(device)
}
