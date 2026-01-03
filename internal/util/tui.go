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
	"math"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/elliotchance/orderedmap/v2"
)

const (
	DisplayBacklightPath = "/sys/class/backlight"
	LedsPath             = "/sys/class/leds"
	MaxBrightness        = "max_brightness"
	Brightness           = "brightness"
)

// FindOpenWindows returns a list of currently open windows
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

// RoundToTwoDecimals rounds a float to (at most) two decimal places
func RoundToTwoDecimals(number float64) float64 {
	return math.Round(number*100) / 100
}

// PrintFormattedTableOrdered prints a formatted table to the console
func PrintFormattedTableOrdered(title string, properties *orderedmap.OrderedMap[string, string]) {
	if len(title) > 0 {
		title = fmt.Sprintf("%s", title)
		fmt.Println(title)
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	for key, value := range properties.Iterator() {
		_, _ = fmt.Fprintf(w, "  %s:\t%s\t\n", key, value)
	}
	_ = w.Flush()
}

func ParseDelimitedTable[T any](
	input string,
	delimiter string,
	producer func(row []string) T,
) ([]T, error) {
	result := make([]T, 0)

	lines := strings.Split(input, "\n")
	lines = FilterFunc(lines, func(e string) bool {
		return len(e) > 0
	})
	for i := 0; i < len(lines); i++ {
		row := strings.Split(lines[i], delimiter)
		result = append(result, producer(row))
	}

	return result, nil
}

const DefaultColumnHeaderRegexPattern = "\\S+\\s*"

// ParseTable attempts to parse the given input string as a table, converting each row into
// a slice of structs using the provided producer function.
//
// The first line is expected to be the header line, all following lines are expected to be data lines.
// The header line is used to determine the number of columns, their with and order.
// The cellSeparator is a regex that is used to separate columns within the header lines. The regex match
// must include any whitespace that follows the column title.
//
// The producer function is expected to map the values of a row into a struct of type T.
func ParseTable[T any](
	input string,
	cellSeparator string,
	producer func(row []string) T,
) ([]T, error) {
	result := make([]T, 0)

	lines := strings.Split(input, "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid table format")
	}

	headerCellRegex := regexp.MustCompile(cellSeparator)
	header := headerCellRegex.FindAllString(lines[0], -1)
	if len(header) < 2 {
		return nil, fmt.Errorf("invalid table format")
	}

	for i := 1; i < len(lines); i++ {
		row := make([]string, 0)
		currentLine := lines[i]
		startIdx := 0
		for i := 0; i < len(header); i++ {
			endIdx := startIdx + len(header[i])
			columnValue := SubstringRunes(currentLine, startIdx, endIdx)
			row = append(row, columnValue)
			startIdx = endIdx
		}
		if len(row) < 2 {
			return nil, fmt.Errorf("invalid table format")
		}
		result = append(result, producer(row))
	}

	return result, nil
}
