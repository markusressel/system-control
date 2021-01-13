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
package cmd

import (
	"github.com/spf13/cobra"
)

// switchCmd represents the switch command
var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch the default sink",
	Long: `Switches the default audio sink and moves all existing audio streams to the given one.
You can specify the audio sink using its index, but also using other strings that occur in its description:

> system-control audio sink switch "headphone"

> system-control audio sink switch "NVIDIA"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		searchString := args[0]
		//sinkIdx := findSinkPulse(searchString)
		sinkIdx := findSinkPipewire(searchString)
		switchSinkPulse(sinkIdx)

		//list-objects        	List objects or current remote. [<interface>]
		//info                	Get info about an object. <object-id>|all
		//create-link         	Create a link between nodes. <node-id> <port-id> <node-id> <port-id> [<properties>]

		//enum-params         	Enumerate params of an object <object-id> <param-id>
		//set-param           	Set param of an object <object-id> <param-id> <param-json>
	},
}

func init() {
	sinkCmd.AddCommand(switchCmd)

	// Here you will define your flags and configuration settings.

}
