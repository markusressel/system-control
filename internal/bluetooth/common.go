package bluetooth

import (
	"github.com/markusressel/system-control/internal/util"
	"strings"
)

// BluetoothDevice represents a bluetooth device
type BluetoothDevice struct {
	// Name of the device
	Name string
	// Address MAC address of the device
	Address string
}

// PairBluetoothDevice pairs a bluetooth device with the system
func PairBluetoothDevice(device BluetoothDevice) error {
	_, err := util.ExecCommand(
		"bluetoothctl",
		"pair",
		device.Address,
	)
	return err
}

// ConnectToBluetoothDevice connects to a bluetooth device
func ConnectToBluetoothDevice(device BluetoothDevice) error {
	_, err := util.ExecCommand(
		"bluetoothctl",
		"connect",
		device.Address,
	)
	return err
}

// DisconnectAllBluetoothDevices disconnects all bluetooth devices
func DisconnectAllBluetoothDevices() error {
	_, err := util.ExecCommand(
		"bluetoothctl",
		"disconnect",
	)
	return err
}

// DisconnectBluetoothDevice disconnects a specific bluetooth device
func DisconnectBluetoothDevice(device BluetoothDevice) error {
	_, err := util.ExecCommand(
		"bluetoothctl",
		"disconnect",
		device.Address,
	)
	return err
}

// GetBluetoothDevices returns a list of all paired bluetooth devices
func GetBluetoothDevices() ([]BluetoothDevice, error) {
	result, err := util.ExecCommand(
		"bluetoothctl",
		"devices",
	)
	if err != nil {
		return nil, err
	}

	resultLines := strings.Split(result, "\n")
	devices := make([]BluetoothDevice, 0)
	for _, line := range resultLines {
		if strings.Contains(line, "Device ") {
			parts := strings.Split(line, " ")
			device := BluetoothDevice{
				Address: parts[1],
				Name:    parts[2],
			}
			devices = append(devices, device)
		}
	}

	return devices, nil
}
