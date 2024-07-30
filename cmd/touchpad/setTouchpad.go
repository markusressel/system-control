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
package touchpad

import (
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
	"strconv"
)

var setTouchpadCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the emabled state of the Touchpad",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		enabled, err := strconv.ParseBool(args[0])
		if err != nil {
			return err
		}

		return util.SetTouchpadEnabled(enabled)
	},
}

func init() {
	Command.AddCommand(setTouchpadCmd)
}
