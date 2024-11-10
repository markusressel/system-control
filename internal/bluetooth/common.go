package bluetooth

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/markusressel/system-control/internal/util"
	"strconv"
	"strings"
	"time"
)

// TurnOnBluetoothAdapter turns on the bluetooth adapter
func TurnOnBluetoothAdapter() error {
	_, err := util.ExecCommand("bluetoothctl", "power", "on")
	return err
}

// TurnOffBluetoothAdapter turns off the bluetooth adapter
func TurnOffBluetoothAdapter() error {
	_, err := util.ExecCommand("bluetoothctl", "power", "off")
	return err
}

// simpleDeviceInfo is a simple representation of a known bluetooth device and only used internally.
// BluetoothDevice is a much more detailed representation.
type simpleDeviceInfo struct {
	// Name of the device
	Name string
	// Address MAC address of the device
	Address string
}

// BluetoothDevice represents a bluetooth device
type BluetoothDevice struct {
	Name              string // LG-TONE-FP9
	Address           string // B8:F8:BE:13:A4:72
	Alias             string // LG-TONE-FP9
	Class             string // 0x00240404 (2360324)
	Icon              string // audio-headset
	Paired            bool   // yes
	Bonded            bool   // yes
	Trusted           bool   // yes
	Blocked           bool   // no
	Connected         bool   // yes
	LegacyPairing     bool   // no
	UUIDs             map[string]uuid.UUID
	BatteryPercentage *int64 // 0x4b (75)
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

	return retrieveDevicesForResult(result)
}

// GetConnectedBluetoothDevices returns a list of all connected bluetooth devices
// bluetoothctl devices Connected
func GetConnectedBluetoothDevices() ([]BluetoothDevice, error) {
	result, err := util.ExecCommand(
		"bluetoothctl",
		"devices",
		"Connected",
	)
	if err != nil {
		return nil, err
	}

	return retrieveDevicesForResult(result)
}

// GetPairedBluetoothDevices returns a list of all paired bluetooth devices
func GetPairedBluetoothDevices() ([]BluetoothDevice, error) {
	result, err := util.ExecCommand(
		"bluetoothctl",
		"devices",
		"Paired",
	)
	if err != nil {
		return nil, err
	}

	return retrieveDevicesForResult(result)
}

func RemoveBluetoothDevice(device BluetoothDevice) error {
	_, err := util.ExecCommand(
		"bluetoothctl",
		"remove",
		device.Address,
	)
	return err
}

func getFullInfoDevices(simpleDeviceInfos []simpleDeviceInfo) (devices []BluetoothDevice, err error) {
	for _, deviceId := range simpleDeviceInfos {
		info, err := GetBluetoothDeviceInfo(deviceId)
		if err != nil {
			return nil, err
		}
		devices = append(devices, info)
	}
	return devices, err
}

func parseSimpleDeviceList(result string) ([]simpleDeviceInfo, error) {
	resultLines := strings.Split(result, "\n")
	devices := make([]simpleDeviceInfo, 0)
	for _, line := range resultLines {
		if strings.Contains(line, "Device ") {
			parts := strings.Split(line, " ")
			device := simpleDeviceInfo{
				Address: parts[1],
				Name:    parts[2],
			}
			devices = append(devices, device)
		}
	}

	return devices, nil
}

func GetBluetoothDeviceInfo(device simpleDeviceInfo) (BluetoothDevice, error) {
	result, err := util.ExecCommand(
		"bluetoothctl",
		"info",
		device.Address,
	)
	if err != nil {
		return BluetoothDevice{}, err
	}

	return parseBluetoothDeviceInfo(result)
}

// parseBluetoothDeviceInfo parses the output of "bluetoothctl info <address>"
func parseBluetoothDeviceInfo(input string) (result BluetoothDevice, err error) {
	lines := strings.Split(input, "\n")

	result.UUIDs = make(map[string]uuid.UUID)

	for _, line := range lines {
		if strings.Contains(line, "Device ") {
			parts := strings.Split(input, " ")
			result.Address = parts[1]
		} else if strings.Contains(line, "Name: ") {
			result.Name = strings.TrimSpace(strings.ReplaceAll(line, "Name: ", ""))
		} else if strings.Contains(line, "Alias: ") {
			result.Alias = strings.TrimSpace(strings.ReplaceAll(line, "Alias: ", ""))
		} else if strings.Contains(line, "Class: ") {
			result.Class = strings.TrimSpace(strings.ReplaceAll(line, "Class: ", ""))
		} else if strings.Contains(line, "Icon: ") {
			result.Icon = strings.TrimSpace(strings.ReplaceAll(line, "Icon: ", ""))
		} else if strings.Contains(line, "Paired: ") {
			result.Paired = strings.Contains(line, "yes")
		} else if strings.Contains(line, "Bonded: ") {
			result.Bonded = strings.Contains(line, "yes")
		} else if strings.Contains(line, "Trusted: ") {
			result.Trusted = strings.Contains(line, "yes")
		} else if strings.Contains(line, "Blocked: ") {
			result.Blocked = strings.Contains(line, "yes")
		} else if strings.Contains(line, "Connected: ") {
			result.Connected = strings.Contains(line, "yes")
		} else if strings.Contains(line, "LegacyPairing: ") {
			result.LegacyPairing = strings.Contains(line, "yes")
		} else if strings.Contains(line, "UUID: ") {
			indexOfTitle := strings.Index(line, "UUID: ")
			value := strings.TrimSpace(line[indexOfTitle+5:])
			indexOfLastBracket := strings.LastIndex(value, "(")
			key := strings.TrimSpace(value[:indexOfLastBracket])
			uuidString := strings.TrimSpace(value[indexOfLastBracket+1 : len(value)-1])

			uuidValue, err := uuid.Parse(uuidString)
			if err != nil {
				return result, err
			}
			result.UUIDs[key] = uuidValue
		} else if strings.Contains(line, "Battery Percentage: ") {
			parts := strings.Split(line, " ")
			batteryPercentageString := parts[2]
			batteryPercentage, err := strconv.ParseInt(batteryPercentageString, 0, 64)
			if err != nil {
				return result, err
			}
			result.BatteryPercentage = &batteryPercentage
		}
	}

	return result, nil
}

func retrieveDevicesForResult(result string) ([]BluetoothDevice, error) {
	deviceIdList, err := parseSimpleDeviceList(result)
	if err != nil {
		return nil, err
	}

	return getFullInfoDevices(deviceIdList)
}

// SetBluetoothScan enables or disables bluetooth scanning
func SetBluetoothScan(enable bool) error {
	arg := "off"
	if enable {
		arg = "on"
	}

	err := util.ExecCommandOneshot(
		5*time.Second,
		"bluetoothctl",
		"scan",
		arg,
	)
	if errors.Is(err, context.DeadlineExceeded) {
		return nil
	}
	return err
}
