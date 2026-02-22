package bluetooth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/markusressel/system-control/internal/util"
)

var (
	ansiEscape    = regexp.MustCompile(`\x1b\[[0-9;]*[mGKHFJ]`)
	devicePattern = regexp.MustCompile(`Device ([0-9A-Fa-f:]{17})\s+(.+)`)
)

// execBluetoothCtl runs bluetoothctl in interactive mode, piping the given
// commands (plus "quit") via stdin, strips ANSI codes, and returns stdout.
func execBluetoothCtl(commands ...string) (string, error) {
	input := strings.Join(append(commands, "quit"), "\n")

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("bluetoothctl")
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("bluetoothctl: %w: %s", err, stderr.String())
	}
	return ansiEscape.ReplaceAllString(stdout.String(), ""), nil
}

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
	result, err := execBluetoothCtl("devices")
	if err != nil {
		return nil, err
	}

	return retrieveDevicesForResult(result)
}

// GetConnectedBluetoothDevices returns a list of all connected bluetooth devices
// bluetoothctl devices Connected
func GetConnectedBluetoothDevices() ([]BluetoothDevice, error) {
	result, err := execBluetoothCtl("devices Connected")
	if err != nil {
		return nil, err
	}

	return retrieveDevicesForResult(result)
}

// GetPairedBluetoothDevices returns a list of all paired bluetooth devices
func GetPairedBluetoothDevices() ([]BluetoothDevice, error) {
	result, err := execBluetoothCtl("devices Paired")
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
	devices := make([]simpleDeviceInfo, 0)
	for _, line := range strings.Split(result, "\n") {
		m := devicePattern.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		devices = append(devices, simpleDeviceInfo{
			Address: m[1],
			Name:    strings.TrimSpace(m[2]),
		})
	}
	return devices, nil
}

func GetBluetoothDeviceInfo(device simpleDeviceInfo) (BluetoothDevice, error) {
	result, err := execBluetoothCtl("info " + device.Address)
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
			m := devicePattern.FindStringSubmatch(line)
			if m != nil {
				result.Address = m[1]
			}
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
		} else if strings.Contains(line, "CONUUID: ") {
			indexOfTitle := strings.Index(line, "CONUUID: ")
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var input string
	if enable {
		input = "scan on\n" // no quit — let it run until killed by timeout
	} else {
		input = "scan off\nquit\n"
	}

	cmd := exec.CommandContext(ctx, "bluetoothctl")
	cmd.Stdin = strings.NewReader(input)
	err := cmd.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil
	}
	return err
}
