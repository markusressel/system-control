package redshift

import (
	"context"
	"fmt"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
)

const (
	redshiftLockName              = "cmd-display-redshift"
	defaultBrightnessPollInterval = 5 * time.Second
)

func init() {
	daemon.Register(daemon.Op{
		Name:     "display.redshift.brightness",
		Requires: []string{"DISPLAY"},
		// querying needs no lock, setting is serialized (and queue-dropped) like the steps
		LockFor: func(a daemon.Args) string {
			if a.Flag("brightness") == "-1" {
				return ""
			}
			return redshiftLockName
		},
		Run: runBrightness,
		// setting a value can't be watched
		Watch: func(a daemon.Args) bool { return a.Flag("brightness") == "-1" },
		// Changes made through this tool refresh the watchers right away, polling only
		// catches the rest.
		Poll: func() time.Duration {
			if configured := configuration.Get().Daemon.PollIntervals["display"]; configured > 0 {
				return configured
			}
			return defaultBrightnessPollInterval
		},
		Invalidates: []string{"display."},
	})

	daemon.Register(daemon.Op{
		Name:        "display.redshift.brightness.inc",
		Lock:        redshiftLockName,
		Requires:    []string{"DISPLAY"},
		Run:         stepBrightness(1),
		Invalidates: []string{"display."},
	})

	daemon.Register(daemon.Op{
		Name:        "display.redshift.brightness.dec",
		Lock:        redshiftLockName,
		Requires:    []string{"DISPLAY"},
		Run:         stepBrightness(-1),
		Invalidates: []string{"display."},
	})
}

func runBrightness(ctx context.Context, a daemon.Args) ([]string, error) {
	value, err := a.Float("brightness")
	if err != nil {
		return nil, err
	}

	if value == -1 {
		line, err := queryBrightness(a.Flag("display"), a.Flag("format"))
		return []string{line}, err
	}
	return setBrightness(a.Flag("display"), value, a.Flag("format"))
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

// setBrightness is run with the redshift lock held, see the op registration.
func setBrightness(displayName string, value float64, format string) ([]string, error) {
	displays, err := parseDisplayParam(displayName)
	if err != nil {
		return nil, err
	}

	var lastSet float64
	for _, d := range displays {
		newBrightness := clampBrightnessToConfig(value)
		if newBrightness != getLastSetBrightness(d) {
			if err := ApplyRedshift(d, -1, newBrightness, -1); err != nil {
				return nil, err
			}
		}
		lastSet = getLastSetBrightness(d)
	}
	return []string{fmt.Sprintf(format, lastSet)}, nil
}

// stepBrightness returns the op implementation to change the brightness by --step.
// The op holds the redshift lock.
func stepBrightness(direction float64) func(ctx context.Context, a daemon.Args) ([]string, error) {
	return func(ctx context.Context, a daemon.Args) ([]string, error) {
		step, err := a.Float("step")
		if err != nil {
			return nil, err
		}
		if err := validateStep(step); err != nil {
			return nil, err
		}

		displays, err := parseDisplayParam(a.Flag("display"))
		if err != nil {
			return nil, err
		}

		var lines []string
		for _, d := range displays {
			newBrightness := clampBrightnessToConfig(getLastSetBrightness(d) + direction*step)
			if err := ApplyRedshift(d, -1, newBrightness, -1); err != nil {
				return nil, err
			}
			lines = append(lines, fmt.Sprintf("%.2f", newBrightness))
		}
		return lines, nil
	}
}

// notifyDaemon tells a running daemon that persisted display state changed.
func notifyDaemon() {
	daemon.Notify("display.")
}
