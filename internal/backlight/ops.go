package backlight

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
	"github.com/markusressel/system-control/internal/util"
)

const (
	displayLock  = "backlight-display"
	keyboardLock = "backlight-keyboard"

	// Backlights are changed by hardware keys and other tools without any event.
	defaultPollInterval = 2 * time.Second
)

func pollInterval(name string) func() time.Duration {
	return func() time.Duration {
		if configured := configuration.Get().Daemon.PollIntervals[name]; configured > 0 {
			return configured
		}
		return defaultPollInterval
	}
}

func percentArg(a daemon.Args) (int, error) {
	arg, err := a.Arg(0)
	if err != nil {
		return 0, err
	}
	p, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("invalid brightness %q: not an integer", arg)
	}
	return p, nil
}

func step(direction int, f func(int) error) func(context.Context, daemon.Args) ([]string, error) {
	return func(context.Context, daemon.Args) ([]string, error) {
		return nil, f(direction)
	}
}

func init() {
	displayInvalidates := []string{"display.backlight."}

	daemon.Register(daemon.Op{
		Name:  "display.backlight.brightness",
		Watch: func(daemon.Args) bool { return true },
		Poll:  pollInterval("backlight"),
		Run: func(context.Context, daemon.Args) ([]string, error) {
			p, err := displayPercent()
			if err != nil {
				return nil, err
			}
			return []string{fmt.Sprint(p)}, nil
		},
		Invalidates: displayInvalidates,
	})
	daemon.Register(daemon.Op{
		Name:        "display.backlight.brightness.inc",
		Lock:        displayLock,
		Run:         step(1, stepDisplay),
		Invalidates: displayInvalidates,
	})
	daemon.Register(daemon.Op{
		Name:        "display.backlight.brightness.dec",
		Lock:        displayLock,
		Run:         step(-1, stepDisplay),
		Invalidates: displayInvalidates,
	})
	daemon.Register(daemon.Op{
		Name: "display.backlight.brightness.set",
		Lock: displayLock,
		Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
			p, err := percentArg(a)
			if err != nil {
				return nil, err
			}
			return nil, setDisplay(p)
		},
		Invalidates: displayInvalidates,
	})

	keyboardInvalidates := []string{"keyboard."}

	// `keyboard brightness [value]` both queries and sets
	daemon.Register(daemon.Op{
		Name: "keyboard.brightness",
		LockFor: func(a daemon.Args) string {
			if len(a.Pos) > 0 {
				return keyboardLock
			}
			return ""
		},
		Watch: func(a daemon.Args) bool { return len(a.Pos) == 0 },
		Poll:  pollInterval("keyboard"),
		Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
			if len(a.Pos) > 0 {
				p, err := percentArg(a)
				if err != nil {
					return nil, err
				}
				_, err = util.SetKeyboardBrightness(p)
				return nil, err
			}
			b, err := util.GetKeyboardBrightness()
			if err != nil {
				return nil, err
			}
			return []string{fmt.Sprint(b)}, nil
		},
		Invalidates: keyboardInvalidates,
	})
	daemon.Register(daemon.Op{
		Name:        "keyboard.brightness.inc",
		Lock:        keyboardLock,
		Run:         step(1, stepKeyboard),
		Invalidates: keyboardInvalidates,
	})
	daemon.Register(daemon.Op{
		Name:        "keyboard.brightness.dec",
		Lock:        keyboardLock,
		Run:         step(-1, stepKeyboard),
		Invalidates: keyboardInvalidates,
	})
}
