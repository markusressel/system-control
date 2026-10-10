package daemon

import (
	"errors"
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
