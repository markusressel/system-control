package wifi

import (
	"github.com/markusressel/system-control/internal/wifi"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to a known WiFi network",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		networkName := args[0]
		return wifi.Connect(networkName)
	},
}

func init() {
	Command.AddCommand(connectCmd)
}
