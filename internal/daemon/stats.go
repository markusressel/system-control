package daemon

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// opStats counts what happened to the executions of one op, shown by `daemon status`.
// A dropped or failed command is otherwise invisible: the client only sees an exit code.
type opStats struct {
	mu        sync.Mutex
	calls     int
	failed    int
	dropped   int
	unhandled int
	total     time.Duration
	slowest   time.Duration
	lastErr   string
	lastErrAt time.Time
}

func (s *opStats) record(took time.Duration, dropped bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch {
	case errors.Is(err, ErrUnhandled):
		s.unhandled++
		return
	case dropped:
		s.dropped++
	case err != nil:
		s.failed++
		s.lastErr, s.lastErrAt = err.Error(), time.Now()
	}
	s.calls++
	s.total += took
	s.slowest = max(s.slowest, took)
}

// line returns a summary, or "" if the op was never executed.
func (s *opStats) line(name string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.calls == 0 && s.unhandled == 0 {
		return ""
	}
	line := fmt.Sprintf("%s: %d calls", name, s.calls)
	if s.calls > 0 {
		line += fmt.Sprintf(", avg %s, max %s",
			(s.total / time.Duration(s.calls)).Round(time.Millisecond), s.slowest.Round(time.Millisecond))
	}
	if s.dropped > 0 {
		line += fmt.Sprintf(", %d dropped", s.dropped)
	}
	if s.unhandled > 0 {
		line += fmt.Sprintf(", %d unhandled", s.unhandled)
	}
	if s.failed > 0 {
		line += fmt.Sprintf(", %d failed (last %s ago: %s)",
			s.failed, time.Since(s.lastErrAt).Round(time.Second), s.lastErr)
	}
	return line
}
