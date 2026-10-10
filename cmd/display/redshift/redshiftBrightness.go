package redshift

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var (
	brightnessFormatString string
)

var brightnessCmd = &cobra.Command{
	Use:   "brightness",
	Short: "Show current display brightness",
	Long:  ``,
	RunE:  commandutil.OpRunE("display.redshift.brightness"),
}

func clampBrightnessToConfig(value float64) float64 {
	config := configuration.Get()
	minBr := config.Redshift.Brightness.MinimumBrightness
	maxBr := config.Redshift.Brightness.MaximumBrightness
	newBrightnessValue := util.Clamp(value, minBr, maxBr)
	newBrightnessValue = util.RoundTo2DP(newBrightnessValue)
	return newBrightnessValue
}

func init() {
	brightnessCmd.Flags().BoolP("watch", "w", false, "Continuously watch for changes")

	brightnessCmd.Flags().Float64VarP(
		&brightnessValue,
		"brightness", "b",
		-1,
		"Brightness value to set (between 0.1 and 1.0)",
	)

	brightnessCmd.Flags().StringVarP(
		&brightnessFormatString,
		"format", "f",
		"%.2f",
		"Format string for the output value",
	)

	Command.AddCommand(brightnessCmd)
}
