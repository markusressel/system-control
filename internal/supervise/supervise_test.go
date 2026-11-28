package supervise

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestLoop_RestartsAndRecovers(t *testing.T) {
	l := &Loop{Name: "test", InitialDelay: time.Millisecond}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var runs, recoveries atomic.Int64
	reachedThird := make(chan struct{})
	done := make(chan struct{})
	go func() {
		l.Run(ctx, func(ctx context.Context, ready func()) error {
			switch runs.Add(1) {
			case 1:
				ready()
				return errors.New("lost") // fails after it worked
			case 2:
				return errors.New("still lost") // fails before it works
			default:
				ready()
				close(reachedThird)
				<-ctx.Done()
				return nil
			}
		}, func() { recoveries.Add(1) })
		close(done)
	}()

	select {
	case <-reachedThird:
	case <-time.After(5 * time.Second):
		t.Fatal("no third run")
	}
	if !l.Active() || l.Restarts() != 2 || recoveries.Load() != 1 {
		t.Errorf("active=%v restarts=%d recoveries=%d, want true, 2, 1", l.Active(), l.Restarts(), recoveries.Load())
	}

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after the context was canceled")
	}
	if l.Active() {
		t.Error("Active() = true after Run returned")
	}
}

func TestLoop_ReturnsWhileWaiting(t *testing.T) {
	l := &Loop{Name: "test", InitialDelay: time.Hour}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		l.Run(ctx, func(context.Context, func()) error { return errors.New("down") }, nil)
		close(done)
	}()
	time.Sleep(10 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return during its backoff")
	}
}
