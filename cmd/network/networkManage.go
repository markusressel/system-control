package network

import (
	"github.com/markusressel/system-control/internal/wifi"
	"github.com/spf13/cobra"
)

var manageCmd = &cobra.Command{
	Use:   "manage",
	Short: "Manage WiFi devices and networks",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return wifi.OpenManageGui()
	},
}

func init() {
	Command.AddCommand(manageCmd)
}
