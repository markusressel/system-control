package redshift

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
)

const (
	redshiftLockName    = "cmd-display-redshift"
	defaultPollInterval = 5 * time.Second
)

// displayPollInterval is how often watched display state is re-queried. Changes made through
// this tool refresh the watchers right away, polling only catches the rest.
func displayPollInterval() time.Duration {
	if configured := configuration.Get().Daemon.PollIntervals["display"]; configured > 0 {
		return configured
	}
	return defaultPollInterval
}

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
		// the format is applied per watcher
		OutputFlags: []string{"format"},
		Render:      renderFormat,
		// setting a value can't be watched
		Watch:       func(a daemon.Args) bool { return a.Flag("brightness") == "-1" },
		Poll:        displayPollInterval,
		Invalidates: []string{"display."},
	})

	daemon.Register(daemon.Op{
		Name:        "display.redshift.brightness.inc",
		Requires:    []string{"DISPLAY"},
		Run:         stepBrightness(1),
		Invalidates: []string{"display."},
	})

	daemon.Register(daemon.Op{
		Name:        "display.redshift.brightness.dec",
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
// Steps are coalesced, see coalescedStep.
func stepBrightness(direction float64) func(ctx context.Context, a daemon.Args) ([]string, error) {
	return func(ctx context.Context, a daemon.Args) ([]string, error) {
		step, err := a.Float("step")
		if err != nil {
			return nil, err
		}
		if err := validateStep(step); err != nil {
			return nil, err
		}

		return coalescedStep("brightness|"+a.Flag("display"), direction*step, func(total float64) ([]string, error) {
			displays, err := parseDisplayParam(a.Flag("display"))
			if err != nil {
				return nil, err
			}

			var lines []string
			for _, d := range displays {
				newBrightness := clampBrightnessToConfig(getLastSetBrightness(d) + total)
				if err := ApplyRedshift(d, -1, newBrightness, -1); err != nil {
					return nil, err
				}
				lines = append(lines, fmt.Sprintf("%.2f", newBrightness))
			}
			return lines, nil
		})
	}
}

// notifyDaemon tells a running daemon that persisted display state changed.
func notifyDaemon() {
	daemon.Notify("display.")
}

// renderFormat prints the numbers of a shared query with the --format of one watcher.
// The query itself runs with the default format.
func renderFormat(a daemon.Args, lines []string) []string {
	format := a.Flag("format")
	out := make([]string, len(lines))
	for i, line := range lines {
		value, err := strconv.ParseFloat(line, 64)
		switch {
		case err != nil:
			out[i] = line
		case format == "":
			out[i] = fmt.Sprint(value)
		default:
			out[i] = fmt.Sprintf(format, value)
		}
	}
	return out
}
