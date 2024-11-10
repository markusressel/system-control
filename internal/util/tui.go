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
package util

import (
	"fmt"
	"github.com/elliotchance/orderedmap/v2"
	"math"
	"os"
	"regexp"
	"text/tabwriter"
)

const (
	DisplayBacklightPath = "/sys/class/backlight"
	LedsPath             = "/sys/class/leds"
	MaxBrightness        = "max_brightness"
	Brightness           = "brightness"
)

func FindOpenWindows() ([]string, error) {
	result, err := ExecCommand("wmctrl", "-l")
	if err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	r := regexp.MustCompile("(0x[0-9a-f]+) +(\\d+) +(" + hostname + "|N/A) +(.*)")
	matches := r.FindAllString(result, -1)
	return matches, nil
}

func RoundToTwoDecimals(number float64) float64 {
	return math.Round(number*100) / 100
}

func PrintFormattedTable(title string, properties map[string]string) {
	orderedProperties := orderedmap.NewOrderedMap[string, string]()
	for key, value := range properties {
		orderedProperties.Set(key, value)
	}
	PrintFormattedTableOrdered(title, orderedProperties)
}

func PrintFormattedTableOrdered(title string, properties *orderedmap.OrderedMap[string, string]) {
	fmt.Println(title)
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	for key, value := range properties.Iterator() {
		_, _ = fmt.Fprintf(w, "  %s:\t%s\t\n", key, value)
	}
	_ = w.Flush()
}
