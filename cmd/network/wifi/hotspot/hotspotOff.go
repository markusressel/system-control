package wifi

import (
	"errors"

	"github.com/markusressel/system-control/internal/wifi"
	"github.com/spf13/cobra"
)

var force bool

var offCmd = &cobra.Command{
	Use:   "off",
	Short: "Turn off the WiFi Hotspot",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		iface, ssid, err := hotspotSettings()
		if err != nil {
			return err
		}
		hotspotName := createHotspotConfigName(ssid)

		if !force {
			// check if there are still devices connected
			hotspotDevices, err := wifi.GetConnectedHotspotDevices(iface, hotspotName)
			if err != nil {
				return err
			}

			if len(hotspotDevices) > 0 {
				for _, device := range hotspotDevices {
					printHotspotDevice(device)
				}
				return errors.New("there are still devices connected to the hotspot, disconnect them first or use --force")
			}
		}

		return wifi.TurnOffHotspot(hotspotName)
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
}
