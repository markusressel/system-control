package battery

import (
	batteryutil "github.com/markusressel/system-control/internal/battery"
	"github.com/spf13/cobra"
)

var batteryChargingThresholdCmd = &cobra.Command{
	Use:   "threshold",
	Short: "Get/Set battery charging threshold",
	Long:  ``,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return batteryutil.Threshold(Name, args)
	},
}

var batteryChargingThresholdSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save the current battery charging threshold on disk",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return batteryutil.SaveThreshold(Name)
	},
}

var batteryChargingThresholdRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore the last saved battery charging threshold",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return batteryutil.RestoreThreshold(Name)
	},
}

func init() {
	Command.AddCommand(batteryChargingThresholdCmd)

	batteryChargingThresholdCmd.AddCommand(batteryChargingThresholdSaveCmd)
	batteryChargingThresholdCmd.AddCommand(batteryChargingThresholdRestoreCmd)
}
