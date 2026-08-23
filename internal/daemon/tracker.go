package daemon

import (
	"context"
	"time"
)

// PollingTracker encapsulates the boilerplate for a polling-based daemon module.
type PollingTracker struct {
	Interval   time.Duration
	Path       string
	Updates    chan<- StateUpdate
	FetchState func() (string, error)

	wakeup chan struct{}
}

func NewPollingTracker(interval time.Duration, path string, updates chan<- StateUpdate, fetchState func() (string, error)) *PollingTracker {
	return &PollingTracker{
		Interval:   interval,
		Path:       path,
		Updates:    updates,
		FetchState: fetchState,
		wakeup:     make(chan struct{}, 10),
	}
}

func (t *PollingTracker) Start(ctx context.Context) {
	ticker := time.NewTicker(t.Interval)
	defer ticker.Stop()

	var lastVal string

	updateFunc := func() {
		val, err := t.FetchState()
		if err == nil && val != lastVal {
			lastVal = val
			t.Updates <- StateUpdate{Path: t.Path, Value: val}
		}
	}

	updateFunc()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.wakeup:
			updateFunc()
		case <-ticker.C:
			updateFunc()
		}
	}
}

func (t *PollingTracker) Wakeup() {
	select {
	case t.wakeup <- struct{}{}:
	default:
	}
}
