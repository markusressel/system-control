// Package supervise keeps long running event sources (a `pw-dump -m` process, a D-Bus
// subscription, ...) alive: one that ends is started again with a backoff.
package supervise

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

// Default restart delays of a Loop.
const (
	DefaultInitialDelay = time.Second
	DefaultMaxDelay     = 30 * time.Second
	DefaultHealthyAfter = time.Minute
)

// Loop runs an event source again and again until its context is done. The zero value (with a
// Name) uses the default delays.
type Loop struct {
	// Name prefixes the log messages, e.g. "bluetooth".
	Name string
	// InitialDelay is the wait before the first restart, it doubles up to MaxDelay with every
	// further one. A run that lasted HealthyAfter resets it.
	InitialDelay, MaxDelay, HealthyAfter time.Duration

	active   atomic.Bool
	restarts atomic.Int64
}

// Active reports whether the source currently delivers events (it called ready and did not end
// since), so that nothing has to poll.
func (l *Loop) Active() bool {
	return l.active.Load()
}

// Status describes the health of the source for `daemon status`.
func (l *Loop) Status() []string {
	state := "down"
	if l.Active() {
		state = "active"
	}
	return []string{fmt.Sprintf("%s: %s, %d restarts", l.Name, state, l.Restarts())}
}

// Restarts is how often the source had to be started again.
func (l *Loop) Restarts() int64 {
	return l.restarts.Load()
}

// Run calls once until ctx is done. once runs the source and returns when it ended; it calls
// ready once it delivers events. A failure is logged once, not on every retry, and its
// recovery as well; recovered (may be nil) is called then, so state that changed in between
// is picked up.
func (l *Loop) Run(ctx context.Context, once func(ctx context.Context, ready func()) error, recovered func()) {
	delay := l.initialDelay()
	failed := false
	ready := func() {
		l.active.Store(true)
		if failed {
			failed = false
			log.Printf("%s: working again", l.Name)
			if recovered != nil {
				recovered()
			}
		}
	}

	for first := true; ctx.Err() == nil; first = false {
		if !first {
			l.restarts.Add(1)
		}
		started := time.Now()
		err := once(ctx, ready)
		l.active.Store(false)
		if ctx.Err() != nil {
			return
		}
		if err != nil && !failed {
			log.Printf("%s: %v (retrying in the background)", l.Name, err)
			failed = true
		}

		if time.Since(started) > l.healthyAfter() {
			delay = l.initialDelay()
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
		delay = min(delay*2, l.maxDelay())
	}
}

func (l *Loop) initialDelay() time.Duration {
	if l.InitialDelay > 0 {
		return l.InitialDelay
	}
	return DefaultInitialDelay
}

func (l *Loop) maxDelay() time.Duration {
	if l.MaxDelay > 0 {
		return l.MaxDelay
	}
	return DefaultMaxDelay
}

func (l *Loop) healthyAfter() time.Duration {
	if l.HealthyAfter > 0 {
		return l.HealthyAfter
	}
	return DefaultHealthyAfter
}
