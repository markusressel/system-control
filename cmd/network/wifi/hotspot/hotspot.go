package wifi

import (
	"fmt"
	"os"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/wifi"
	"github.com/spf13/cobra"
)

var (
	wifiInterface string
	hotspotSSID   string
	password      string
)

var Command = &cobra.Command{
	Use:   "hotspot",
	Short: "Control WiFi hotspots",
	Long: `Control WiFi hotspots.

Defaults for the interface, SSID and password can be set in the "hotspot" section of the
configuration file. Without any value the first WiFi device and the hostname are used and
NetworkManager generates a password.`,
}

func init() {
	Command.PersistentFlags().StringVarP(
		&wifiInterface,
		"interface", "i",
		"",
		"WiFi interface to use for the hotspot",
	)
	Command.PersistentFlags().StringVarP(
		&hotspotSSID,
		"ssid", "s",
		"",
		"SSID of the hotspot",
	)
	// former name of --ssid
	Command.PersistentFlags().StringVar(&hotspotSSID, "hotspotSSID", "", "SSID of the hotspot")
	_ = Command.PersistentFlags().MarkDeprecated("hotspotSSID", "use --ssid instead")
}

// hotspotSettings resolves interface and SSID from flags, configuration and detection.
func hotspotSettings() (iface string, ssid string, err error) {
	config := configuration.Get().Hotspot

	iface = firstNonEmpty(wifiInterface, config.Interface)
	if iface == "" {
		iface, err = wifi.GetWifiInterface()
		if err != nil {
			return "", "", err
		}
	}

	ssid = firstNonEmpty(hotspotSSID, config.SSID)
	if ssid == "" {
		ssid, err = os.Hostname()
		if err != nil {
			return "", "", err
		}
	}

	return iface, ssid, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// createHotspotConfigName returns the name of the NetworkManager connection of a hotspot.
func createHotspotConfigName(hotspotSSID string) string {
	return fmt.Sprintf("%s Hotspot", hotspotSSID)
}
