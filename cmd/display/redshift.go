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
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
	"strconv"
)

var colorTemperature int
var brightness float64
var gamma float64

var redshiftCmd = &cobra.Command{
	Use:   "redshift",
	Short: "Apply the given redshift",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {

		if colorTemperature != -1 && (colorTemperature < 1000 || colorTemperature > 25000) {
			return errors.New("Color temperature must be between 1000 and 25000")
		}

		if brightness < 0.1 || brightness > 1.0 {
			return errors.New("Brightness must be between 0.1 and 1.0")
		}

		err := SetRedshiftCBG(colorTemperature, brightness, gamma)
		if err != nil {
			return err
		}

		return nil
	},
}

// SetRedshiftCBG sets the redshift color temperature, brightness and gamma
// colorTemperature: the color temperature in Kelvin, between 1000 and 25000 (-1 to ignore)
// brightness: the brightness value between 0.1 and 1.0 (-1 to ignore)
// gamma: the gamma value between 0.1 and 1.0 (-1 to ignore)
// immediate: apply the changes immediately, without transition
func SetRedshiftCBG(colorTemperature int, brightness float64, gamma float64) error {
	args := []string{
		"-x", // reset previous "mode"
		"-P", // reset previous gamma ramps
		"-o", // one shot mode
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
	redshiftCmd.PersistentFlags().IntVarP(
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
