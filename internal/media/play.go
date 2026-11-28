package media

import (
	"strings"

	"github.com/markusressel/system-control/internal/util"
)

// Statuses returns the status of every player, by player instance.
func Statuses() (map[string]string, error) {
	output, err := util.ExecCommand("playerctl", "--all-players", "--format", followFormat, "status")
	if err != nil {
		return nil, err
	}
	statuses := map[string]string{}
	for _, line := range strings.Split(output, "\n") {
		if player, status, ok := parseFollowLine(line); ok {
			statuses[player] = status
		}
	}
	return statuses, nil
}

// Play starts the players matching the pattern. Without a pattern it resumes a single player,
// the one that was playing last (see LastPlayed), not every paused one: an old paused voice
// message must not start along with the video. Nothing happens while a player is playing.
// Without a known last player (no daemon, or no player played since it started), playerctl
// picks one.
func Play(playerPattern string) error {
	if playerPattern != "" {
		_, err := RunPlayerCtl("play", playerPattern, false)
		return err
	}
	statuses, err := Statuses()
	if err != nil {
		return err
	}
	if anyPlaying(statuses) {
		return nil
	}
	return playLastPlayed(statuses)
}

// TogglePlayPause toggles the players matching the pattern. Without a pattern it pauses every
// playing player, or, if none plays, resumes a single one like Play.
func TogglePlayPause(playerPattern string) error {
	if playerPattern != "" {
		_, err := RunPlayerCtl("play-pause", playerPattern, false)
		return err
	}
	statuses, err := Statuses()
	if err != nil {
		return err
	}
	if anyPlaying(statuses) {
		_, err := util.ExecCommand("playerctl", "--all-players", "pause")
		return err
	}
	return playLastPlayed(statuses)
}

func anyPlaying(statuses map[string]string) bool {
	for _, status := range statuses {
		if status == "Playing" {
			return true
		}
	}
	return false
}

func playLastPlayed(statuses map[string]string) error {
	players := make([]string, 0, len(statuses))
	for player := range statuses {
		players = append(players, player)
	}
	if player := LastPlayed(players); player != "" {
		_, err := util.ExecCommand("playerctl", "--player", player, "play")
		return err
	}
	_, err := util.ExecCommand("playerctl", "play")
	return err
}
