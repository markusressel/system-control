package bluetooth

import (
	"context"
	"time"

	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/markusressel/system-control/internal/daemon"
)

func init() {
	daemon.Register(daemon.Op{
		Name:  "bluetooth.connected",
		Run:   runConnected,
		Watch: func(daemon.Args) bool { return true },
		Poll:  pollInterval,
	})
}

// runConnected answers "yes" or "no" for the device matching the first argument.
func runConnected(ctx context.Context, a daemon.Args) ([]string, error) {
	name, err := a.Arg(0)
	if err != nil {
		return nil, err
	}
	devices, err := bluetooth.GetBluetoothDevices()
	if err != nil {
		return nil, err
	}
	matchingDevices := findBluetoothDeviceFuzzy(name, devices)
	if len(matchingDevices) > 1 {
		return nil, multipleDevicesError(matchingDevices)
	}
	if len(matchingDevices) == 1 && matchingDevices[0].Connected {
		return []string{"yes"}, nil
	}
	return []string{"no"}, nil
}

// unwatchedPollInterval is how often watched bluetooth state is queried while BlueZ's
// signals can't be received.
const unwatchedPollInterval = 5 * time.Second

// pollInterval polls only while changes are not delivered (0 disables polling).
func pollInterval() time.Duration {
	if bluetooth.Watching() {
		return 0
	}
	return unwatchedPollInterval
}

// EventSource refreshes watched bluetooth state whenever BlueZ reports a change.
// It is a daemon.Source.
func EventSource(ctx context.Context, refresh func(prefixes ...string)) {
	go bluetooth.WatchChanges(ctx, func() { refresh("bluetooth.") })
}
