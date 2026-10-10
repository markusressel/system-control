package redshift

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
		Name: "display.redshift.brightness",
		Run:  runBrightness,
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
		Run:         stepBrightness(1),
		Invalidates: []string{"display."},
	})

	daemon.Register(daemon.Op{
		Name:        "display.redshift.brightness.dec",
		Lock:        redshiftLockName,
		Run:         stepBrightness(-1),
		Invalidates: []string{"display."},
	})
}

// prepare must be called by every op before it touches the displays.
//
// The daemon is started by the user service manager, possibly before the session imported
// DISPLAY into it. The daemon then picks it up from the manager, and if there is none the
// client has to do the work, it has DISPLAY.
func prepare() error {
	if !daemon.Running() || os.Getenv("DISPLAY") != "" {
		return nil
	}

	importSessionEnvironment()
	if os.Getenv("DISPLAY") == "" {
		return daemon.ErrUnhandled
	}
	return nil
}

// importSessionEnvironment copies the X11 variables of the user service manager into this process.
func importSessionEnvironment() {
	out, err := exec.Command("systemctl", "--user", "show-environment").Output()
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(out), "\n") {
		name, value, ok := strings.Cut(line, "=")
		if ok && (name == "DISPLAY" || name == "XAUTHORITY") {
			_ = os.Setenv(name, value)
		}
	}
}

// reloadConfig makes a long running daemon pick up edits of the configuration file.
// Must be called with the redshift lock held, the configuration is read by the clamping below.
func reloadConfig() {
	if daemon.Running() {
		_ = configuration.Reload()
	}
}

func runBrightness(ctx context.Context, a daemon.Args) ([]string, error) {
	if err := prepare(); err != nil {
		return nil, err
	}

	value, err := strconv.ParseFloat(a.Flag("brightness"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid brightness: %w", err)
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

func setBrightness(displayName string, value float64, format string) ([]string, error) {
	lock, err := acquireRedshiftLock()
	if err != nil {
		return nil, err
	}
	defer lock.Unlock()
	reloadConfig()

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
		if err := prepare(); err != nil {
			return nil, err
		}

		step, err := strconv.ParseFloat(a.Flag("step"), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid step: %w", err)
		}
		if err := validateStep(step); err != nil {
			return nil, err
		}
		reloadConfig()

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
