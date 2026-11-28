package touchpad

import (
	"context"

	"github.com/markusressel/system-control/internal/daemon"
	"github.com/markusressel/system-control/internal/util"
)

const lockName = "touchpad"

func register(name string, run func(daemon.Args) error) {
	daemon.Register(daemon.Op{
		Name:     name,
		Lock:     lockName,
		Requires: []string{"DISPLAY"},
		Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
			return nil, run(a)
		},
	})
}

func init() {
	register("touchpad.on", func(daemon.Args) error { return util.SetTouchpadEnabled(true) })
	register("touchpad.off", func(daemon.Args) error { return util.SetTouchpadEnabled(false) })
	register("touchpad.set", func(a daemon.Args) error {
		arg, err := a.Arg(0)
		if err != nil {
			return err
		}
		return SetArg(arg)
	})
	register("touchpad.toggle", func(daemon.Args) error { return Toggle() })
	register("touchpad.toggleVerticalScrollDirection", func(daemon.Args) error {
		return util.ToggleTouchpadScrollVerticalDirection()
	})
}
