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
package video

import (
	"github.com/markusressel/system-control/internal"
	"github.com/spf13/cobra"
)

var unloadDriverCmd = &cobra.Command{
	Use:   "unload",
	Short: "Unload the video driver, preventing access to any video device",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		internal.ExecCommand("lsmod", "uvcvideo")
		internal.ExecCommand("rmmod", "-f", "uvcvideo")
	},
}

func init() {
	Command.AddCommand(unloadDriverCmd)
}
