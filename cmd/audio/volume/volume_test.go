package volume

import "testing"

func TestVolumeCommandsRegistered(t *testing.T) {
	expectedCmds := []string{"inc", "dec", "mute", "unmute", "toggle-mute", "save", "restore"}

	for _, expected := range expectedCmds {
		found := false
		for _, cmd := range VolumeCmd.Commands() {
			if cmd.Use == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command '%s' to be registered under VolumeCmd, but it was not", expected)
		}
	}
}
