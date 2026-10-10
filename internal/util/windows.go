package util

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Window is a top level window as listed by wmctrl.
type Window struct {
	ID      string
	Desktop string
	Title   string
}

var windowLineRegex = regexp.MustCompile(`^(0x[0-9a-fA-F]+)\s+(-?\d+)\s+(\S+)\s*(.*)$`)

// ParseWindowList parses the output of `wmctrl -l`.
func ParseWindowList(output string) []Window {
	windows := []Window{}
	for _, line := range strings.Split(output, "\n") {
		match := windowLineRegex.FindStringSubmatch(strings.TrimSpace(line))
		if match == nil {
			continue
		}
		windows = append(windows, Window{ID: match[1], Desktop: match[2], Title: match[4]})
	}
	return windows
}

// ListOpenWindows returns the currently open windows. Sticky windows (desktop -1,
// e.g. panels and desktop backgrounds) are not included.
func ListOpenWindows() ([]Window, error) {
	output, err := ExecCommand("wmctrl", "-l")
	if err != nil {
		return nil, err
	}

	windows := []Window{}
	for _, window := range ParseWindowList(output) {
		if window.Desktop != "-1" {
			windows = append(windows, window)
		}
	}
	return windows, nil
}

// CloseAllWindows asks every open window to close and waits until all are gone.
// It returns the windows that are still open after the timeout.
func CloseAllWindows(timeout time.Duration) (remaining []Window, err error) {
	windows, err := ListOpenWindows()
	if err != nil {
		return nil, err
	}

	for _, window := range windows {
		// a window may have closed in the meantime, so a failure here is not fatal
		_, _ = ExecCommand("wmctrl", "-i", "-c", window.ID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		remaining, err = ListOpenWindows()
		if err != nil {
			return nil, err
		}
		if len(remaining) == 0 {
			return nil, nil
		}

		select {
		case <-ctx.Done():
			return remaining, nil
		case <-time.After(500 * time.Millisecond):
		}
	}
}

// DescribeWindows formats windows for an error message.
func DescribeWindows(windows []Window) string {
	titles := make([]string, 0, len(windows))
	for _, window := range windows {
		titles = append(titles, fmt.Sprintf("%q", window.Title))
	}
	return strings.Join(titles, ", ")
}
