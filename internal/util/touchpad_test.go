package util

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const synclientOutput = `Parameter settings:
    LeftEdge                = 1310
    VertScrollDelta         = -111
    TouchpadOff             = 0
`

func TestParseSynclientList(t *testing.T) {
	settings := ParseSynclientList(synclientOutput)
	assert.Equal(t, "-111", settings["VertScrollDelta"])
	assert.Equal(t, "0", settings["TouchpadOff"])
	assert.NotContains(t, settings, "Parameter settings:")
}

func TestParseSynclientListWithoutDriver(t *testing.T) {
	settings := ParseSynclientList("Couldn't find synaptics properties. No synaptics driver loaded?")
	assert.Empty(t, settings)
}

// fakeCommands answers commands by their joined command line.
func fakeCommands(t *testing.T, answers map[string]string) {
	restore := SetRunner(func(ctx context.Context, command string, args ...string) (string, error) {
		line := strings.Join(append([]string{command}, args...), " ")
		if answer, ok := answers[line]; ok {
			if strings.HasPrefix(answer, "ERR:") {
				return "", errors.New(strings.TrimPrefix(answer, "ERR:"))
			}
			return answer, nil
		}
		return "", errors.New("unexpected command: " + line)
	})
	t.Cleanup(restore)
}

func TestIsTouchpadEnabledWithoutSynaptics(t *testing.T) {
	fakeCommands(t, map[string]string{
		"synclient -l":                           "ERR:no synaptics driver",
		"xinput list --name-only":                "Virtual core pointer\nSynPS/2 Synaptics TouchPad",
		"xinput list SynPS/2 Synaptics TouchPad": "⎜   ↳ SynPS/2 Synaptics TouchPad",
	})

	enabled, err := IsTouchpadEnabled()
	assert.NoError(t, err)
	assert.True(t, enabled)
}

func TestIsTouchpadEnabledNothingAvailable(t *testing.T) {
	fakeCommands(t, map[string]string{
		"synclient -l":            "ERR:no synaptics driver",
		"xinput list --name-only": "Virtual core pointer",
	})

	_, err := IsTouchpadEnabled()
	assert.Error(t, err)
}

func TestSetTouchpadEnabledSucceedsWithOneBackend(t *testing.T) {
	fakeCommands(t, map[string]string{
		"synclient TouchpadOff=1":                   "ERR:no synaptics driver",
		"xinput list --name-only":                   "SynPS/2 Synaptics TouchPad",
		"xinput disable SynPS/2 Synaptics TouchPad": "",
	})

	assert.NoError(t, SetTouchpadEnabled(false))
}

func TestSetTouchpadEnabledFailsWithoutBackend(t *testing.T) {
	fakeCommands(t, map[string]string{
		"synclient TouchpadOff=0": "ERR:no synaptics driver",
		"xinput list --name-only": "Virtual core pointer",
	})

	assert.Error(t, SetTouchpadEnabled(true))
}
