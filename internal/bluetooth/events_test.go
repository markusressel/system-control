package bluetooth

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
)

func TestIsRelevant(t *testing.T) {
	changed := func(iface string) *dbus.Signal {
		return &dbus.Signal{Name: "org.freedesktop.DBus.Properties.PropertiesChanged", Body: []interface{}{iface}}
	}

	assert.True(t, isRelevant(changed(deviceInterface)))
	assert.True(t, isRelevant(changed(batteryInterface)))
	assert.False(t, isRelevant(changed("org.bluez.MediaTransport1")))
	assert.False(t, isRelevant(&dbus.Signal{Name: "org.freedesktop.DBus.Properties.PropertiesChanged"}))
	assert.True(t, isRelevant(&dbus.Signal{Name: "org.freedesktop.DBus.ObjectManager.InterfacesAdded"}))
}

// A lost connection is established again, and what changed in between is picked up.
func TestWatchChanges_ReconnectsAfterALostConnection(t *testing.T) {
	previous, delay := watchOnce, watchRestartDelay
	watchRestartDelay = 10 * time.Millisecond
	t.Cleanup(func() { watchOnce, watchRestartDelay = previous, delay })

	var runs atomic.Int64
	reconnected := make(chan struct{})
	watchOnce = func(ctx context.Context, onChange func(), started func()) error {
		started()
		if runs.Add(1) == 1 {
			return errors.New("lost connection to the system bus")
		}
		close(reconnected)
		<-ctx.Done()
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	var changes atomic.Int64
	done := make(chan struct{})
	go func() { WatchChanges(ctx, func() { changes.Add(1) }); close(done) }()

	select {
	case <-reconnected:
	case <-time.After(5 * time.Second):
		t.Fatal("no reconnect")
	}
	if changes.Load() != 1 {
		t.Errorf("%d changes after the reconnect, want 1 to catch up", changes.Load())
	}
	if !Watching() {
		t.Error("Watching() = false after the reconnect")
	}

	cancel()
	<-done
	if Watching() {
		t.Error("Watching() = true after WatchChanges returned")
	}
}
