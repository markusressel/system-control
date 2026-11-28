package bluetooth

import (
	"time"

	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/spf13/cobra"
)

var scanDuration time.Duration

var bluetoothScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for available Bluetooth Devices",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return bluetooth.Scan(scanDuration, cmd.ErrOrStderr())
	},
}

func init() {
	Command.AddCommand(bluetoothScanCmd)

	bluetoothScanCmd.Flags().DurationVar(&scanDuration, "duration", 5*time.Second, "How long to scan for devices")
}
