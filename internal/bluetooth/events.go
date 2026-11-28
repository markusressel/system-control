package bluetooth

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"time"

	"github.com/godbus/dbus/v5"
)

// debounce merges the burst of signals BlueZ emits for one change.
const debounce = 50 * time.Millisecond

// restart delays of a lost subscription, variables for tests
var (
	watchRestartDelay = time.Second
	watchMaxDelay     = 30 * time.Second
	watchHealthyAfter = time.Minute
)

// watchOnce is the subscription WatchChanges keeps alive, a variable for tests.
var watchOnce = subscribeOnce

var watching atomic.Bool

// Watching reports whether changes are currently delivered, so that nothing has to poll.
func Watching() bool {
	return watching.Load()
}

// WatchChanges calls onChange whenever a device or adapter of BlueZ changed (connected,
// paired, battery level, ...), until ctx is done. A lost connection to the system bus is
// established again with a backoff, followed by an onChange for what changed in between.
func WatchChanges(ctx context.Context, onChange func()) {
	delay := watchRestartDelay
	loggedFailure := false
	for ctx.Err() == nil {
		started := time.Now()
		err := watchOnce(ctx, onChange, func() {
			watching.Store(true)
			if loggedFailure {
				log.Printf("bluetooth: watching changes again")
				loggedFailure = false
				onChange()
			}
		})
		watching.Store(false)
		if ctx.Err() != nil {
			return
		}
		if !loggedFailure {
			log.Printf("bluetooth: no change notifications, watched state is polled: %v", err)
			loggedFailure = true
		}
		if time.Since(started) > watchHealthyAfter {
			delay = watchRestartDelay
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
		delay = min(delay*2, watchMaxDelay)
	}
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
