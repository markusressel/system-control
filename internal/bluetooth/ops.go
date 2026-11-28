package bluetooth

import (
	"context"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
)

func init() {
	daemon.RegisterSource(eventSource, WatchStatus)

	daemon.Register(daemon.Op{
		Name:  "bluetooth.connected",
		Run:   runConnected,
		Watch: daemon.Always,
		Poll:  pollInterval,
	})
}

// runConnected answers "yes" or "no" for the device matching the first argument.
func runConnected(ctx context.Context, a daemon.Args) ([]string, error) {
	name, err := a.Arg(0)
	if err != nil {
		return nil, err
	}
	devices, err := GetBluetoothDevices()
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

var configuredPollInterval = configuration.PollInterval("bluetooth", 0)

// pollInterval polls while changes are not delivered, and at a configured interval as a safety
// net (0 disables polling).
func pollInterval() time.Duration {
	if configured := configuredPollInterval(); configured > 0 {
		return configured
	}
	if Watching() {
		return 0
	}
	return unwatchedPollInterval
}

// eventSource refreshes watched bluetooth state whenever BlueZ reports a change.
func eventSource(ctx context.Context, refresh func(prefixes ...string)) {
	go WatchChanges(ctx, func() { refresh("bluetooth.") })
}
