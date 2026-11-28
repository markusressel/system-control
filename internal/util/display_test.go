package util

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestXrandrArgs(t *testing.T) {
	args := XrandrArgs([]DisplayConfig{
		{Name: "DP-1", Mode: "3840x2160", Rate: 144},
		{Name: "DP-2", Primary: true, Position: "3840x0", Auto: true, Mode: "ignored"},
		{Name: "HDMI-1", Off: true, Primary: true, Mode: "ignored"},
	})

	assert.Equal(t, []string{
		"--output", "DP-1", "--mode", "3840x2160", "--rate", "144",
		"--output", "DP-2", "--primary", "--auto", "--pos", "3840x0",
		"--output", "HDMI-1", "--off",
	}, args)
}

func TestGetDisplays_Cache(t *testing.T) {
	calls := 0
	restore := SetRunner(func(ctx context.Context, command string, args ...string) (string, error) {
		calls++
		return "Monitors: 1\n 0: +*DP-1 3840/600x2160/340+0+0  DP-1", nil
	})
	defer restore()
	defer EnableDisplayCache(0)

	// no cache by default: every call asks
	for i := 0; i < 2; i++ {
		displays, err := GetDisplays()
		assert.NoError(t, err)
		assert.Equal(t, []DisplayInfo{{Name: "DP-1"}}, displays)
	}
	assert.Equal(t, 2, calls)

	// with a cache the result is reused, and a caller can not change it
	EnableDisplayCache(time.Minute)
	first, _ := GetDisplays()
	first[0].Name = "changed"
	second, _ := GetDisplays()
	assert.Equal(t, 3, calls)
	assert.Equal(t, "DP-1", second[0].Name)

	// an expired entry is fetched again
	EnableDisplayCache(time.Nanosecond)
	_, _ = GetDisplays()
	time.Sleep(time.Millisecond)
	_, _ = GetDisplays()
	assert.Equal(t, 5, calls)
}
