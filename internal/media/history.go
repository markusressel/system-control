package media

import (
	"strings"
	"sync"
	"time"
)

// history remembers when each player was last seen playing, so that play resumes the player
// that was playing last instead of every paused one. It is fed by Follow, i.e. only in the
// daemon, and lost when the daemon restarts.
var history = struct {
	sync.Mutex
	status      map[string]string
	lastPlaying map[string]time.Time
}{status: map[string]string{}, lastPlaying: map[string]time.Time{}}

// now is a variable for tests.
var now = time.Now

// recordStatus notes the status of a player reported by Follow. A player counts as playing
// until it reports that it stopped playing.
func recordStatus(player string, status string) {
	history.Lock()
	defer history.Unlock()
	if status == "Playing" || history.status[player] == "Playing" {
		history.lastPlaying[player] = now()
	}
	history.status[player] = status
}

// resetHistory forgets all players. For tests.
func resetHistory() {
	history.Lock()
	defer history.Unlock()
	clear(history.status)
	clear(history.lastPlaying)
}

// LastPlayed returns the one of the given players that was playing most recently, or "" if
// none of them was seen playing.
func LastPlayed(players []string) string {
	history.Lock()
	defer history.Unlock()
	best := ""
	var bestAt time.Time
	for _, player := range players {
		if at, ok := history.lastPlaying[player]; ok && at.After(bestAt) {
			best, bestAt = player, at
		}
	}
	return best
}

// parseFollowLine splits a line of the follow command ("<player>\t<status>"). An empty line
// (no players left) has no player.
func parseFollowLine(line string) (player string, status string, ok bool) {
	player, status, ok = strings.Cut(strings.TrimSpace(line), "\t")
	if !ok || player == "" {
		return "", "", false
	}
	return player, strings.TrimSpace(status), true
}
