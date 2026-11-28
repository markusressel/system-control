package bluetooth

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"time"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/markusressel/system-control/internal/util"
)

func printBluetoothDevices(devices []BluetoothDevice) {
	for i, device := range devices {
		printBluetoothDevice(device)
		if i < len(devices)-1 {
			fmt.Println()
		}
	}
}

func printBluetoothDevice(device BluetoothDevice) {
	properties := orderedmap.NewOrderedMap[string, string]()
	properties.Set("Address", device.Address)
	properties.Set("Connected", strconv.FormatBool(device.Connected))
	properties.Set("Paired", strconv.FormatBool(device.Paired))
	properties.Set("Bonded", strconv.FormatBool(device.Bonded))
	properties.Set("Trusted", strconv.FormatBool(device.Trusted))
	properties.Set("Blocked", strconv.FormatBool(device.Blocked))

	if device.BatteryPercentage != nil {
		properties.Set("Battery", fmt.Sprintf("%v%%", *device.BatteryPercentage))
	}

	util.PrintFormattedTableOrdered(device.Name, properties)
}

// PrintDevices prints all known devices sorted by name, only the connected and/or paired
// ones if requested.
func PrintDevices(onlyConnected bool, onlyPaired bool) error {
	devices, err := GetBluetoothDevices()
	if err != nil {
		return err
	}

	// sort by name
	sort.Sort(devices)

	filteredDevices := util.FilterFunc(devices, func(device BluetoothDevice) bool {
		if onlyConnected && !device.Connected {
			return false
		}
		if onlyPaired && !device.Paired {
			return false
		}
		return true
	})

	printBluetoothDevices(filteredDevices)

	return nil
}

// Scan scans for devices for the given duration and prints the ones that are neither
// connected nor paired. Progress messages go to stderr.
func Scan(duration time.Duration, stderr io.Writer) (err error) {
	err = SetBluetoothScan(true)
	if err != nil {
		return err
	}
	defer func() {
		if scanErr := SetBluetoothScan(false); scanErr != nil {
			fmt.Fprintln(stderr, "Failed to turn off Bluetooth scanning")
		}
	}()

	fmt.Fprintln(stderr, "Scanning for Bluetooth Devices...")
	time.Sleep(duration)

	devices, err := GetBluetoothDevices()
	if err != nil {
		return err
	}
	availableDevices := util.FilterFunc(devices, func(device BluetoothDevice) bool {
		return !device.Connected && !device.Paired
	})

	printBluetoothDevices(availableDevices)

	return nil
}

// Pair scans for the device with the exact name or address `deviceName` for up to 30 seconds
// and pairs it. removeExisting removes a known device of that name first, autoConnect
// connects to it after pairing.
func Pair(deviceName string, autoConnect bool, removeExisting bool) error {
	if removeExisting {
		devices, err := GetBluetoothDevices()
		if err != nil {
			return err
		}
		for _, device := range devices {
			if device.Name == deviceName || device.Address == deviceName {
				err := RemoveBluetoothDevice(device)
				if err != nil {
					return fmt.Errorf("failed to remove existing device %s: %v", deviceName, err)
				}
			}
		}
	}

	defer func() {
		err := SetBluetoothScan(false)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	err := SetBluetoothScan(true)
	if err != nil {
		return fmt.Errorf("failed to start Bluetooth scan: %v", err)
	}

	startTime := time.Now()
	// search for device in loop with timeout
	for time.Now().Sub(startTime) < 30*time.Second {
		devices, err := GetBluetoothDevices()
		if err != nil {
			return err
		}
		for _, device := range devices {
			if device.Name == deviceName || device.Address == deviceName {
				err := PairBluetoothDevice(device)

				if err != nil && autoConnect {
					fmt.Printf("Failed to pair device %s: %v\n", deviceName, err)
				} else if autoConnect {
					err = ConnectToBluetoothDevice(device)
					if err != nil {
						fmt.Printf("Failed to connect to device %s after pairing: %v\n", deviceName, err)
					}
				}
				return err
			}
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("device not found: %v", deviceName)
}
