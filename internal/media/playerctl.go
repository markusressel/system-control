package media

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/markusressel/system-control/internal/util"
)

func ListPlayers() ([]string, error) {
	output, err := util.ExecCommand("playerctl", "-l")
	if err != nil {
		return nil, err
	}

	if output == "" {
		return []string{}, nil
	}

	lines := strings.Split(output, "\n")
	players := make([]string, 0, len(lines))
	for _, line := range lines {
		player := strings.TrimSpace(line)
		if player == "" {
			continue
		}
		players = append(players, player)
	}

	slices.Sort(players)
	players = slices.Compact(players)

	return players, nil
}

func MatchPlayers(players []string, pattern string) ([]string, error) {
	if pattern == "" {
		return players, nil
	}

	for _, player := range players {
		if strings.EqualFold(player, pattern) {
			return []string{player}, nil
		}
	}

	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid player regex %q: %w", pattern, err)
	}

	matches := make([]string, 0)
	for _, player := range players {
		if re.MatchString(player) {
			matches = append(matches, player)
		}
	}

	return matches, nil
}

func ResolvePlayer(pattern string) (string, error) {
	players, err := ListPlayers()
	if err != nil {
		return "", err
	}

	matches, err := MatchPlayers(players, pattern)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no player matches pattern %q", pattern)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("multiple players match pattern %q: %s", pattern, strings.Join(matches, ", "))
	}

	return matches[0], nil
}

func RunPlayerCtl(command string, playerPattern string, targetAllWhenUnspecified bool) (string, error) {
	args := []string{}

	if playerPattern != "" {
		player, err := ResolvePlayer(playerPattern)
		if err != nil {
			return "", err
		}
		args = append(args, "-p", player)
	} else if targetAllWhenUnspecified {
		args = append(args, "-a")
	}

	args = append(args, command)

	return util.ExecCommand("playerctl", args...)
}
