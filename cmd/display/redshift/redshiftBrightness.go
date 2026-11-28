package redshift

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/spf13/cobra"
)

var brightnessCmd = &cobra.Command{
	Use:   "brightness",
	Short: "Show current display brightness",
	Long:  ``,
	RunE:  commandutil.OpRunE("display.redshift.brightness"),
}

func init() {
	commandutil.AddWatchFlag(brightnessCmd)

	brightnessCmd.Flags().Float64P(
		"brightness", "b",
		-1,
		"Brightness value to set (between 0.1 and 1.0)",
	)

	brightnessCmd.Flags().StringP(
		"format", "f",
		"%.2f",
		"Format string for the output value",
	)

	Command.AddCommand(brightnessCmd)
}
