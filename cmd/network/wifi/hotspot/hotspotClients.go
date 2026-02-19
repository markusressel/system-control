package wifi

import (
	"fmt"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/markusressel/system-control/internal/util"
	"github.com/markusressel/system-control/internal/wifi"
	"github.com/spf13/cobra"
)

var clientsCmd = &cobra.Command{
	Use:   "clients",
	Short: "List currently connected Hotspot Client Devices",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		isHotspotUp, err := wifi.IsHotspotUp(hotspotSSID)
		if err != nil {
			return err
		}
		if !isHotspotUp {
			fmt.Println("Hotspot is not running")
			return nil
		}

		hotspotDevices, err := wifi.GetConnectedHotspotDevices(wifiInterface, hotspotSSID)
		if err != nil {
			return err
		}

		for _, device := range hotspotDevices {
			printHotspotDevice(device)
		}

		return err
	},
}

func init() {
	Command.AddCommand(clientsCmd)

	clientsCmd.PersistentFlags().StringVarP(
		&hotspotSSID,
		"ssid", "s",
		"",
		"SSID of the hotspot",
	)
}

func printHotspotDevice(device wifi.HotspotLease) {
	properties := orderedmap.NewOrderedMap[string, string]()
	properties.Set("IP", device.IP)
	properties.Set("MAC", device.MAC)

	util.PrintFormattedTableOrdered(device.Name, properties)
}
