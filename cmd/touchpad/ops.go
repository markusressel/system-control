package touchpad

import (
	"context"
	"strconv"

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
		enabled, err := strconv.ParseBool(arg)
		if err != nil {
			return err
		}
		return util.SetTouchpadEnabled(enabled)
	})
	register("touchpad.toggle", func(daemon.Args) error {
		enabled, err := util.IsTouchpadEnabled()
		if err != nil {
			return err
		}
		return util.SetTouchpadEnabled(!enabled)
	})
	register("touchpad.toggleVerticalScrollDirection", func(daemon.Args) error {
		return util.ToggleTouchpadScrollVerticalDirection()
	})
}
