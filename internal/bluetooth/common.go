package bluetooth

import (
	"github.com/markusressel/system-control/internal/util"
	"strings"
)

type BluetoothDevice struct {
	Name    string
	Address string
}

func PairBluetoothDevice(device BluetoothDevice) error {
	_, err := util.ExecCommand(
		"bluetoothctl",
		"pair",
		device.Address,
	)
	return err
}

func ConnectToBluetoothDevice(device BluetoothDevice) error {
	_, err := util.ExecCommand(
		"bluetoothctl",
		"connect",
		device.Address,
	)
	return err
}

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
