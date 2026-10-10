package media

import "testing"

func TestHighestStatus(t *testing.T) {
	for in, want := range map[string]string{
		"Paused\nPlaying\nStopped": "Playing",
		"Stopped\nPaused":          "Paused",
		"Stopped":                  "Stopped",
		"Custom\nOther":            "Custom",
	} {
		if got := HighestStatus(in); got != want {
			t.Errorf("HighestStatus(%q) = %q, want %q", in, got, want)
		}
	}
}
