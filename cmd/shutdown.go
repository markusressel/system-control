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
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

var shutdownCmd = &cobra.Command{
	Use:   "shutdown",
	Short: "Shutdown the system gracefully",
	Long:  `Shuts down the system in a graceful way, first closing all opened applications.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		openWindows, err := util.FindOpenWindows()
		if err != nil {
			return err
		}

		for _, element := range openWindows {
			windowId := strings.Split(element, " ")[0]
			_, err := util.ExecCommand("wmctrl", "-i", "-c", windowId)
			if err != nil {
				return err
			}
		}

		// wait for all windows to disappear
		for {
			openWindows, err = util.FindOpenWindows()
			if err != nil {
				return err
			}
			if len(openWindows) <= 0 {
				break
			} else {
				time.Sleep(time.Second)
			}
		}

		_, err = util.ExecCommand("poweroff")
		return err
	},
}

func init() {
	RootCmd.AddCommand(shutdownCmd)
}
