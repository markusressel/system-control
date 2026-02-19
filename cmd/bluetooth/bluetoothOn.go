package bluetooth

import (
	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/spf13/cobra"
)

var bluetoothOnCmd = &cobra.Command{
	Use:   "on",
	Short: "Turn on the Bluetooth Adapter",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return bluetooth.TurnOnBluetoothAdapter()
	},
}

func init() {
	Command.AddCommand(bluetoothOnCmd)
}
