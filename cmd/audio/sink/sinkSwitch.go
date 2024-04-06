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
package sink

import (
	"errors"
	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch the default sink",
	Long: `Switches the default audio sink and moves all existing audio streams to the given one.
You can specify the audio sink using its index, but also using other strings that occur in its description:

> system-control audio sink switch "headphone"

> system-control audio sink switch "NVIDIA"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		searchString := args[0]
		//sinkIdx := findSinkPulse(searchString)
		//switchSinkPulse(sinkIdx)

		state := pipewire.PwDump()

		nodes := state.FindNodesByName(searchString)
		if len(nodes) <= 0 {
			return errors.New("no sink found")
		}
		if len(nodes) > 1 {
			return errors.New("ambiguous sink name")
		}
		node := nodes[0]
		return state.SwitchSinkTo(node)
	},
}

func init() {
	SinkCmd.AddCommand(switchCmd)
}
