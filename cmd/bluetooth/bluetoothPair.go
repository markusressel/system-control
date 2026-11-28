package bluetooth

import (
	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/spf13/cobra"
)

var autoConnect bool
var removeExisting bool

var bluetoothPairCmd = &cobra.Command{
	Use:   "pair",
	Short: "Pair a Bluetooth Device",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return bluetooth.Pair(args[0], autoConnect, removeExisting)
	},
}

func init() {
	Command.AddCommand(bluetoothPairCmd)
	bluetoothPairCmd.Flags().BoolVarP(&autoConnect, "connect", "c", false, "Automatically connect after pairing")
	bluetoothPairCmd.Flags().BoolVarP(&removeExisting, "remove-existing", "r", false, "Remove existing device (if it exists) before pairing")
}
