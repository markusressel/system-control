package bluetooth

import (
	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/spf13/cobra"
)

var bluetoothRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a Bluetooth Device",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return bluetooth.RemoveByName(args[0])
	},
}

func init() {
	Command.AddCommand(bluetoothRemoveCmd)
}
