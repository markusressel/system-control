package wifi

import (
	hotspot "github.com/markusressel/system-control/cmd/network/wifi/hotspot"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "wifi",
	Short: "Control WiFi devices and networks",
	Long:  ``,
}

func init() {
	Command.AddCommand(hotspot.Command)
}
