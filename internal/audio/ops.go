package audio

import (
	"context"
	"fmt"
	"strconv"

	"github.com/markusressel/system-control/internal/audio/pipewire"
	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
)

const volumeLock = "cmd-audio-volume"

// service is the audio service of this process: the daemon's, or the one of a single CLI run.
var service = NewService(pipewireBackend())

// audio is driven by PipeWire events, polling is off unless configured as a safety net
var pollAudio = configuration.PollInterval("audio", 0)

func init() {
	daemon.Register(daemon.Op{
		Name: "audio.volume",
		Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
			v, err := service.Volume(a.Flag("device"), a.Flag("stream"))
			return []string{v}, err
		},
		Watch: daemon.Always,
		Poll:  pollAudio,
	})

	daemon.Register(daemon.Op{
		Name: "audio.volume.muted",
		Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
			v, err := service.Muted(a.Flag("device"), a.Flag("stream"))
			return []string{v}, err
		},
		Watch: daemon.Always,
		Poll:  pollAudio,
	})

	stepOp := func(name string, increase bool) daemon.Op {
		return daemon.Op{
			Name: name,
			Lock: volumeLock,
			Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
				return service.Step(a.Flag("device"), a.Flag("stream"), increase)
			},
			Invalidates: []string{"audio."},
		}
	}
	daemon.Register(stepOp("audio.volume.inc", true))
	daemon.Register(stepOp("audio.volume.dec", false))

	daemon.Register(daemon.Op{
		Name: "audio.volume.set",
		Lock: volumeLock,
		Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
			arg, err := a.Arg(0)
			if err != nil {
				return nil, err
			}
			percent, err := strconv.Atoi(arg)
			if err != nil {
				return nil, fmt.Errorf("invalid volume %q: not an integer", arg)
			}
			return service.Set(a.Flag("device"), a.Flag("stream"), percent)
		},
		Invalidates: []string{"audio."},
	})

	muteOp := func(name string, action MuteAction) daemon.Op {
		return daemon.Op{
			Name: name,
			Lock: volumeLock,
			Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
				return service.Mute(a.Flag("device"), a.Flag("stream"), action)
			},
			Invalidates: []string{"audio."},
		}
	}
	daemon.Register(muteOp("audio.volume.mute", MuteOn))
	daemon.Register(muteOp("audio.volume.unmute", MuteOff))
	daemon.Register(muteOp("audio.volume.toggle-mute", MuteToggle))

	daemon.Register(daemon.Op{
		Name: "audio.volume.save",
		Lock: volumeLock,
		Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
			return nil, service.Save(a.Flag("device"), a.Flag("stream"))
		},
	})

	daemon.Register(daemon.Op{
		Name: "audio.volume.restore",
		Lock: volumeLock,
		Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
			return service.Restore(a.Flag("device"), a.Flag("stream"))
		},
		Invalidates: []string{"audio."},
	})

	daemon.Register(daemon.Op{
		Name: "audio.sink.next",
		Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
			return nil, pipewire.RotateActiveSinkPipewire(false)
		},
		Invalidates: []string{"audio."},
	})

	daemon.Register(daemon.Op{
		Name: "audio.sink.previous",
		Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
			return nil, pipewire.RotateActiveSinkPipewire(true)
		},
		Invalidates: []string{"audio."},
	})

	daemon.Register(daemon.Op{
		Name: "audio.sink.switch",
		Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
			name, err := a.Arg(0)
			if err != nil {
				return nil, err
			}
			return nil, SwitchSink(name)
		},
		Invalidates: []string{"audio."},
	})

	daemon.Register(daemon.Op{
		Name:  "audio.sink.active",
		Run:   runSinkActive,
		Watch: daemon.Always,
		Poll:  pollAudio,
	})

	daemon.Register(daemon.Op{
		Name: "audio.source.switch",
		Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
			name, err := a.Arg(0)
			if err != nil {
				return nil, err
			}
			return nil, SwitchSource(name)
		},
		Invalidates: []string{"audio."},
	})

	daemon.Register(daemon.Op{
		Name: "audio.source.active",
		Run: func(ctx context.Context, a daemon.Args) ([]string, error) {
			state := pipewire.PwDump()
			node, err := state.GetDefaultSourceNode()
			if err != nil {
				return nil, err
			}
			return nodeColumns(node, a.List("columns"))
		},
		Watch: daemon.Always,
		Poll:  pollAudio,
	})

	daemon.Register(daemon.Op{
		Name:        "audio.device.profile",
		Run:         runDeviceProfile,
		Watch:       func(a daemon.Args) bool { return len(a.Pos) == 0 },
		Poll:        pollAudio,
		Invalidates: []string{"audio."},
	})
}

// runSinkActive prints the requested columns of the default sink, or with an argument whether
// that sink is the default one.
func runSinkActive(ctx context.Context, a daemon.Args) ([]string, error) {
	state := pipewire.PwDump()
	if len(a.Pos) > 0 {
		return []string{fmt.Sprint(state.ContainsActiveSink(a.Pos[0]))}, nil
	}
	node, err := state.GetDefaultSinkNode()
	if err != nil {
		return nil, err
	}
	return nodeColumns(node, a.List("columns"))
}

// runDeviceProfile prints the active profile of the device, or with an argument activates the
// profile of that name.
func runDeviceProfile(ctx context.Context, a daemon.Args) ([]string, error) {
	state := pipewire.PwDump()
	device, err := state.FindDeviceByName(a.Flag("device"))
	if err != nil {
		return nil, err
	}

	if len(a.Pos) > 0 {
		profile, err := device.GetProfileIdByName(a.Pos[0])
		if err != nil {
			return nil, err
		}
		return nil, device.SetProfileByName(profile.Name)
	}

	profile, err := device.GetActiveProfile()
	if err != nil {
		return nil, err
	}
	return []string{profile.Description}, nil
}

// nodeColumns returns the requested columns (id, name, description) of a node, one per line.
func nodeColumns(node pipewire.InterfaceNode, columns []string) ([]string, error) {
	var lines []string
	for _, col := range columns {
		switch col {
		case "id":
			lines = append(lines, fmt.Sprint(node.Id))
		case "name":
			name, err := node.GetName()
			if err != nil {
				return nil, err
			}
			lines = append(lines, name)
		case "description":
			desc, err := node.GetDescription()
			if err != nil {
				return nil, err
			}
			lines = append(lines, desc)
		default:
			return nil, fmt.Errorf("unknown column: %s", col)
		}
	}
	return lines, nil
}

// Status describes the health of the PipeWire monitor for `daemon status`.
func Status() []string {
	return []string{fmt.Sprintf("pipewire monitor restarts: %d", pipewire.MonitorRestarts())}
}

// EventSource refreshes watched audio state whenever the PipeWire graph changes.
// It is a daemon.Source.
func EventSource(ctx context.Context, refresh func(prefixes ...string)) {
	pipewire.StartMonitor()
	pipewire.RegisterUpdateCallback(func() {
		refresh("audio.")
	})
}
