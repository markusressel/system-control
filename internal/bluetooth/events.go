package bluetooth

import (
	"context"
	"errors"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/markusressel/system-control/internal/supervise"
)

// debounce merges the burst of signals BlueZ emits for one change.
const debounce = 50 * time.Millisecond

// watchOnce is the subscription WatchChanges keeps alive, a variable for tests.
var watchOnce = subscribeOnce

// watcher keeps the subscription alive.
var watcher = &supervise.Loop{Name: "bluetooth: change notifications"}

// WatchStatus describes the health of the subscription for `daemon status`.
func WatchStatus() []string {
	return watcher.Status()
}

// Watching reports whether changes are currently delivered, so that nothing has to poll.
func Watching() bool {
	return watcher.Active()
}

// WatchChanges calls onChange whenever a device or adapter of BlueZ changed (connected,
// paired, battery level, ...), until ctx is done. A lost connection to the system bus is
// established again with a backoff, followed by an onChange for what changed in between.
func WatchChanges(ctx context.Context, onChange func()) {
	watcher.Run(ctx, func(ctx context.Context, ready func()) error {
		return watchOnce(ctx, onChange, ready)
	}, onChange)
}

// subscribeOnce subscribes to the signals of BlueZ, calls started once that worked and
// handles the signals until ctx is done or the connection is lost.
func subscribeOnce(ctx context.Context, onChange func(), started func()) error {
	// a private connection, signals are delivered to every handler of a shared one
	bus, err := dbus.ConnectSystemBus()
	if err != nil {
		return err
	}
	defer bus.Close()

	matches := [][]dbus.MatchOption{
		{
			dbus.WithMatchSender(bluezService),
			dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
			dbus.WithMatchMember("PropertiesChanged"),
		},
		{
			dbus.WithMatchSender(bluezService),
			dbus.WithMatchInterface("org.freedesktop.DBus.ObjectManager"),
		},
	}
	for _, match := range matches {
		if err := bus.AddMatchSignalContext(ctx, match...); err != nil {
			return err
		}
	}

	signals := make(chan *dbus.Signal, 64)
	bus.Signal(signals)
	started()

	var timer *time.Timer
	var timerC <-chan time.Time
	for {
		select {
		case <-ctx.Done():
			return nil
		case signal, ok := <-signals:
			if !ok {
				return errors.New("lost connection to the system bus")
			}
			if !isRelevant(signal) {
				continue
			}
			if timer == nil {
				timer = time.NewTimer(debounce)
				timerC = timer.C
			}
		case <-timerC:
			timer, timerC = nil, nil
			onChange()
		}
	}
}

// isRelevant filters out changes of interfaces that don't affect the state of devices.
func isRelevant(signal *dbus.Signal) bool {
	if signal.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
		if len(signal.Body) == 0 {
			return false
		}
		changedInterface, _ := signal.Body[0].(string)
		switch changedInterface {
		case deviceInterface, adapterInterface, batteryInterface:
			return true
		}
		return false
	}
	return true
}
