package redshift

import (
	"context"
	"fmt"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
)

const defaultBrightnessPollInterval = 5 * time.Second

func init() {
	// Only the read-only form of `display redshift brightness` is an op. Setting a value
	// (-b) is handled by the command itself, which notifies the daemon through ApplyRedshift.
	daemon.Register(daemon.Op{
		Name: "display.redshift.brightness",
		Flags: []daemon.FlagDef{
			{Name: "display", Short: "d"},
			{Name: "format", Short: "f", Default: "%.2f"},
			{Name: "brightness", Short: "b", Default: "-1", Mutating: true},
		},
		Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
			value, err := queryBrightness(a.Flag("display"), a.Flag("format"))
			if err != nil {
				return nil, err
			}
			return []string{value}, nil
		},
		Watch: func(daemon.Args) bool { return true },
		// Changes made through this tool notify the daemon, polling only catches the rest.
		Poll: func() time.Duration {
			if configured := configuration.CurrentConfig.Daemon.PollIntervals["display"]; configured > 0 {
				return configured
			}
			return defaultBrightnessPollInterval
		},
	})
}

// queryBrightness returns the formatted last set brightness of the given display
// (of the last display, if none is given).
func queryBrightness(displayName string, format string) (string, error) {
	displays, err := parseDisplayParam(displayName)
	if err != nil {
		return "", err
	}
	if len(displays) == 0 {
		return "", fmt.Errorf("no displays found")
	}

	var lastSet float64
	for _, d := range displays {
		lastSet = getLastSetBrightness(d)
	}
	return fmt.Sprintf(format, lastSet), nil
}

// notifyDaemon tells a running daemon that persisted display state changed.
func notifyDaemon() {
	daemon.Notify("display.")
}
