package audio

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
	"github.com/markusressel/system-control/internal/util"
)

type DaemonModule struct {
	trackers []*daemon.PollingTracker
}

func NewDaemonModule() *DaemonModule {
	m := &DaemonModule{
		trackers: make([]*daemon.PollingTracker, 0),
	}

	pipewire.StartMonitor()
	pipewire.RegisterUpdateCallback(func() {
		m.Wakeup()
	})

	return m
}

func (m *DaemonModule) Name() string {
	return "audio"
}

func (m *DaemonModule) getInterval() time.Duration {
	if configured, ok := configuration.CurrentConfig.Daemon.PollIntervals["audio"]; ok && configured > 0 {
		return configured
	}
	return 1 * time.Second
}

func (m *DaemonModule) Track(ctx context.Context, command string, flags map[string]string, updates chan<- daemon.StateUpdate) error {
	device := flags["device"]
	fullCmd := "audio." + command
	path := daemon.GenerateStatePath(fullCmd, flags)

	switch command {
	case "volume":
		tracker := daemon.NewPollingTracker(m.getInterval(), path, updates, func() (string, error) {
			state := pipewire.PwDump()
			volume, err := state.GetVolumeByName(device)
			if err != nil {
				return "", err
			}
			volume = util.RoundToTwoDecimals(volume)

			if device == "" {
				target, _ := state.GetDefaultSinkNode()
				audioLock.Lock()
				cachedVolume = volume
				if target.Id != 0 {
					cachedTargetId = target.Id
				}
				audioLock.Unlock()
			}

			return fmt.Sprintf("%d", (int)(volume*100)), nil
		})
		m.trackers = append(m.trackers, tracker)
		go tracker.Start(ctx)
		return nil

	case "muted":
		tracker := daemon.NewPollingTracker(m.getInterval(), path, updates, func() (string, error) {
			state := pipewire.PwDump()
			node, err := state.GetDefaultSinkNode()
			if err != nil {
				return "", err
			}
			muted, err := state.IsMuted(node.Id)
			if err != nil {
				return "", err
			}
			if muted {
				return "yes", nil
			}
			return "no", nil
		})
		m.trackers = append(m.trackers, tracker)
		go tracker.Start(ctx)
		return nil

	case "sink.active":
		tracker := daemon.NewPollingTracker(m.getInterval(), path, updates, func() (string, error) {
			state := pipewire.PwDump()
			node, err := state.GetDefaultSinkNode()
			if err != nil {
				return "", err
			}

			// Polybar often uses --columns description
			if flags["columns"] == "description" {
				desc, err := node.GetDescription()
				if err == nil {
					return desc, nil
				}
			}

			name, err := node.GetName()
			if err != nil {
				return "", err
			}
			return name, nil
		})
		m.trackers = append(m.trackers, tracker)
		go tracker.Start(ctx)
		return nil

	case "device.profile":
		tracker := daemon.NewPollingTracker(m.getInterval(), path, updates, func() (string, error) {
			state := pipewire.PwDump()
			deviceName := flags["device"]

			deviceNode, err := state.FindDeviceByName(deviceName)
			if err != nil {
				return "", err
			}

			profile, err := deviceNode.GetActiveProfile()
			if err != nil {
				return "", err
			}

			return profile.Description, nil
		})
		m.trackers = append(m.trackers, tracker)
		go tracker.Start(ctx)
		return nil

	default:
		return fmt.Errorf("unknown audio command: %s", command)
	}
}

func (m *DaemonModule) Wakeup() {
	for _, t := range m.trackers {
		t.Wakeup()
	}
}

var audioLock sync.Mutex
var cachedVolume float64 = -1.0
var cachedTargetId int = -1

func (m *DaemonModule) Execute(ctx context.Context, args []string) error {
	if len(args) >= 2 && args[0] == "sink" {
		if args[1] == "next" {
			err := pipewire.RotateActiveSinkPipewire(false)
			m.Wakeup() // Trigger UI update immediately
			return err
		}
		if args[1] == "previous" {
			err := pipewire.RotateActiveSinkPipewire(true)
			m.Wakeup() // Trigger UI update immediately
			return err
		}
		if args[1] == "switch" && len(args) >= 3 {
			searchString := args[2]
			state := pipewire.PwDump()
			nodes := state.FindNodesByName(searchString)
			var audioSinkNodes []pipewire.InterfaceNode
			for _, node := range nodes {
				if mc, _ := node.GetMediaClass(); mc == pipewire.MediaClassAudioSink {
					audioSinkNodes = append(audioSinkNodes, node)
				}
			}
			if len(audioSinkNodes) <= 0 {
				return fmt.Errorf("no sink found")
			}
			if len(audioSinkNodes) > 1 {
				return fmt.Errorf("ambiguous sink name")
			}
			err := state.SwitchSinkTo(audioSinkNodes[0])
			m.Wakeup()
			return err
		}
	}

	if len(args) < 2 {
		return fmt.Errorf("insufficient arguments")
	}

	subcommand := args[0]
	action := args[1]

	if subcommand == "volume" {
		audioLock.Lock()
		defer audioLock.Unlock()

		if queuedAt, ok := ctx.Value("queuedAt").(time.Time); ok {
			if time.Since(queuedAt) > 200*time.Millisecond {
				return nil
			}
		}

		if action == "inc" || action == "dec" {
			if cachedTargetId == -1 || cachedVolume == -1.0 {
				state := pipewire.PwDump()
				target, err := state.GetDefaultSinkNode()
				if err != nil {
					return err
				}
				cachedTargetId = target.Id
				volume, err := state.GetVolumeByName("")
				if err != nil {
					return err
				}
				cachedVolume = util.RoundToTwoDecimals(volume)
			}

			start := time.Now()
			change := CalculateAppropriateVolumeChange(cachedVolume*100, action == "inc") / 100.0
			targetVolume := cachedVolume
			if action == "inc" {
				targetVolume += change
			} else {
				targetVolume -= change
			}
			if targetVolume < 0.0 {
				targetVolume = 0.0
			}

			// Optimistically update cache so the next rapid scroll uses the updated target
			cachedVolume = util.RoundToTwoDecimals(targetVolume)

			t1 := time.Now()
			err := pipewire.WpCtlSetVolume(cachedTargetId, cachedVolume)
			t2 := time.Now()
			if err == nil {
				m.Wakeup() // Immediately trigger update
			}
			t3 := time.Now()
			fmt.Printf("Execute timings: prep %v, wpctl %v, wakeup %v\n", t1.Sub(start), t2.Sub(t1), t3.Sub(t2))
			return err
		} else if action == "toggle-mute" || action == "mute" || action == "unmute" {
			state := pipewire.PwDump()
			target, err := state.GetDefaultSinkNode()
			if err != nil {
				return err
			}
			if action == "toggle-mute" {
				err = pipewire.WpCtlToggleMute(target.Id)
			} else if action == "mute" {
				err = pipewire.WpCtlSetMute(target.Id, true)
			} else if action == "unmute" {
				err = pipewire.WpCtlSetMute(target.Id, false)
			}

			if err == nil {
				m.Wakeup()
			}
			return err
		}
	}
	return fmt.Errorf("unsupported command: %v", args)
}
