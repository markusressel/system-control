package media

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/markusressel/system-control/internal/util"
)

// fakePlayerctl answers `playerctl --all-players --format ... status` with the given statuses
// and records every other call.
func fakePlayerctl(t *testing.T, statuses string) *[]string {
	t.Helper()
	var calls []string
	restore := util.SetRunner(func(ctx context.Context, command string, args ...string) (string, error) {
		if command == "playerctl" && slices.Contains(args, "status") {
			return statuses, nil
		}
		calls = append(calls, strings.Join(append([]string{command}, args...), " "))
		return "", nil
	})
	t.Cleanup(restore)
	return &calls
}

// fakeClock makes recordStatus use one minute later on each call.
func fakeClock(t *testing.T) {
	t.Helper()
	previous := now
	at := time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC)
	now = func() time.Time { at = at.Add(time.Minute); return at }
	resetHistory()
	t.Cleanup(func() { now = previous; resetHistory() })
}

func TestRecordStatus_PlayingUntilItStops(t *testing.T) {
	fakeClock(t)
	// the voice message played after the video started, but the video played on longer
	recordStatus("firefox", "Playing")
	recordStatus("beeper", "Playing")
	recordStatus("beeper", "Paused")
	recordStatus("firefox", "Paused")
	if got := LastPlayed([]string{"beeper", "firefox"}); got != "firefox" {
		t.Errorf("LastPlayed = %q, want firefox", got)
	}

	// a status of a player that was not playing does not count
	recordStatus("beeper", "Stopped")
	if got := LastPlayed([]string{"beeper", "firefox"}); got != "firefox" {
		t.Errorf("LastPlayed = %q after a stop, want firefox", got)
	}
	if got := LastPlayed([]string{"spotify"}); got != "" {
		t.Errorf("LastPlayed of an unknown player = %q, want none", got)
	}
}

func TestParseFollowLine(t *testing.T) {
	for line, want := range map[string][2]string{
		"firefox.instance_1_42\tPlaying\n": {"firefox.instance_1_42", "Playing"},
		"spotify\tPaused":                  {"spotify", "Paused"},
		"":                                 {"", ""},
		"No players found":                 {"", ""},
	} {
		player, status, _ := parseFollowLine(line)
		if player != want[0] || status != want[1] {
			t.Errorf("parseFollowLine(%q) = %q, %q, want %q, %q", line, player, status, want[0], want[1])
		}
	}
}

func TestPlayAndToggle(t *testing.T) {
	tests := []struct {
		name     string
		run      func(string) error
		pattern  string
		statuses string
		want     []string
	}{
		{"play resumes the last player", Play, "", "beeper\tPaused\nfirefox\tPaused", []string{"playerctl --player firefox play"}},
		{"play while one plays", Play, "", "beeper\tPaused\nfirefox\tPlaying", nil},
		{"play without history", Play, "", "spotify\tPaused", []string{"playerctl play"}},
		{"toggle resumes the last player", TogglePlayPause, "", "beeper\tPaused\nfirefox\tPaused", []string{"playerctl --player firefox play"}},
		{"toggle pauses all", TogglePlayPause, "", "beeper\tPaused\nfirefox\tPlaying", []string{"playerctl --all-players pause"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClock(t)
			recordStatus("beeper", "Playing")
			recordStatus("firefox", "Playing")
			recordStatus("beeper", "Paused")
			recordStatus("firefox", "Paused")
			calls := fakePlayerctl(t, tt.statuses)

			if err := tt.run(tt.pattern); err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(*calls, tt.want) {
				t.Errorf("calls = %q, want %q", *calls, tt.want)
			}
		})
	}
}
