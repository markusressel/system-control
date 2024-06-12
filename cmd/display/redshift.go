/*
 * system-control
 * Copyright (c) 2019. Markus Ressel
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
package display

import (
	"errors"
	"fmt"
	"github.com/markusressel/system-control/internal/persistence"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
	"strconv"
)

const (
	KeyRedshiftColorTemp  = "redshift.colorTemperature"
	KeyRedshiftBrightness = "redshift.brightness"
	KeyRedshiftGamma      = "redshift.gamma"
)

var (
	colorTemperature int64
	brightness       float64
	gamma            float64
)

var redshiftCmd = &cobra.Command{
	Use:   "redshift",
	Short: "Apply the given redshift",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {

		if colorTemperature != -1 && (colorTemperature < 1000 || colorTemperature > 25000) {
			return errors.New("color temperature must be between 1000 and 25000")
		}

		if brightness != -1 && (brightness < 0.1 || brightness > 1.0) {
			return errors.New("brightness must be between 0.1 and 1.0")
		}

		if gamma != -1 && (gamma < 0.1 || gamma > 2.0) {
			return errors.New("gamma must be between 0.1 and 2.0")
		}

		lastSetColorTemperature := getLastSetColorTemperature(display)
		lastSetBrightness := getLastSetBrightness(display)
		lastSetGamma := getLastSetGamma(display)

		if colorTemperature == -1 {
			colorTemperature = getLastSetColorTemperature(display)
		}
		if brightness == -1 {
			brightness = getLastSetBrightness(display)
		}
		if gamma == -1 {
			gamma = getLastSetGamma(display)
		}

		err := ApplyRedshift(display, colorTemperature, brightness, gamma)
		if err != nil {
			return err
		}

		// print current values
		fmt.Printf("Color Temperature: %d -> %d\n", lastSetColorTemperature, colorTemperature)
		fmt.Printf("Brightness: %.2f -> %.2f\n", lastSetBrightness, brightness)
		fmt.Printf("Gamma: %.2f -> %.2f\n", lastSetGamma, gamma)

		return nil
	},
}

func getLastSetColorTemperature(display string) int64 {
	key := KeyRedshiftColorTemp + "." + display
	lastSetColorTemperature, err := persistence.ReadInt(key)
	if err != nil {
		lastSetColorTemperature = -1
	}
	return lastSetColorTemperature
}

func getLastSetBrightness(display string) float64 {
	key := KeyRedshiftBrightness + "." + display
	lastSetBrightness, err := persistence.ReadFloat(key)
	if err != nil {
		lastSetBrightness = -1
	}
	if lastSetBrightness < 0.1 {
		lastSetBrightness = 0.1
		saveLastSetBrightness(display, lastSetBrightness)
	}
	if lastSetBrightness > 1.0 {
		lastSetBrightness = 1.0
		saveLastSetBrightness(display, lastSetBrightness)
	}
	return lastSetBrightness
}

func getLastSetGamma(display string) float64 {
	key := KeyRedshiftGamma + "." + display
	lastSetGamma, err := persistence.ReadFloat(key)
	if err != nil {
		lastSetGamma = -1
	}
	return lastSetGamma
}

func saveLastSetColorTemperature(display string, colorTemperature int64) error {
	key := KeyRedshiftColorTemp + "." + display
	return persistence.SaveInt(key, int(colorTemperature))
}

func saveLastSetBrightness(display string, brightness float64) error {
	key := KeyRedshiftBrightness + "." + display
	return persistence.SaveFloat(key, brightness)
}

func saveLastSetGamma(display string, gamma float64) error {
	key := KeyRedshiftGamma + "." + display
	return persistence.SaveFloat(key, gamma)
}

func ApplyRedshift(display string, colorTemperature int64, brightness float64, gamma float64) error {
	if colorTemperature == -1 {
		colorTemperature = getLastSetColorTemperature(display)
	}
	if brightness == -1 {
		brightness = getLastSetBrightness(display)
	}
	if gamma == -1 {
		gamma = getLastSetGamma(display)
	}

	displayIndex := -1
	if display == "DisplayPort-1" {
		displayIndex = 1
	} else if display == "DisplayPort-2" {
		displayIndex = 0
	}

	err := SetRedshiftCBG(displayIndex, colorTemperature, brightness, gamma)
	if err != nil {
		return err
	}

	err = saveLastSetColorTemperature(display, colorTemperature)
	if err != nil {
		return err
	}
	err = saveLastSetBrightness(display, brightness)
	if err != nil {
		return err
	}
	err = saveLastSetGamma(display, gamma)
	if err != nil {
		return err
	}

	return nil
}

// SetRedshiftCBG sets the redshift color temperature, brightness and gamma
// colorTemperature: the color temperature in Kelvin, between 1000 and 25000 (-1 to ignore, 6500 is default
// brightness: the brightness value between 0.1 and 1.0 (-1 to ignore, 1.0 is default)
// gamma: the gamma value between 0.1 and 1.0 (-1 to ignore, 1.0 is default)
// immediate: apply the changes immediately, without transition
func SetRedshiftCBG(display int, colorTemperature int64, brightness float64, gamma float64) error {
	args := []string{
		"-x", // reset previous "mode"
		"-P", // reset previous gamma ramps
		"-o", // one shot mode
	}

	if display > -1 {
		// -m randr:crtc=1
		args = append(args, "-m", fmt.Sprintf("randr:crtc=%d", display))
	}

	if colorTemperature != -1 {
		// set color temperature
		args = append(args, "-O", fmt.Sprintf("%d", colorTemperature))
	}

	if brightness != -1 {
		// set brightness
		args = append(args, "-b", strconv.FormatFloat(brightness, 'f', -1, 64))
	}

	if gamma != -1 {
		// set gamma
		args = append(args, "-g", strconv.FormatFloat(gamma, 'f', -1, 64))
	}

	if len(args) == 0 {
		return errors.New("no changes to apply")
	}
	_, err := util.ExecCommand("redshift", args...)
	return err
}

func ResetRedshift() error {
	args := []string{
		"-x", // reset previous "mode"
		"-P", // reset previous gamma ramps
		"-r", // apply immediately
	}
	_, err := util.ExecCommand("redshift", args...)
	return err
}

func init() {
	redshiftCmd.PersistentFlags().Int64VarP(
		&colorTemperature,
		"temperature", "t",
		-1,
		"Color Temperature",
	)

	redshiftCmd.PersistentFlags().Float64VarP(
		&brightness,
		"brightness", "b",
		-1,
		"Brightness",
	)

	redshiftCmd.PersistentFlags().Float64VarP(
		&gamma,
		"gamma", "g",
		-1,
		"Gamma",
	)

	Command.AddCommand(redshiftCmd)
}
