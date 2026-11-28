package daemon

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func receive(t *testing.T, updates <-chan update) update {
	t.Helper()
	select {
	case u := <-updates:
		return u
	case <-time.After(5 * time.Second):
		t.Fatal("no update")
		return update{}
	}
}

// A subscriber joining while the last one leaves must keep the entry alive: it would otherwise
// join an entry that no refresh ever reaches again.
func TestStore_SubscriberJoiningWhileTheLastLeaves(t *testing.T) {
	var value atomic.Int64
	block := make(chan struct{})
	var blockNext atomic.Bool
	op := &Op{Name: "test.value", Run: func(ctx context.Context, a Args) ([]string, error) {
		if blockNext.CompareAndSwap(true, false) {
			<-block
		}
		return []string{string(rune('0' + value.Load()))}, nil
	}}
	s := newStore()

	_, cancelFirst, _ := s.subscribe(context.Background(), op, Args{})

	blockNext.Store(true)
	joined := make(chan (<-chan update))
	go func() {
		updates, _, _ := s.subscribe(context.Background(), op, Args{})
		joined <- updates
	}()
	waitUntil(t, func() bool { return !blockNext.Load() }) // the second subscriber queries

	cancelFirst()
	close(block)
	updates := <-joined
	if u := receive(t, updates); u.lines[0] != "0" {
		t.Fatalf("first value = %v", u.lines)
	}

	value.Store(1)
	s.refreshMatching("test.")
	if u := receive(t, updates); u.lines[0] != "1" {
		t.Errorf("after a change = %v, want 1", u.lines)
	}
}

// A query that started before a change must not publish its result after the query that saw
// the change.
func TestStore_QueriesOfAnEntryDoNotOverlap(t *testing.T) {
	var running, maxRunning atomic.Int64
	op := &Op{Name: "test.value", Run: func(ctx context.Context, a Args) ([]string, error) {
		n := running.Add(1)
		defer running.Add(-1)
		if n > maxRunning.Load() {
			maxRunning.Store(n)
		}
		time.Sleep(20 * time.Millisecond)
		return []string{"x"}, nil
	}}
	s := newStore()

	_, cancel, _ := s.subscribe(context.Background(), op, Args{})
	defer cancel()
	done := make(chan struct{})
	go func() {
		_, cancelSecond, _ := s.subscribe(context.Background(), op, Args{})
		cancelSecond()
		close(done)
	}()
	s.refreshMatching("test.")
	<-done
	waitUntil(t, func() bool { return running.Load() == 0 })

	if maxRunning.Load() > 1 {
		t.Errorf("%d queries of one entry ran at the same time", maxRunning.Load())
	}
}

func waitUntil(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("timed out")
		}
		time.Sleep(2 * time.Millisecond)
	}
}
