package util

import (
	"testing"

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
