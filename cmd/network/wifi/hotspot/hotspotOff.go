package wifi

import (
	"os"

	"github.com/markusressel/system-control/internal/wifi"
	"github.com/spf13/cobra"
)

var force bool

var offCmd = &cobra.Command{
	Use:   "off",
	Short: "Turn off the WiFi Hotspot",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if wifiInterface == "" {
			wifiInterface = "wlo1"
		}
		if hotspotSSID == "" {
			hotspotSSID = "M16"
		}

		hotspotName := createHotspotConfigName(hotspotSSID)

		if !force {
			// check if there are still devices connected
			hotspotDevices, err := wifi.GetConnectedHotspotDevices(wifiInterface, hotspotSSID)
			if err != nil {
				return err
			}

			if len(hotspotDevices) > 0 {
				for _, device := range hotspotDevices {
					printHotspotDevice(device)
				}
				println("There are still devices connected to the hotspot. Please disconnect them first or use --force parameter.")
				os.Exit(1)
			}
		}

		err := wifi.TurnOffHotspot(hotspotName)
		return err
	},
}

func init() {
	Command.AddCommand(offCmd)

	offCmd.Flags().BoolVarP(
		&force,
		"force", "f",
		false,
		"Force turn off the Hotspot",
	)

	offCmd.PersistentFlags().StringVarP(
		&hotspotSSID,
		"ssid", "s",
		"",
		"SSID of the hotspot",
	)
}
