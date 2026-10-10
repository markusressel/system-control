package cmd

import (
	"slices"
	"testing"

	"github.com/markusressel/system-control/internal/daemon"
)

// Every op must be implemented by a cobra command, the daemon derives its command line
// parsing from it.
func TestOpsBindToCommands(t *testing.T) {
	if err := daemon.Default.Bind(RootCmd); err != nil {
		t.Fatal(err)
	}
}

func TestDaemonUnderstandsRealCommandLines(t *testing.T) {
	if err := daemon.Default.Bind(RootCmd); err != nil {
		t.Fatal(err)
	}

	handled := map[string][]string{
		"audio.volume":                    {"audio", "volume"},
		"audio.volume.muted":              {"audio", "volume", "muted", "--watch"},
		"audio.volume.inc":                {"audio", "volume", "inc", "--device", "X"},
		"audio.volume.set":                {"audio", "volume", "set", "50"},
		"audio.volume.toggle-mute":        {"audio", "volume", "toggle-mute"},
		"audio.sink.active":               {"audio", "sink", "active", "-c", "description", "-w"},
		"audio.sink.switch":               {"audio", "sink", "switch", "headphone"},
		"audio.device.profile":            {"audio", "device", "profile", "-d", "X"},
		"bluetooth.connected":             {"bluetooth", "connected", "Headset", "--watch"},
		"fan.mode":                        {"fan", "mode", "--type", "number", "-w"},
		"display.redshift.brightness":     {"display", "redshift", "brightness", "-d", "DP-1", "--format", "%.1f", "--watch"},
		"display.redshift.brightness.inc": {"display", "redshift", "brightness", "inc", "-d", "DP-1"},
		"display.redshift.brightness.dec": {"display", "redshift", "brightness", "dec", "--step", "0.2"},
	}
	for want, argv := range handled {
		op, _, ok := daemon.Default.MatchArgv(argv)
		if !ok || op.Name != want {
			t.Errorf("%v: got op=%v ok=%v, want %s", argv, op, ok, want)
		}
	}

	// everything else is left to cobra, in particular lookalikes of a handled command
	unhandled := [][]string{
		{"display", "redshift", "brightness", "reset"},                // other sub command of an op
		{"display", "redshift", "brightness", "-b", "0.5", "--watch"}, // a set can't be watched
		{"display", "redshift", "gamma", "inc"},                       // not an op
		{"audio", "volume", "set"},                                    // cobra.ExactArgs(1)
		{"audio", "volume", "set", "1", "2"},                          // cobra.ExactArgs(1)
		{"audio", "volume", "--configuration", "x.yaml"},              // root flag
		{"audio", "volume", "--help"},
		{"audio", "device", "profile", "hdmi", "--watch"}, // set can't be watched
		{"audio", "sink", "next", "--watch"},
		{"--configuration", "x.yaml", "audio", "volume"},
		{"bluetooth", "connected"},          // cobra.ExactArgs(1)
		{"bluetooth", "connect", "Headset"}, // not an op
		{"fan", "mode", "set", "1"},         // needs root, not an op
	}
	for _, argv := range unhandled {
		if op, _, ok := daemon.Default.MatchArgv(argv); ok {
			t.Errorf("%v must not be handled by the daemon, matched %s", argv, op.Name)
		}
	}

	if !slices.ContainsFunc(daemon.Default.Ops(), func(op *daemon.Op) bool { return op.Name == "audio.volume.inc" }) {
		t.Error("audio ops are not registered")
	}
}
