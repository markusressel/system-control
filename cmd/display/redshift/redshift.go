package redshift

import (
	"errors"
	"fmt"
	internalredshift "github.com/markusressel/system-control/internal/redshift"

	"github.com/markusressel/system-control/internal/daemon"
	"github.com/spf13/cobra"
)

var (
	display string

	colorTemperature int64
	brightness       float64
	gamma            float64
)

var Command = &cobra.Command{
	Use:   "redshift",
	Short: "Apply the given redshift",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		unlock, err := daemon.AcquireLock(internalredshift.LockName)
		if err != nil {
			return err
		}
		defer unlock()

		if colorTemperature != -1 && (colorTemperature < internalredshift.MinColorTemperature || colorTemperature > internalredshift.MaxColorTemperature) {
			return errors.New("color temperature must be between 1000 and 25000")
		}

		if brightness != -1 && (brightness < 0.1 || brightness > 1.0) {
			return errors.New("brightness must be between 0.1 and 1.0")
		}

		if gamma != -1 && (gamma < 0.1 || gamma > 2.0) {
			return errors.New("gamma must be between 0.1 and 2.0")
		}

		displays, err := internalredshift.ParseDisplays(display)
		if err != nil {
			return err
		}

		for _, display := range displays {

			lastSetColorTemperature := internalredshift.LastColorTemperature(display)
			lastSetBrightness := internalredshift.LastBrightness(display)
			lastSetGamma := internalredshift.LastGamma(display)

			err := internalredshift.Apply(display, colorTemperature, brightness, gamma)
			if err != nil {
				return err
			}

			// print current values
			fmt.Printf("Display: %s\n", display)
			if colorTemperature != -1 {
				fmt.Printf("  Color Temperature: %d -> %d\n", lastSetColorTemperature, colorTemperature)
			} else {
				fmt.Printf("  Color Temperature: %d\n", lastSetColorTemperature)
			}
			if brightness != -1 {
				fmt.Printf("  Brightness: %.2f -> %.2f\n", lastSetBrightness, brightness)
			} else {
				fmt.Printf("  Brightness: %.2f\n", lastSetBrightness)
			}
			if gamma != -1 {
				fmt.Printf("  Gamma: %.2f -> %.2f\n", lastSetGamma, gamma)
			} else {
				fmt.Printf("  Gamma: %.2f\n", lastSetGamma)
			}
		}

		return nil
	},
}

func init() {
	Command.PersistentFlags().StringVarP(
		&display,
		"display", "d",
		"",
		"Display",
	)

	Command.PersistentFlags().Int64VarP(
		&colorTemperature,
		"temperature", "t",
		-1,
		"Color Temperature",
	)

	Command.PersistentFlags().Float64VarP(
		&brightness,
		"brightness", "b",
		-1,
		"Brightness",
	)

	Command.PersistentFlags().Float64VarP(
		&gamma,
		"gamma", "g",
		-1,
		"Gamma",
	)
}
