package redshift

import (
	"fmt"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var brightnessCmd = &cobra.Command{
	Use:   "brightness",
	Short: "Show current display brightness",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {

		displays, err := parseDisplayParam(display)
		if err != nil {
			return err
		}

		for _, display := range displays {
			if brightnessValue != -1 {
				newBrightness := clampBrightnessToConfig(brightnessValue)
				ApplyRedshift(display, -1, newBrightness, -1)
			}

			lastSetBrightness := getLastSetBrightness(display)
			fmt.Println(lastSetBrightness)
		}
		return nil
	},
}

func clampBrightnessToConfig(value float64) float64 {
	minBr := configuration.CurrentConfig.Redshift.Brightness.MinimumBrightness
	maxBr := configuration.CurrentConfig.Redshift.Brightness.MaximumBrightness
	newBrightnessValue := util.Clamp(value, minBr, maxBr)
	newBrightnessValue = util.RoundTo2DP(newBrightnessValue)
	return newBrightnessValue
}

func init() {
	brightnessCmd.Flags().Float64VarP(
		&brightnessValue,
		"brightness", "b",
		-1,
		"Brightness value to set (between 0.1 and 1.0)",
	)

	Command.AddCommand(brightnessCmd)
}
