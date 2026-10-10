package bluetooth

import (
	"context"
	"log"

	"github.com/markusressel/system-control/internal/bluetooth"
	"github.com/markusressel/system-control/internal/daemon"
)

func init() {
	daemon.Register(daemon.Op{
		Name:  "bluetooth.connected",
		Run:   runConnected,
		Watch: func(daemon.Args) bool { return true },
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

// EventSource refreshes watched bluetooth state whenever BlueZ reports a change.
// It is a daemon.Source.
func EventSource(ctx context.Context, refresh func(prefixes ...string)) {
	err := bluetooth.WatchChanges(ctx, func() { refresh("bluetooth.") })
	if err != nil {
		log.Printf("bluetooth: no change notifications, watched state will not update: %v", err)
	}
}
