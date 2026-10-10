package media

import (
	"context"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
	internalmedia "github.com/markusressel/system-control/internal/media"
)

const lockName = "media"

// MPRIS players are not observed, so watched state is polled.
func pollInterval(def time.Duration) func() time.Duration {
	return func() time.Duration {
		if configured := configuration.Get().Daemon.PollIntervals["media"]; configured > 0 {
			return configured
		}
		return def
	}
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
			Poll:  pollInterval(2 * time.Second),
		})
	}
	daemon.Register(daemon.Op{
		Name:  "media.position",
		Run:   playerctl("position"),
		Watch: always,
		Poll:  pollInterval(time.Second),
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
