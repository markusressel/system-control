package bluetooth

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var Name string

var Command = &cobra.Command{
	Use:              "bluetooth",
	Short:            "Control Bluetooth Devices",
	Long:             ``,
	TraverseChildren: true,
}

func init() {
	Command.PersistentFlags().StringVarP(
		&Name,
		"name", "n",
		"",
		"Device Name",
	)
}

func printBluetoothDevices(devices []bluetooth.BluetoothDevice) {
	for i, device := range devices {
		printBluetoothDevice(device)
		if i < len(devices)-1 {
			fmt.Println()
		}
	}
}

func printBluetoothDevice(device bluetooth.BluetoothDevice) {
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

func findBluetoothDeviceFuzzy(name string, devices []bluetooth.BluetoothDevice) []bluetooth.BluetoothDevice {
	// check exact address matches first
	for _, device := range devices {
		if util.EqualsIgnoreCase(device.Address, name) {
			return []bluetooth.BluetoothDevice{device}
		}
	}

	// then check fuzzy name matches
	deviceNames := make([]string, len(devices))
	for i, device := range devices {
		deviceNames[i] = device.Name
	}

	fuzzyMatches := fuzzy.RankFindNormalizedFold(name, deviceNames)
	sort.Sort(fuzzyMatches)

	result := make([]bluetooth.BluetoothDevice, 0)
	for _, match := range fuzzyMatches {
		for _, device := range devices {
			if device.Name == match.Target {
				result = append(result, device)
			}
		}
	}

	return result
}

func createDeviceNameList(devices []bluetooth.BluetoothDevice) []string {
	return util.MapFunc(devices, func(device bluetooth.BluetoothDevice) string {
		return fmt.Sprintf("%s (%s)", device.Name, device.Address)
	})
}
