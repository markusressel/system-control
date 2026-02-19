package bluetooth

import (
	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/spf13/cobra"
)

var bluetoothOffCmd = &cobra.Command{
	Use:   "off",
	Short: "Turn off the Bluetooth Adapter",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return bluetooth.TurnOffBluetoothAdapter()
	},
}

func init() {
	Command.AddCommand(bluetoothOffCmd)
}
