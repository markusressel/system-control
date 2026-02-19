package wifi

import (
	"github.com/markusressel/system-control/internal/wifi"
	"github.com/spf13/cobra"
)

var onCmd = &cobra.Command{
	Use:   "on",
	Short: "Turn on the WiFi device",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return wifi.TurnOnWifiAdapter()
	},
}

func init() {
	Command.AddCommand(onCmd)
}
