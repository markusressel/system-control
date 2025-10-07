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
package wifi

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/markusressel/system-control/internal/util"
	"github.com/markusressel/system-control/internal/wifi"
	"github.com/spf13/cobra"
)

var (
	flagFilterConnected bool
	filterBSSID         string
	filterSSID          string
	filterMode          string
	filterChannel       int
	filterBandwidth     string
	filterFrequency     string
	filterRate          string
	filterSignal        int
	filterSecurity      string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all known WiFi networks",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		networks, err := wifi.GetNetworks()
		if err != nil {
			return err
		}

		// filter entries
		networks = util.FilterFunc(networks, func(network wifi.WiFiNetwork) bool {
			if flagFilterConnected && !network.Connected {
				return false
			}
			if filterBSSID != "" && !util.ContainsIgnoreCase(network.BSSID, filterBSSID) {
				return false
			}
			if filterSSID != "" && !util.ContainsIgnoreCase(network.SSID, filterSSID) {
				return false
			}
			if filterMode != "" && !util.ContainsIgnoreCase(network.Mode, filterMode) {
				return false
			}
			if filterChannel != 0 && network.Channel != filterChannel {
				return false
			}
			if filterBandwidth != "" && !util.ContainsIgnoreCase(network.Bandwidth, filterBandwidth) {
				return false
			}
			if filterFrequency != "" && !util.ContainsIgnoreCase(network.Frequency, filterFrequency) {
				return false
			}
			if filterRate != "" && !util.ContainsIgnoreCase(network.Rate, filterRate) {
				return false
			}
			if filterSignal != 0 && network.Signal != filterSignal {
				return false
			}
			if filterSecurity != "" && !util.ContainsIgnoreCase(network.Security, filterSecurity) {
				return false
			}
			return true
		})

		// sort entries
		slices.SortFunc(networks, func(a, b wifi.WiFiNetwork) int {
			return cmp.Or(
				// connected networks first
				-1*cmp.Compare(strconv.FormatBool(a.Connected), strconv.FormatBool(b.Connected)),
				// then sort by signal strength
				-1*cmp.Compare(a.Signal, b.Signal),
				// then sort by SSID
				util.CompareIgnoreCase(a.SSID, b.SSID),
			)
		})

		for i, network := range networks {
			properties := orderedmap.NewOrderedMap[string, string]()
			properties.Set("Connected", strconv.FormatBool(network.Connected))
			properties.Set("SSID", network.SSID)
			properties.Set("BSSID", network.BSSID)
			properties.Set("Mode", network.Mode)
			properties.Set("Channel", fmt.Sprintf("%v", network.Channel))
			properties.Set("Bandwidth", fmt.Sprintf("%v", network.Bandwidth))
			properties.Set("Frequency", fmt.Sprintf("%v", network.Frequency))
			properties.Set("Rate", fmt.Sprintf("%v", network.Rate))
			properties.Set("Signal", fmt.Sprintf("%v", network.Signal))
			properties.Set("Bars", fmt.Sprintf("%v", network.Bars))
			properties.Set("Security", network.Security)

			util.PrintFormattedTableOrdered(network.SSID, properties)

			if i < len(networks)-1 {
				fmt.Println()
			}
		}

		return nil
	},
}

func init() {
	Command.AddCommand(listCmd)

	listCmd.Flags().BoolVarP(&flagFilterConnected, "connected", "c", false, "Filter by connected state")
	listCmd.Flags().StringVarP(&filterBSSID, "bssid", "b", "", "Filter by BSSID")
	listCmd.Flags().StringVarP(&filterSSID, "ssid", "s", "", "Filter by SSID")
	listCmd.Flags().StringVarP(&filterMode, "mode", "m", "", "Filter by mode")
	listCmd.Flags().IntVarP(&filterChannel, "channel", "C", 0, "Filter by channel")
	listCmd.Flags().StringVarP(&filterBandwidth, "bandwidth", "B", "", "Filter by bandwidth")
	listCmd.Flags().StringVarP(&filterFrequency, "frequency", "F", "", "Filter by frequency")
	listCmd.Flags().StringVarP(&filterRate, "rate", "r", "", "Filter by rate")
	listCmd.Flags().IntVarP(&filterSignal, "signal", "S", 0, "Filter by signal")
	listCmd.Flags().StringVarP(&filterSecurity, "security", "q", "", "Filter by security")
}
