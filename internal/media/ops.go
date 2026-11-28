package media

import (
	"context"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
)

const (
	// PlayerFlag is the name of the flag of the `media` commands that selects the player by name or regex.
	PlayerFlag = "player"

	lockName = "media"
)

// followedPollInterval is the safety net while player changes are followed, in case playerctl
// misses a change.
const followedPollInterval = time.Minute

// pollInterval is how often watched media state is re-queried. A status that is followed is
// pushed, so only positions need to be polled.
func pollInterval(def time.Duration, followed bool) func() time.Duration {
	return func() time.Duration {
		if configured := configuration.Get().Daemon.PollIntervals["media"]; configured > 0 {
			return configured
		}
		if followed && Following() {
			return followedPollInterval
		}
		return def
	}
}

// eventSource refreshes watched media state whenever a player changes its status.
func eventSource(ctx context.Context, refresh func(prefixes ...string)) {
	go Follow(ctx, func() { refresh("media") })
}

// playerctl returns an op implementation printing the output of a playerctl command.
func playerctl(command string) func(context.Context, daemon.Args) ([]string, error) {
	return func(ctx context.Context, a daemon.Args) ([]string, error) {
		output, err := Query(command, a.Flag(PlayerFlag))
		if err != nil || output == "" {
			return nil, err
		}
		return []string{output}, nil
	}
}

func init() {
	daemon.RegisterSource(eventSource, FollowStatus)

	// `media` without a sub command shows the status
	for _, name := range []string{"media", "media.status"} {
		daemon.Register(daemon.Op{
			Name:  name,
			Run:   playerctl("status"),
			Watch: daemon.Always,
			Poll:  pollInterval(2*time.Second, true),
		})
	}
	daemon.Register(daemon.Op{
		Name:  "media.position",
		Run:   playerctl("position"),
		Watch: daemon.Always,
		Poll:  pollInterval(time.Second, false),
	})

	// playing something changes the status and position right away
	for name, command := range map[string]string{
		"media.pause":    "pause",
		"media.stop":     "stop",
		"media.next":     "next",
		"media.previous": "previous",
	} {
		daemon.Register(daemon.Op{
			Name:        name,
			Lock:        lockName,
			Run:         playerctl(command),
			Invalidates: []string{"media"},
		})
	}

	// without --player both resume a single player, see Play
	for name, run := range map[string]func(string) error{
		"media.play":            Play,
		"media.togglePlayPause": TogglePlayPause,
	} {
		daemon.Register(daemon.Op{
			Name: name,
			Lock: lockName,
			Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
				return nil, run(a.Flag(PlayerFlag))
			},
			Invalidates: []string{"media"},
		})
	}

	daemon.Register(daemon.Op{
		Name: "media.players",
		Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
			players, err := ListPlayers()
			if err != nil {
				return nil, err
			}
			return MatchPlayers(players, a.Flag(PlayerFlag))
		},
	})
}
