package daemon

import (
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

var ansi = regexp.MustCompile("\x1b\\[[0-9;?]*[a-zA-Z]")

func sampleStatus(now time.Time) StatusInfo {
	return StatusInfo{
		PID: 42, Socket: "/run/user/1000/system-control/ipc.sock", StartedAt: now.Add(-90 * time.Second), Ops: 20,
		Watches: []WatchStatus{
			{Op: "audio.volume", Args: "", Subscribers: 2, Value: []string{"45"}},
			{Op: "display.redshift.brightness", Args: "display=DP-1 format=%.1f", Subscribers: 1, Value: []string{"0.8"}},
			{Op: "bluetooth.connected", Args: "Headset", Subscribers: 1, Error: "no adapter"},
		},
		Executions: []ExecStatus{
			{Op: "audio.volume.inc", Calls: 40, Dropped: 12, Avg: 800 * time.Microsecond, Max: 3 * time.Millisecond},
			{Op: "fan.mode", Calls: 3, Failed: 1, Avg: time.Millisecond, Max: 2 * time.Millisecond,
				LastError: "no fan policy files found", LastErrorAt: now.Add(-5 * time.Second)},
		},
		Details: []string{"pipewire monitor restarts: 0"},
	}
}

func TestRenderStatus_PlainHasNoEscapes(t *testing.T) {
	now := time.Now()
	out := RenderStatus(sampleStatus(now), RenderOptions{Now: now})

	if ansi.MatchString(out) {
		t.Errorf("plain output contains escape sequences:\n%q", out)
	}
	for _, want := range []string{
		"running", "pid 42", "up 1m30s", "20 ops",
		"Watchers (3)", "audio.volume", "display=DP-1 format=%.1f", "error: no adapter",
		"Executions (2)", "audio.volume.inc", "800µs", "last error 5s ago: no fan policy files found",
		"Sources (1)", "pipewire monitor restarts: 0",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestRenderStatus_ColorKeepsAlignment(t *testing.T) {
	now := time.Now()
	plain := RenderStatus(sampleStatus(now), RenderOptions{Now: now})
	colored := RenderStatus(sampleStatus(now), RenderOptions{Now: now, Color: true})

	if !ansi.MatchString(colored) {
		t.Fatal("colored output has no escape sequences")
	}
	if got := ansi.ReplaceAllString(colored, ""); got != plain {
		t.Errorf("colors changed the layout:\n--- plain\n%s\n--- colored without escapes\n%s", plain, got)
	}
}

func TestRenderStatus_FitsWidth(t *testing.T) {
	now := time.Now()
	info := sampleStatus(now)
	info.Watches[0].Value = []string{strings.Repeat("x", 300)}

	for _, line := range strings.Split(RenderStatus(info, RenderOptions{Now: now, Width: 80}), "\n") {
		if n := len([]rune(line)); n > 80 && !strings.HasPrefix(line, "socket") {
			t.Errorf("line has %d columns: %q", n, line)
		}
	}
}

func TestFetchStatus(t *testing.T) {
	var counter atomic.Int64
	eng := startTestEngine(t, &counter)
	_ = eng

	info, legacy, err := FetchStatus()
	if err != nil || info == nil || legacy != nil {
		t.Fatalf("info=%v legacy=%v err=%v", info, legacy, err)
	}
	if info.Ops == 0 || info.PID == 0 || info.Socket == "" {
		t.Errorf("incomplete status: %+v", info)
	}

	// the plain request still gets text
	lines, err := Status()
	if err != nil || len(lines) == 0 || strings.HasPrefix(lines[0], "{") {
		t.Errorf("lines=%v err=%v", lines, err)
	}
}
