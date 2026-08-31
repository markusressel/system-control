package redshift

import (
	"context"
	"fmt"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
)

type DaemonModule struct {
	trackers []*daemon.PollingTracker
}

func NewDaemonModule() *DaemonModule {
	return &DaemonModule{
		trackers: make([]*daemon.PollingTracker, 0),
	}
}

func (m *DaemonModule) Name() string {
	return "display"
}

func (m *DaemonModule) getInterval() time.Duration {
	if configured, ok := configuration.CurrentConfig.Daemon.PollIntervals["display"]; ok && configured > 0 {
		return configured
	}
	return 1 * time.Second
}

func (m *DaemonModule) Track(ctx context.Context, command string, flags map[string]string, updates chan<- daemon.StateUpdate) error {
	device := flags["display"]
	if device == "" {
		device = flags["d"] // short flag
	}

	// Reconstruct the full command for GenerateStatePath
	fullCmd := "display." + command
	path := daemon.GenerateStatePath(fullCmd, flags)

	switch command {
	case "redshift.brightness":
		tracker := daemon.NewPollingTracker(m.getInterval(), path, updates, func() (string, error) {
			redshiftLock := lockRedshift()
			defer redshiftLock.Unlock()

			dInfo, err := parseDisplayParam(device)
			if err != nil {
				return "", err
			}
			if len(dInfo) == 0 {
				return "", fmt.Errorf("no displays found")
			}

			format := flags["format"]
			if format == "" {
				format = flags["f"]
			}
			if format == "" {
				format = "%.2f"
			}

			lastSet := getLastSetBrightness(dInfo[0])
			return fmt.Sprintf(format, lastSet), nil
		})
		m.trackers = append(m.trackers, tracker)
		go tracker.Start(ctx)
		return nil

	default:
		return fmt.Errorf("unknown redshift command: %s", command)
	}
}

func (m *DaemonModule) Wakeup() {
	for _, t := range m.trackers {
		t.Wakeup()
	}
}
