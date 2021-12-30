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
	"github.com/markusressel/system-control/internal"
	"github.com/spf13/cobra"
	"log"
	"strings"
	"time"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Reboot the system gracefully",
	Long:  `Reboots the system gracefully by first closing all currently open windows.`,
	Run: func(cmd *cobra.Command, args []string) {
		openWindows := internal.FindOpenWindows()

		for _, element := range openWindows {
			windowId := strings.Split(element, " ")[0]
			_, err := internal.ExecCommand("wmctrl", "-i", "-c", windowId)
			if err != nil {
				log.Fatal(err)
			}
		}

		// wait for all windows to disappear
		for {
			openWindows = internal.FindOpenWindows()
			if len(openWindows) <= 0 {
				break
			} else {
				time.Sleep(time.Second)
			}
		}

		_, err := internal.ExecCommand("reboot")
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(restartCmd)
}
