package media

import (
	"context"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
	internalmedia "github.com/markusressel/system-control/internal/media"
)

const lockName = "media"

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
		if followed && internalmedia.Following() {
			return followedPollInterval
		}
		return def
	}
}

// eventSource refreshes watched media state whenever a player changes its status.
func eventSource(ctx context.Context, refresh func(prefixes ...string)) {
	go internalmedia.Follow(ctx, func() { refresh("media") })
}

// playerctl returns an op implementation printing the output of a playerctl command.
func playerctl(command string) func(context.Context, daemon.Args) ([]string, error) {
	return func(ctx context.Context, a daemon.Args) ([]string, error) {
		output, err := internalmedia.Query(command, a.Flag("player"))
		if err != nil || output == "" {
			return nil, err
		}
		return []string{output}, nil
	}
}

func init() {
	daemon.RegisterSource(eventSource, internalmedia.FollowStatus)

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

	// without --player both resume a single player, see internalmedia.Play
	for name, run := range map[string]func(string) error{
		"media.play":            internalmedia.Play,
		"media.togglePlayPause": internalmedia.TogglePlayPause,
	} {
		daemon.Register(daemon.Op{
			Name: name,
			Lock: lockName,
			Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
				return nil, run(a.Flag("player"))
			},
			Invalidates: []string{"media"},
		})
	}

	daemon.Register(daemon.Op{
		Name: "media.players",
		Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
			players, err := internalmedia.ListPlayers()
			if err != nil {
				return nil, err
			}
			return internalmedia.MatchPlayers(players, a.Flag("player"))
		},
	})
}
