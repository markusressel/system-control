package redshift

import (
	"github.com/markusressel/system-control/internal/commandutil"
	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var gammaCmd = &cobra.Command{
	Use:   "gamma",
	Short: "Show current display gamma",
	Long:  ``,
	RunE:  commandutil.OpRunE("display.redshift.gamma"),
}

func clampGammaToConfig(value float64) float64 {
	config := configuration.Get()
	minG := config.Redshift.Gamma.MinimumGamma
	maxG := config.Redshift.Gamma.MaximumGamma
	newGammaValue := util.Clamp(value, minG, maxG)
	newGammaValue = util.RoundTo2DP(newGammaValue)
	return newGammaValue
}

func init() {
	gammaCmd.Flags().BoolP("watch", "w", false, "Continuously watch for changes")

	gammaCmd.Flags().Float64P(
		"gamma", "g",
		-1,
		"Gamma value to set (between 0.1 and 2.0)",
	)

	gammaCmd.Flags().StringP(
		"format", "f",
		"%.2f",
		"Format string for the output value",
	)

	Command.AddCommand(gammaCmd)
}
