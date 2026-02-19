package wifi

import (
	"github.com/markusressel/system-control/internal/wifi"
	"github.com/spf13/cobra"
)

var offCmd = &cobra.Command{
	Use:   "off",
	Short: "Turn off the WiFi device",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return wifi.TurnOffWifiAdapter()
	},
}

func init() {
	Command.AddCommand(offCmd)
}
