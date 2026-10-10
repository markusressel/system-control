package daemon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"
)

// entry is the observed state of one watched invocation. Its subscribers share a single
// query; the entry lives exactly as long as it has subscribers.
type entry struct {
	key  string
	op   *Op
	args Args

	// mu guards the fields below
	mu      sync.Mutex
	last    []string
	hasLast bool
	subs    map[chan []string]struct{}

	running bool
	again   bool
	lastErr string

	stopPoll context.CancelFunc
}

// store keeps all watched entries of an engine.
type store struct {
	mu      sync.Mutex
	entries map[string]*entry
}

func newStore() *store {
	return &store{entries: map[string]*entry{}}
}

// subscribe returns a channel that receives the current value immediately and then every change.
// Slow consumers only ever see the latest value. cancel must be called to release the subscription.
//
// If the first query fails (the state may just not be there yet, e.g. a device that is about
// to show up), the subscription stays valid and the error is returned as pending. The values
// follow once a query succeeds. Only ErrUnhandled ends the subscription right away.
func (s *store) subscribe(ctx context.Context, op *Op, a Args) (updates <-chan []string, cancel func(), pending error) {
	key := a.key(op.Name)

	s.mu.Lock()
	e, exists := s.entries[key]
	if !exists {
		e = &entry{key: key, op: op, args: a, subs: map[chan []string]struct{}{}}
		s.entries[key] = e
	}
	s.mu.Unlock()

	ch := make(chan []string, 1)

	err := e.refresh(ctx)
	if errors.Is(err, ErrUnhandled) && !e.hasValue() {
		s.release(e, nil)
		return nil, nil, err
	}
	if err != nil && !e.hasValue() {
		pending = err
	}

	e.mu.Lock()
	e.subs[ch] = struct{}{}
	if e.hasLast {
		ch <- slices.Clone(e.last)
	}
	e.mu.Unlock()

	if !exists {
		s.startPolling(e)
	}

	return ch, func() { s.release(e, ch) }, pending
}

func (s *store) release(e *entry, ch chan []string) {
	e.mu.Lock()
	if ch != nil {
		delete(e.subs, ch)
	}
	empty := len(e.subs) == 0
	e.mu.Unlock()

	if !empty {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// re-check under the store lock, a new subscriber may have joined in between
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.subs) == 0 && s.entries[e.key] == e {
		delete(s.entries, e.key)
		if e.stopPoll != nil {
			e.stopPoll()
		}
	}
}

func (s *store) startPolling(e *entry) {
	if e.op.Poll == nil {
		return
	}
	interval := e.op.Poll()
	if interval <= 0 {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	e.mu.Lock()
	e.stopPoll = cancel
	e.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				e.trigger()
			}
		}
	}()
}

// refreshMatching re-queries all entries whose op name starts with one of the prefixes.
func (s *store) refreshMatching(prefixes ...string) {
	s.mu.Lock()
	var matching []*entry
	for _, e := range s.entries {
		for _, p := range prefixes {
			if strings.HasPrefix(e.op.Name, p) {
				matching = append(matching, e)
				break
			}
		}
	}
	s.mu.Unlock()

	for _, e := range matching {
		e.trigger()
	}
}

func (e *entry) hasValue() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.hasLast
}

// trigger asynchronously re-queries the entry. Triggers arriving while a query runs are
// coalesced into a single follow-up query.
func (e *entry) trigger() {
	e.mu.Lock()
	if e.running {
		e.again = true
		e.mu.Unlock()
		return
	}
	e.running = true
	e.mu.Unlock()

	go func() {
		for {
			e.logRefresh(e.refresh(context.Background()))

			e.mu.Lock()
			if !e.again {
				e.running = false
				e.mu.Unlock()
				return
			}
			e.again = false
			e.mu.Unlock()
		}
	}()
}

// logRefresh logs a failing watch once, not on every event.
func (e *entry) logRefresh(err error) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}

	e.mu.Lock()
	changed := msg != e.lastErr
	e.lastErr = msg
	e.mu.Unlock()

	if changed && err != nil {
		log.Printf("watch %s: %v", e.op.Name, err)
	}
}

// refresh queries the current value and notifies subscribers if it changed.
func (e *entry) refresh(ctx context.Context) (err error) {
	// Run is arbitrary op code working on external state, it must not take the daemon down.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	lines, err := e.op.Run(ctx, e.args)
	if err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.hasLast && slices.Equal(e.last, lines) {
		return nil
	}
	e.last = lines
	e.hasLast = true

	for ch := range e.subs {
		// latest value wins
		select {
		case <-ch:
		default:
		}
		ch <- slices.Clone(lines)
	}
	return nil
}

// describe lists the watched entries with their number of subscribers.
func (s *store) describe() []string {
	s.mu.Lock()
	entries := make([]*entry, 0, len(s.entries))
	for _, e := range s.entries {
		entries = append(entries, e)
	}
	s.mu.Unlock()

	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		e.mu.Lock()
		subscribers := len(e.subs)
		e.mu.Unlock()
		lines = append(lines, fmt.Sprintf("%s (%d subscribers)", strings.ReplaceAll(e.key, "\x00", " "), subscribers))
	}
	slices.Sort(lines)
	return lines
}
