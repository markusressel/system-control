package wifi

import (
	"github.com/markusressel/system-control/internal/wifi"
	"github.com/spf13/cobra"
)

var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect the current WiFi network",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return wifi.Disconnect()
	},
}

func init() {
	Command.AddCommand(disconnectCmd)
}
