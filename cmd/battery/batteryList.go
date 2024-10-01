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
package battery

import (
	"fmt"
	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

var batteryListCmd = &cobra.Command{
	Use:   "list",
	Short: "Get a list of all known batteries",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		batteries, err := util.GetBatteryList()
		if err != nil {
			return err
		}
		for i, battery := range batteries {
			fmt.Println(battery.Name)

			w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
			fmt.Fprintf(w, "  Path:\t%s\t\n", battery.Path)
			fmt.Fprintf(w, "  Type:\t%s\t\n", battery.Type)
			fmt.Fprintf(w, "  Manufacturer:\t%s\t\n", battery.Manufacturer)
			fmt.Fprintf(w, "  Model:\t%s\t\n", battery.Model)
			fmt.Fprintf(w, "  Serial:\t%s\t\n", battery.SerialNumber)
			fmt.Fprintf(w, "  Capacity:\t%d\t\n", battery.Capacity)
			fmt.Fprintf(w, "  Capacity Level:\t%s\t\n", battery.CapacityLevel)
			fmt.Fprintf(w, "  Online:\t%v\t\n", battery.Online)
			fmt.Fprintf(w, "  Status:\t%s\t\n", battery.Status)
			fmt.Fprintf(w, "  Scope:\t%s\t\n", battery.Scope)
			w.Flush()

			if i < len(batteries)-1 {
				fmt.Println()
			}

		}
		return nil
	},
}

func init() {
	Command.AddCommand(batteryListCmd)
}
