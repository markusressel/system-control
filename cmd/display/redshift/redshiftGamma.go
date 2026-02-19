package redshift

import (
	"fmt"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var (
	gammaValue float64
)

var gammaCmd = &cobra.Command{
	Use:   "gamma",
	Short: "Show current display gamma",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		redshiftLock = lockRedshift()
		defer redshiftLock.Unlock()

		displays, err := parseDisplayParam(display)
		if err != nil {
			return err
		}

		for _, display := range displays {
			if gammaValue != -1 {
				newGamma := clampGammaToConfig(gammaValue)
				lastSetGamma := getLastSetGamma(display)
				if newGamma != lastSetGamma {
					err := ApplyRedshift(display, -1, -1, newGamma)
					if err != nil {
						return err
					}
				}
			}

			lastSetGamma := getLastSetGamma(display)
			fmt.Println(lastSetGamma)
		}
		return nil
	},
}

func clampGammaToConfig(value float64) float64 {
	minG := configuration.CurrentConfig.Redshift.Gamma.MinimumGamma
	maxG := configuration.CurrentConfig.Redshift.Gamma.MaximumGamma
	newGammaValue := util.Clamp(value, minG, maxG)
	newGammaValue = util.RoundTo2DP(newGammaValue)
	return newGammaValue
}

func init() {
	gammaCmd.Flags().Float64VarP(
		&gammaValue,
		"gamma", "g",
		-1,
		"Gamma value to set (between 0.1 and 2.0)",
	)

	Command.AddCommand(gammaCmd)
}
