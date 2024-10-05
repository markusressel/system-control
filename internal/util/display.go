package util

import (
	"strings"
)

func GetDisplays() (displays []string, err error) {
	result, err := ExecCommand(
		"xrandr",
		"--listmonitors",
	)
	if err != nil {
		return nil, err
	}

	resultLines := strings.Split(result, "\n")

	for i, line := range resultLines {
		if i == 0 {
			continue
		} else {
			segments := strings.Split(line, " ")
			displayName := segments[len(segments)-1]
			displays = append(displays, displayName)
		}
	}
	return displays, nil
}
