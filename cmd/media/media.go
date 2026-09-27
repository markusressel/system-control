package media

import (
	"fmt"
	"strings"

	internalmedia "github.com/markusressel/system-control/internal/media"
	"github.com/spf13/cobra"
)

var player string

var Command = &cobra.Command{
	Use:              "media",
	Short:            "Control media players via playerctl",
	TraverseChildren: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printPlayerCtlOutput("status", true)
	},
}

// statusPriority ranks MPRIS states: Playing > Paused > Stopped
var statusPriority = map[string]int{
	"Playing": 3,
	"Paused":  2,
	"Stopped": 1,
}

func printPlayerCtlOutput(command string, targetAllWhenUnspecified bool) error {
	output, err := internalmedia.RunPlayerCtl(command, player, targetAllWhenUnspecified)
	if err != nil {
		return err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}

	// If evaluating status across all players, pick the highest-priority state
	if command == "status" && targetAllWhenUnspecified && player == "" {
		output = getHighestStatus(output)
	}

	if output != "" {
		fmt.Println(output)
	}
	return nil
}

func getHighestStatus(rawOutput string) string {
	lines := strings.Split(rawOutput, "\n")
	bestStatus := ""
	highestScore := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if score, exists := statusPriority[trimmed]; exists {
			if score > highestScore {
				highestScore = score
				bestStatus = trimmed
			}
		}
	}

	// Fallback to the first non-empty line if custom/unrecognized statuses appear
	if bestStatus == "" && len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}

	return bestStatus
}

func init() {
	Command.PersistentFlags().StringVarP(
		&player,
		"player", "p",
		"",
		"Player name or regex",
	)
}
