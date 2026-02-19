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
package redshift

import (
	"github.com/spf13/cobra"
)

var redshiftResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the currently applied redshift",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		redshiftLock = lockRedshift()
		defer redshiftLock.Unlock()

		displays, err := parseDisplayParam(display)
		if err != nil {
			return err
		}

		for _, display := range displays {
			err := ResetRedshift(display)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	Command.AddCommand(redshiftResetCmd)
}
