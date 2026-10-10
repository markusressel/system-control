package redshift

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var colorTempCmd = &cobra.Command{
	Use:   "color-temperature",
	Short: "Show current display color temperature",
	Long:  ``,
	RunE:  commandutil.OpRunE("display.redshift.color-temperature"),
}

func clampColorTemperatureToConfig(colorTemperature int64) int64 {
	config := configuration.Get()
	minCT := config.Redshift.ColorTemperature.MinimumColorTemperature
	maxCT := config.Redshift.ColorTemperature.MaximumColorTemperature
	return util.Clamp(colorTemperature, minCT, maxCT)
}

func init() {
	colorTempCmd.Flags().BoolP("watch", "w", false, "Continuously watch for changes")

	colorTempCmd.Flags().Int64P(
		"temperature", "t",
		-1,
		"Color temperature to apply (between 1000 and 25000)")

	Command.AddCommand(colorTempCmd)
}
