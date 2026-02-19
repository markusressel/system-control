package redshift

import (
	"fmt"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var (
	colorTempValue int64
)

var colorTempCmd = &cobra.Command{
	Use:   "color-temperature",
	Short: "Show current display color temperature",
	Long:  ``,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if colorTempValue != -1 && (colorTempValue < 1000 || colorTempValue > 25000) {
			return fmt.Errorf("color temperature must be between 1000 and 25000")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		redshiftLock = lockRedshift()
		defer redshiftLock.Unlock()

		displays, err := parseDisplayParam(display)
		if err != nil {
			return err
		}

		for _, display := range displays {
			if len(displays) > 1 {
				fmt.Printf("Display: %s\n", display)
			}

			if colorTempValue != -1 {
				newColorTemperature := clampColorTemperatureToConfig(colorTempValue)
				lastSetColorTemperature := getLastSetColorTemperature(display)
				if newColorTemperature != lastSetColorTemperature {
					err := ApplyRedshift(display, newColorTemperature, -1, -1)
					if err != nil {
						return err
					}
				}
			}

			lastSetColorTemperature := getLastSetColorTemperature(display)
			fmt.Println(lastSetColorTemperature)
		}
		return nil
	},
}

func clampColorTemperatureToConfig(colorTemperature int64) int64 {
	config := configuration.CurrentConfig
	minCT := config.Redshift.ColorTemperature.MinimumColorTemperature
	maxCT := config.Redshift.ColorTemperature.MaximumColorTemperature
	return util.Clamp(colorTemperature, minCT, maxCT)
}

func init() {
	colorTempCmd.Flags().Int64VarP(
		&colorTempValue,
		"temperature", "t",
		-1,
		"Color temperature to apply (between 1000 and 25000)")

	Command.AddCommand(colorTempCmd)
}
