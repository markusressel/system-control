package backlight

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/markusressel/system-control/internal/daemon"
	"github.com/markusressel/system-control/internal/util"
)

func fakeSysfs(t *testing.T, dir *string, name, brightness, max string) string {
	t.Helper()
	base := t.TempDir()
	led := filepath.Join(base, name)
	if err := os.MkdirAll(led, 0755); err != nil {
		t.Fatal(err)
	}
	for file, value := range map[string]string{util.Brightness: brightness, util.MaxBrightness: max} {
		if err := os.WriteFile(filepath.Join(led, file), []byte(value), 0644); err != nil {
			t.Fatal(err)
		}
	}
	previous := *dir
	*dir = base
	t.Cleanup(func() { *dir = previous })
	return led
}

func read(t *testing.T, led string) int {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(led, util.Brightness))
	if err != nil {
		t.Fatal(err)
	}
	v, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func run(t *testing.T, name string, a daemon.Args) ([]string, error) {
	t.Helper()
	op, ok := daemon.Default.Lookup(name)
	if !ok {
		t.Fatalf("op %s not registered", name)
	}
	if a.Flags == nil {
		a.Flags = map[string]string{}
	}
	return op.Execute(context.Background(), a, time.Now())
}

func TestStepPercent(t *testing.T) {
	for percentage, want := range map[int]int{0: 1, 9: 1, 10: 2, 19: 2, 20: 4, 39: 4, 40: 8, 100: 8} {
		if got := stepPercent(percentage); got != want {
			t.Errorf("stepPercent(%d) = %d, want %d", percentage, got, want)
		}
	}
}

func TestRawStepIsAtLeastOne(t *testing.T) {
	// 1% of a backlight with 7 levels is below one level
	if got := rawStep(3, 7); got != 1 {
		t.Errorf("rawStep(3, 7) = %d, want 1", got)
	}
	// 8% of 1000 levels at 50% brightness
	if got := rawStep(500, 1000); got != 80 {
		t.Errorf("rawStep(500, 1000) = %d, want 80", got)
	}
}

func TestDisplayOps(t *testing.T) {
	t.Setenv(daemon.EnvRuntimeDir, t.TempDir())
	led := fakeSysfs(t, &util.DisplayBacklightPath, "intel_backlight", "500", "1000")

	lines, err := run(t, "display.backlight.brightness", daemon.Args{})
	if err != nil || len(lines) != 1 || lines[0] != "50" {
		t.Fatalf("get: %v, %v", lines, err)
	}

	if _, err := run(t, "display.backlight.brightness.inc", daemon.Args{}); err != nil || read(t, led) != 580 {
		t.Errorf("inc: brightness %d, err %v", read(t, led), err)
	}
	if _, err := run(t, "display.backlight.brightness.dec", daemon.Args{}); err != nil || read(t, led) != 500 {
		t.Errorf("dec: brightness %d, err %v", read(t, led), err)
	}
	if _, err := run(t, "display.backlight.brightness.set", daemon.Args{Pos: []string{"20"}}); err != nil || read(t, led) != 200 {
		t.Errorf("set: brightness %d, err %v", read(t, led), err)
	}
	if _, err := run(t, "display.backlight.brightness.set", daemon.Args{Pos: []string{"x"}}); err == nil {
		t.Error("set with a non-number must fail")
	}
	if _, err := run(t, "display.backlight.brightness.set", daemon.Args{}); err == nil {
		t.Error("set without an argument must fail")
	}
}

func TestKeyboardOps(t *testing.T) {
	t.Setenv(daemon.EnvRuntimeDir, t.TempDir())
	led := fakeSysfs(t, &util.LedsPath, "asus::kbd_backlight", "1", "3")

	lines, err := run(t, "keyboard.brightness", daemon.Args{})
	if err != nil || len(lines) != 1 || lines[0] != "1" {
		t.Fatalf("get: %v, %v", lines, err)
	}

	if _, err := run(t, "keyboard.brightness.inc", daemon.Args{}); err != nil || read(t, led) != 2 {
		t.Errorf("inc: brightness %d, err %v", read(t, led), err)
	}
	// clamped to the maximum
	run(t, "keyboard.brightness.inc", daemon.Args{})
	if _, err := run(t, "keyboard.brightness.inc", daemon.Args{}); err != nil || read(t, led) != 3 {
		t.Errorf("inc at max: brightness %d, err %v", read(t, led), err)
	}
	if _, err := run(t, "keyboard.brightness", daemon.Args{Pos: []string{"0"}}); err != nil || read(t, led) != 0 {
		t.Errorf("set: brightness %d, err %v", read(t, led), err)
	}
}

func TestKeyboardBrightnessWatchOnlyForQueries(t *testing.T) {
	op, _ := daemon.Default.Lookup("keyboard.brightness")
	if !op.Watch(daemon.Args{}) {
		t.Error("a query must be watchable")
	}
	if op.Watch(daemon.Args{Pos: []string{"2"}}) {
		t.Error("a set must not be watchable")
	}
}
