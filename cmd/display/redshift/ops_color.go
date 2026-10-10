package redshift

import (
	"context"
	"fmt"

	"github.com/markusressel/system-control/internal/daemon"
	"github.com/markusressel/system-control/internal/util"
)

// minimum and maximum color temperature in Kelvin accepted when setting a value
const (
	minColorTemperature = 1000
	maxColorTemperature = 25000
)

func init() {
	// gamma
	register("display.redshift.gamma", withSetLock(isQuery("gamma", "-1")), true, runGamma)
	register("display.redshift.gamma.inc", constLock, false, stepGamma(1))
	register("display.redshift.gamma.dec", constLock, false, stepGamma(-1))
	register("display.redshift.gamma.reset", constLock, false, resetEach(func(d util.DisplayInfo) error {
		return ApplyRedshift(d, -1, -1, 1.0)
	}, "1"))

	// color temperature
	register("display.redshift.color-temperature", withSetLock(isQuery("temperature", "-1")), true, runColorTemperature)
	register("display.redshift.color-temperature.inc", constLock, false, stepColorTemperature(1))
	register("display.redshift.color-temperature.dec", constLock, false, stepColorTemperature(-1))
	register("display.redshift.color-temperature.reset", constLock, false, resetEach(func(d util.DisplayInfo) error {
		return ApplyRedshift(d, 6500, -1, -1)
	}, "6500"))

	register("display.redshift.brightness.reset", constLock, false, resetEach(func(d util.DisplayInfo) error {
		return ApplyRedshift(d, -1, 1.0, -1)
	}, "1"))
}

type lockPolicy func(a daemon.Args) string

// constLock holds the redshift lock for every invocation.
func constLock(daemon.Args) string { return redshiftLockName }

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
		return redshiftLockName
	}
}

// register adds a redshift op. A watchable op can only be watched while it just queries,
// which is decided by the lock policy (a set takes the lock).
func register(name string, lock lockPolicy, watchable bool, run func(context.Context, daemon.Args) ([]string, error)) {
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
	daemon.Register(op)
}

func runGamma(ctx context.Context, a daemon.Args) ([]string, error) {
	value, err := a.Float("gamma")
	if err != nil {
		return nil, err
	}
	displays, err := parseDisplayParam(a.Flag("display"))
	if err != nil {
		return nil, err
	}

	var lines []string
	for _, d := range displays {
		if value != -1 {
			newGamma := clampGammaToConfig(value)
			if newGamma != getLastSetGamma(d) {
				if err := ApplyRedshift(d, -1, -1, newGamma); err != nil {
					return nil, err
				}
			}
		}

		lastSet := getLastSetGamma(d)
		if format := a.Flag("format"); format != "" {
			lines = append(lines, fmt.Sprintf(format, lastSet))
		} else {
			lines = append(lines, fmt.Sprint(lastSet))
		}
	}
	return lines, nil
}

func stepGamma(direction float64) func(context.Context, daemon.Args) ([]string, error) {
	return func(ctx context.Context, a daemon.Args) ([]string, error) {
		step, err := a.Float("step")
		if err != nil {
			return nil, err
		}
		if step <= 0 {
			return nil, fmt.Errorf("step size must be a positive number (found: %f)", step)
		}
		displays, err := parseDisplayParam(a.Flag("display"))
		if err != nil {
			return nil, err
		}

		var lines []string
		for _, d := range displays {
			newGamma := clampGammaToConfig(getLastSetGamma(d) + direction*step)
			if err := ApplyRedshift(d, -1, -1, newGamma); err != nil {
				return nil, err
			}
			lines = append(lines, fmt.Sprint(newGamma))
		}
		return lines, nil
	}
}

func runColorTemperature(ctx context.Context, a daemon.Args) ([]string, error) {
	value, err := a.Int("temperature")
	if err != nil {
		return nil, err
	}
	if value != -1 && (value < minColorTemperature || value > maxColorTemperature) {
		return nil, fmt.Errorf("color temperature must be between %d and %d", minColorTemperature, maxColorTemperature)
	}
	displays, err := parseDisplayParam(a.Flag("display"))
	if err != nil {
		return nil, err
	}

	var lines []string
	for _, d := range displays {
		if len(displays) > 1 {
			lines = append(lines, fmt.Sprintf("Display: %s", d))
		}

		if value != -1 {
			newTemperature := clampColorTemperatureToConfig(int64(value))
			if newTemperature != getLastSetColorTemperature(d) {
				if err := ApplyRedshift(d, newTemperature, -1, -1); err != nil {
					return nil, err
				}
			}
		}
		lines = append(lines, fmt.Sprint(getLastSetColorTemperature(d)))
	}
	return lines, nil
}

func stepColorTemperature(direction int64) func(context.Context, daemon.Args) ([]string, error) {
	return func(ctx context.Context, a daemon.Args) ([]string, error) {
		step, err := a.Int("step")
		if err != nil {
			return nil, err
		}
		if step <= 0 {
			return nil, fmt.Errorf("step size must be a positive number (found: %d)", step)
		}
		displays, err := parseDisplayParam(a.Flag("display"))
		if err != nil {
			return nil, err
		}

		var lines []string
		for _, d := range displays {
			newTemperature := clampColorTemperatureToConfig(getLastSetColorTemperature(d) + direction*int64(step))
			if err := ApplyRedshift(d, newTemperature, -1, -1); err != nil {
				return nil, err
			}
			lines = append(lines, fmt.Sprint(newTemperature))
		}
		return lines, nil
	}
}

// resetEach applies a reset to every selected display and prints `printed` for each.
func resetEach(reset func(util.DisplayInfo) error, printed string) func(context.Context, daemon.Args) ([]string, error) {
	return func(ctx context.Context, a daemon.Args) ([]string, error) {
		displays, err := parseDisplayParam(a.Flag("display"))
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
