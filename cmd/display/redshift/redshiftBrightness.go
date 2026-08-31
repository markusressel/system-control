package redshift

import (
	"fmt"

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
	RunE: commandutil.StateCommandRunE("display.redshift.brightness", func(cmd *cobra.Command) bool {
		return brightnessValue != -1
	}, func(cmd *cobra.Command) (string, error) {
		redshiftLock = lockRedshift()
		defer redshiftLock.Unlock()

		displays, err := parseDisplayParam(display)
		if err != nil {
			return "", err
		}

		var lastSet float64
		for _, d := range displays {
			if brightnessValue != -1 {
				newBrightness := clampBrightnessToConfig(brightnessValue)
				lastSetBrightness := getLastSetBrightness(d)
				if newBrightness != lastSetBrightness {
					err := ApplyRedshift(d, -1, newBrightness, -1)
					if err != nil {
						return "", err
					}
				}
			}

			lastSet = getLastSetBrightness(d)
		}
		return fmt.Sprintf(brightnessFormatString, lastSet), nil
	}),
}

func clampBrightnessToConfig(value float64) float64 {
	minBr := configuration.CurrentConfig.Redshift.Brightness.MinimumBrightness
	maxBr := configuration.CurrentConfig.Redshift.Brightness.MaximumBrightness
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
