package bluetooth

import (
	"context"
	"log"
	"time"

	"github.com/godbus/dbus/v5"
)

// debounce merges the burst of signals BlueZ emits for one change.
const debounce = 50 * time.Millisecond

// WatchChanges calls onChange whenever a device or adapter of BlueZ changed
// (connected, paired, battery level, ...), until ctx is done. It returns once the
// subscription is established, the signals are handled in the background.
func WatchChanges(ctx context.Context, onChange func()) error {
	// a private connection, signals are delivered to every handler of a shared one
	bus, err := dbus.ConnectSystemBus()
	if err != nil {
		return err
	}

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
			_ = bus.Close()
			return err
		}
	}

	signals := make(chan *dbus.Signal, 64)
	bus.Signal(signals)

	go func() {
		defer bus.Close()
		var timer *time.Timer
		var timerC <-chan time.Time
		for {
			select {
			case <-ctx.Done():
				return
			case signal, ok := <-signals:
				if !ok {
					log.Printf("bluetooth: lost connection to the system bus")
					return
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
	}()

	return nil
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
