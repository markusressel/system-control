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
package audio

import (
	"github.com/markusressel/system-control/internal"
	"log"

	"github.com/spf13/cobra"
)

// sinkCmd represents the sink command
var sinkCmd = &cobra.Command{
	Use:   "sink",
	Short: "Show a list of all available sinks",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		result, err := internal.ExecCommand("pactl", "list", "sinks")
		if err != nil {
			log.Fatal(err)
		}
		print(result)
	},
}

func init() {
	Command.AddCommand(sinkCmd)

	// Here you will define your flags and configuration settings.

}