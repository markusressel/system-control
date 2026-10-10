package media

import (
	"context"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
	internalmedia "github.com/markusressel/system-control/internal/media"
)

const lockName = "media"

// followedPollInterval is the safety net while player changes are followed: playerctl reports
// the most recently updated player only, so a change of another one is caught by polling.
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

// EventSource refreshes watched media state whenever a player changes its status.
// It is a daemon.Source.
func EventSource(ctx context.Context, refresh func(prefixes ...string)) {
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
	always := func(daemon.Args) bool { return true }

	// `media` without a sub command shows the status
	for _, name := range []string{"media", "media.status"} {
		daemon.Register(daemon.Op{
			Name:  name,
			Run:   playerctl("status"),
			Watch: always,
			Poll:  pollInterval(2*time.Second, true),
		})
	}
	daemon.Register(daemon.Op{
		Name:  "media.position",
		Run:   playerctl("position"),
		Watch: always,
		Poll:  pollInterval(time.Second, false),
	})

	// playing something changes the status and position right away
	for name, command := range map[string]string{
		"media.play":            "play",
		"media.pause":           "pause",
		"media.stop":            "stop",
		"media.next":            "next",
		"media.previous":        "previous",
		"media.togglePlayPause": "play-pause",
	} {
		daemon.Register(daemon.Op{
			Name:        name,
			Lock:        lockName,
			Run:         playerctl(command),
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
