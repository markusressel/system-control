package bluetooth

import (
	"sort"
	"strconv"

	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var filterConnected bool
var filterPaired bool

var bluetoothDevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List all known Bluetooth Devices",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		devices, err := bluetooth.GetBluetoothDevices()
		if err != nil {
			return err
		}

		// sort by name
		sort.Sort(devices)

		filteredDevices := util.FilterFunc(devices, func(device bluetooth.BluetoothDevice) bool {
			if filterConnected && !device.Connected {
				return false
			}
			if filterPaired && !device.Paired {
				return false
			}
			return true
		})

		printBluetoothDevices(filteredDevices)

		return nil
	},
}

func parseFlagAsBool(cmd *cobra.Command, flagName string, defaultValue bool) (flagValue bool, err error) {
	connectedFlag := cmd.Flag(flagName)
	flagValue = defaultValue
	if connectedFlag != nil {
		connectedFlagValue := connectedFlag.Value.String()
		if connectedFlagValue != "" {
			flagValue, err = strconv.ParseBool(connectedFlagValue)
			if err != nil {
				return flagValue, err
			}
		}
	}
	return flagValue, nil
}

func init() {
	Command.AddCommand(bluetoothDevicesCmd)

	bluetoothDevicesCmd.Flags().BoolVarP(&filterConnected, "connected", "c", false, "Filter by connected state")
	bluetoothDevicesCmd.Flags().BoolVarP(&filterPaired, "paired", "p", false, "Filter by paired state")
}
