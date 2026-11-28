package redshift

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
	"github.com/markusressel/system-control/internal/util"
)

const defaultPollInterval = 5 * time.Second

// displayPollInterval is how often watched display state is re-queried. Changes made through
// this tool refresh the watchers right away, polling only catches the rest.
var displayPollInterval = configuration.PollInterval("display", defaultPollInterval)

// formatOutput makes the --format of a value op apply per watcher.
func formatOutput(op *daemon.Op) {
	op.OutputFlags = []string{"format"}
	op.Render = renderFormat
}

func init() {
	register("display.redshift.brightness", withSetLock(isQuery("brightness", "-1")), true, runBrightness, formatOutput)
	register("display.redshift.brightness.inc", noLock, false, brightnessChannel.stepOp(1))
	register("display.redshift.brightness.dec", noLock, false, brightnessChannel.stepOp(-1))
	register("display.redshift.brightness.reset", constLock, false, resetEach(func(d util.DisplayInfo) error {
		return Apply(d, -1, 1.0, -1)
	}, "1"))

	register("display.redshift.gamma", withSetLock(isQuery("gamma", "-1")), true, runGamma, formatOutput)
	register("display.redshift.gamma.inc", noLock, false, gammaChannel.stepOp(1))
	register("display.redshift.gamma.dec", noLock, false, gammaChannel.stepOp(-1))
	register("display.redshift.gamma.reset", constLock, false, resetEach(func(d util.DisplayInfo) error {
		return Apply(d, -1, -1, 1.0)
	}, "1"))

	register("display.redshift.color-temperature", withSetLock(isQuery("temperature", "-1")), true, runColorTemperature)
	register("display.redshift.color-temperature.inc", noLock, false, colorTemperatureChannel.stepOp(1))
	register("display.redshift.color-temperature.dec", noLock, false, colorTemperatureChannel.stepOp(-1))
	register("display.redshift.color-temperature.reset", constLock, false, resetEach(func(d util.DisplayInfo) error {
		return Apply(d, 6500, -1, -1)
	}, "6500"))
}

type lockPolicy func(a daemon.Args) string

// constLock holds the redshift lock for every invocation.
func constLock(daemon.Args) string { return LockName }

// noLock is for ops that coalesce and take the lock themselves, see coalescedStep.
func noLock(daemon.Args) string { return "" }

// isQuery reports whether the flag still has its "not set" value.
func isQuery(flag, unset string) func(daemon.Args) bool {
	return func(a daemon.Args) bool { return a.Flag(flag) == unset }
}

// withSetLock locks only invocations that change something.
func withSetLock(query func(daemon.Args) bool) lockPolicy {
	return func(a daemon.Args) string {
		if query(a) {
			return ""
		}
		return LockName
	}
}

// register adds a redshift op. A watchable op can only be watched while it just queries,
// which is decided by the lock policy (a set takes the lock).
func register(name string, lock lockPolicy, watchable bool, run func(context.Context, daemon.Args) ([]string, error), opts ...func(*daemon.Op)) {
	op := daemon.Op{
		Name:        name,
		Requires:    []string{"DISPLAY"},
		LockFor:     lock,
		Run:         run,
		Invalidates: []string{"display."},
	}
	if watchable {
		op.Watch = func(a daemon.Args) bool { return lock(a) == "" }
		op.Poll = displayPollInterval
	}
	for _, opt := range opts {
		opt(&op)
	}
	daemon.Register(op)
}

// channel is one of the values redshift sets (brightness, gamma, color temperature), for the
// relative steps of its inc and dec ops.
type channel struct {
	// name keys the coalesced steps
	name string
	// step reads and checks the --step of an invocation
	step func(a daemon.Args) (float64, error)
	// apply changes the last value of the display by delta and returns the new value as printed
	apply func(d util.DisplayInfo, delta float64) (string, error)
}

var brightnessChannel = channel{
	name: "brightness",
	step: func(a daemon.Args) (float64, error) {
		step, err := a.Float("step")
		if err != nil {
			return 0, err
		}
		return step, validateStep(step)
	},
	apply: func(d util.DisplayInfo, delta float64) (string, error) {
		value := ClampBrightnessToConfig(LastBrightness(d) + delta)
		return fmt.Sprintf("%.2f", value), Apply(d, -1, value, -1)
	},
}

var gammaChannel = channel{
	name: "gamma",
	step: func(a daemon.Args) (float64, error) {
		step, err := a.Float("step")
		if err == nil && step <= 0 {
			err = fmt.Errorf("step size must be a positive number (found: %f)", step)
		}
		return step, err
	},
	apply: func(d util.DisplayInfo, delta float64) (string, error) {
		value := ClampGammaToConfig(LastGamma(d) + delta)
		return fmt.Sprint(value), Apply(d, -1, -1, value)
	},
}

var colorTemperatureChannel = channel{
	name: "color-temperature",
	step: func(a daemon.Args) (float64, error) {
		step, err := a.Int("step")
		if err == nil && step <= 0 {
			err = fmt.Errorf("step size must be a positive number (found: %d)", step)
		}
		return float64(step), err
	},
	apply: func(d util.DisplayInfo, delta float64) (string, error) {
		value := ClampColorTemperatureToConfig(LastColorTemperature(d) + int64(math.Round(delta)))
		return fmt.Sprint(value), Apply(d, value, -1, -1)
	},
}

// stepOp returns the op implementation that changes the channel by --step in the direction.
// Steps are coalesced, see coalescedStep.
func (c channel) stepOp(direction float64) func(context.Context, daemon.Args) ([]string, error) {
	return func(ctx context.Context, a daemon.Args) ([]string, error) {
		step, err := c.step(a)
		if err != nil {
			return nil, err
		}

		return coalescedStep(c.name+"|"+a.Flag("display"), direction*step, func(total float64) ([]string, error) {
			displays, err := ParseDisplays(a.Flag("display"))
			if err != nil {
				return nil, err
			}

			var lines []string
			for _, d := range displays {
				line, err := c.apply(d, total)
				if err != nil {
					return nil, err
				}
				lines = append(lines, line)
			}
			return lines, nil
		})
	}
}

// validateStep checks the --step of a brightness change.
func validateStep(step float64) error {
	if step <= 0 {
		return fmt.Errorf("step size must be a positive number (found: %f)", step)
	}
	if step > 1.0 {
		return fmt.Errorf("step size cannot exceed 1.0 (found: %f)", step)
	}
	return nil
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
	displays, err := ParseDisplays(displayName)
	if err != nil {
		return "", err
	}
	if len(displays) == 0 {
		return "", fmt.Errorf("no displays found")
	}

	var lastSet float64
	for _, d := range displays {
		lastSet = LastBrightness(d)
	}
	return fmt.Sprintf(format, lastSet), nil
}

// setBrightness is run with the redshift lock held, see the op registration.
func setBrightness(displayName string, value float64, format string) ([]string, error) {
	displays, err := ParseDisplays(displayName)
	if err != nil {
		return nil, err
	}

	var lastSet float64
	for _, d := range displays {
		newBrightness := ClampBrightnessToConfig(value)
		if newBrightness != LastBrightness(d) {
			if err := Apply(d, -1, newBrightness, -1); err != nil {
				return nil, err
			}
		}
		lastSet = LastBrightness(d)
	}
	return []string{fmt.Sprintf(format, lastSet)}, nil
}

func runGamma(ctx context.Context, a daemon.Args) ([]string, error) {
	value, err := a.Float("gamma")
	if err != nil {
		return nil, err
	}
	displays, err := ParseDisplays(a.Flag("display"))
	if err != nil {
		return nil, err
	}

	var lines []string
	for _, d := range displays {
		if value != -1 {
			newGamma := ClampGammaToConfig(value)
			if newGamma != LastGamma(d) {
				if err := Apply(d, -1, -1, newGamma); err != nil {
					return nil, err
				}
			}
		}

		lastSet := LastGamma(d)
		if format := a.Flag("format"); format != "" {
			lines = append(lines, fmt.Sprintf(format, lastSet))
		} else {
			lines = append(lines, fmt.Sprint(lastSet))
		}
	}
	return lines, nil
}

func runColorTemperature(ctx context.Context, a daemon.Args) ([]string, error) {
	value, err := a.Int("temperature")
	if err != nil {
		return nil, err
	}
	if value != -1 && (value < MinColorTemperature || value > MaxColorTemperature) {
		return nil, fmt.Errorf("color temperature must be between %d and %d", MinColorTemperature, MaxColorTemperature)
	}
	displays, err := ParseDisplays(a.Flag("display"))
	if err != nil {
		return nil, err
	}

	var lines []string
	for _, d := range displays {
		if len(displays) > 1 {
			lines = append(lines, fmt.Sprintf("Display: %s", d))
		}

		if value != -1 {
			newTemperature := ClampColorTemperatureToConfig(int64(value))
			if newTemperature != LastColorTemperature(d) {
				if err := Apply(d, newTemperature, -1, -1); err != nil {
					return nil, err
				}
			}
		}
		lines = append(lines, fmt.Sprint(LastColorTemperature(d)))
	}
	return lines, nil
}

// resetEach applies a reset to every selected display and prints `printed` for each.
func resetEach(reset func(util.DisplayInfo) error, printed string) func(context.Context, daemon.Args) ([]string, error) {
	return func(ctx context.Context, a daemon.Args) ([]string, error) {
		displays, err := ParseDisplays(a.Flag("display"))
		if err != nil {
			return nil, err
		}
		var lines []string
		for _, d := range displays {
			if err := reset(d); err != nil {
				return nil, err
			}
			lines = append(lines, printed)
		}
		return lines, nil
	}
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
