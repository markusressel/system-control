package wifi

import (
	"fmt"

	"github.com/markusressel/system-control/internal/wifi"
	"github.com/spf13/cobra"
)

var wifiInterface string
var hotspotSSID string
var password string

var onCmd = &cobra.Command{
	Use:   "on",
	Short: "Turn on the WiFi Hotspot",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if wifiInterface == "" {
			wifiInterface = "wlo1"
		}
		if hotspotSSID == "" {
			hotspotSSID = "M16"
		}
		if password == "" {
			password = "M12345678"
		}
		hotspotName := createHotspotConfigName(hotspotSSID)

		err := wifi.TurnOnHotspot(
			hotspotName,
			wifiInterface,
			hotspotSSID,
			password,
		)
		return err
	},
}

func createHotspotConfigName(hotspotSSID string) string {
	return fmt.Sprintf("%s Hotspot", hotspotSSID)
}

func init() {
	Command.AddCommand(onCmd)

	onCmd.PersistentFlags().StringVarP(
		&hotspotSSID,
		"hotspotSSID", "s",
		"",
		"SSID of the hotspot",
	)
	onCmd.PersistentFlags().StringVarP(
		&password,
		"password", "p",
		"",
		"Password of the hotspot",
	)
}
