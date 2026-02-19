package network

import (
	"github.com/markusressel/system-control/cmd/network/device"
	"github.com/markusressel/system-control/cmd/network/wifi"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "network",
	Short: "Control Network Devices and Networks",
	Long:  ``,
}

func init() {
	Command.AddCommand(device.Command)
	Command.AddCommand(wifi.Command)
}
