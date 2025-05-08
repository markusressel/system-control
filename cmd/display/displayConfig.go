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
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/pflag"

	"github.com/spf13/cobra"
)

var displays []string
var modes []string
var positions []string
var rate []int
var primary []bool
var off []bool
var auto []bool

var displayConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Set the desired display configuration",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		displays, err := util.GetDisplays()
		if err != nil {
			return err
		}

		//{% if hostvars[inventory_hostname].device_type == "desktop" %}
		//balias display-left-4k              "xrandr --output DisplayPort-1 --mode 3840x2160 --output DisplayPort-2 --primary --pos 3840x0"
		//balias display-left-2k              "xrandr --output DisplayPort-1 --mode 2560x1440 --output DisplayPort-2 --primary --pos 2560x0"
		//
		//{% elif ansible_hostname == "Alien" %}
		//balias display-external-hdmi-only   "xrandr --output LVDS1 --off  --output HDMI1 --auto --output DP1 --off  --output VGA1 --off"
		//balias display-external-dp-only     "xrandr --output LVDS1 --off  --output HDMI1 --off  --output DP1 --auto --output VGA1 --off"
		//balias display-external-vga-only    "xrandr --output LVDS1 --off  --output HDMI1 --off  --output DP1 --off  --output VGA1 --auto"
		//balias display-internal-only        "xrandr --output LVDS1 --auto --output HDMI1 --off  --output DP1 --off  --output VGA1 --off"
		//balias display-clone-hdmi           "xrandr --output LVDS1 --auto --output HDMI1 --auto --output DP1 --off  --output VGA1 --off"
		//balias display-clone-dp             "xrandr --output LVDS1 --auto --output HDMI1 --off  --output DP1 --auto --output VGA1 --off"
		//balias display-clone-vga            "xrandr --output LVDS1 --auto --output HDMI1 --off --output DP1 --off  --output VGA1 --auto"
		//
		//{% elif ansible_hostname == "M16" %}
		//
		//balias display-external-hdmi-only   "xrandr --output eDP1 --off  --output HDMI1 --auto --output DP1 --off"
		//balias display-external-dp-only     "xrandr --output eDP1 --off  --output HDMI1 --off  --output DP1 --auto"
		//balias display-internal-only        "xrandr --output eDP1 --auto --output HDMI1 --off  --output DP1 --off"
		//balias display-clone-hdmi           "xrandr --output eDP1 --auto --output HDMI1 --auto --output DP1 --off"
		//balias display-clone-dp             "xrandr --output eDP1 --auto --output HDMI1 --off  --output DP1 --auto"

		orderedFlagValues := []*pflag.Flag{}

		cmd.Flags().Visit(func(flag *pflag.Flag) {
			orderedFlagValues = append(orderedFlagValues, flag)
		})

		displayNames := []string{}
		for _, displayInfo := range displays {
			displayNames = append(displayNames, displayInfo.Name)
		}

		displayConfigs := createDisplayConfigurationsFromArgs(displayNames, modes, positions, rate, primary, off, orderedFlagValues)
		return util.SetDisplayConfigs(displayConfigs)
	},
}

func createDisplayConfigurationsFromArgs(
	displayNameArgs []string,
	modeArgs []string,
	positionArgs []string,
	rateArgs []int,
	primaryArgs []bool,
	offArgs []bool,
	args []*pflag.Flag,
) []util.DisplayConfig {
	displayConfigs := make([]util.DisplayConfig, 0)

	for i, displayName := range displayNameArgs {
		displayConfig := util.NewDisplayConfig(displayName)
		if i < len(modeArgs) {
			displayConfig.Mode = modeArgs[i]
		}
		if i < len(positionArgs) {
			displayConfig.Position = positionArgs[i]
		}
		if i < len(rateArgs) {
			displayConfig.Rate = rateArgs[i]
		}
		if i < len(primaryArgs) {
			displayConfig.Primary = primaryArgs[i]
		}
		if i < len(offArgs) {
			displayConfig.Off = offArgs[i]
		}
		displayConfigs = append(displayConfigs, displayConfig)
	}

	return displayConfigs
}

func init() {
	Command.AddCommand(displayConfigCmd)

	displayConfigCmd.Flags().StringSliceVarP(&displays, "display", "d", []string{}, "The display to configure")
	displayConfigCmd.MarkFlagRequired("display")

	displayConfigCmd.Flags().StringSliceVarP(&modes, "mode", "m", []string{}, "The display mode to set")
	displayConfigCmd.Flags().StringSliceVarP(&positions, "position", "p", []string{}, "The display position to set")
	displayConfigCmd.Flags().IntSliceVarP(&rate, "rate", "r", []int{}, "The display refresh rate to set")
	displayConfigCmd.Flags().BoolSliceVarP(&primary, "primary", "P", []bool{}, "Set the display as primary")
	displayConfigCmd.Flags().BoolSliceVarP(&off, "off", "O", []bool{}, "Turn the display off")
	displayConfigCmd.Flags().BoolSliceVarP(&auto, "auto", "A", []bool{}, "Set the display mode automatically")
}
