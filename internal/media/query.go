package media

import "strings"

// statusPriority ranks MPRIS states: Playing > Paused > Stopped
var statusPriority = map[string]int{
	"Playing": 3,
	"Paused":  2,
	"Stopped": 1,
}

// Query runs a playerctl command against the players matching the pattern (all players if
// empty) and returns its trimmed output. The status of all players is reduced to the most
// active one.
func Query(command string, playerPattern string) (string, error) {
	output, err := RunPlayerCtl(command, playerPattern, true)
	if err != nil {
		return "", err
	}

	output = strings.TrimSpace(output)
	if command == "status" && playerPattern == "" && output != "" {
		output = HighestStatus(output)
	}
	return output, nil
}

// HighestStatus returns the most active state of several lines of player status.
func HighestStatus(rawOutput string) string {
	lines := strings.Split(rawOutput, "\n")
	bestStatus := ""
	highestScore := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if score, exists := statusPriority[trimmed]; exists && score > highestScore {
			highestScore = score
			bestStatus = trimmed
		}
	}

	// Fallback to the first line if custom/unrecognized statuses appear
	if bestStatus == "" && len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}

	return bestStatus
}
