package redshift

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/markusressel/system-control/internal/daemon"
)

func TestCoalescedStep_SumsStepsOfAJam(t *testing.T) {
	t.Setenv(daemon.EnvRuntimeDir, t.TempDir())

	unlock, err := daemon.AcquireLock(redshiftLockName)
	if err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var applied []float64
	apply := func(total float64) ([]string, error) {
		mu.Lock()
		applied = append(applied, total)
		mu.Unlock()
		return nil, nil
	}

	// ten ticks arrive while the lock is held
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := coalescedStep("test", 0.1, apply); err != nil {
				t.Error(err)
			}
		}()
	}
	time.Sleep(100 * time.Millisecond)
	unlock()
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	sum := 0.0
	for _, v := range applied {
		sum += v
	}
	if math.Abs(sum-1.0) > 1e-9 {
		t.Errorf("applied %v, sum %v, want 1.0", applied, sum)
	}
	if len(applied) > 3 {
		t.Errorf("%d applications for one jam: %v", len(applied), applied)
	}
}

func TestCoalescedStep_KeepsStepsWhenLockStaysBusy(t *testing.T) {
	t.Setenv(daemon.EnvRuntimeDir, t.TempDir())
	previous := stepLockWait
	stepLockWait = 50 * time.Millisecond
	t.Cleanup(func() { stepLockWait = previous })

	unlock, err := daemon.AcquireLock(redshiftLockName)
	if err != nil {
		t.Fatal(err)
	}

	applied := 0.0
	apply := func(total float64) ([]string, error) { applied += total; return nil, nil }

	if _, err := coalescedStep("busy", 0.3, apply); err != nil || applied != 0 {
		t.Fatalf("applied=%v err=%v while the lock was held", applied, err)
	}
	unlock()

	// the next step brings the earlier one along
	if _, err := coalescedStep("busy", 0.1, apply); err != nil || math.Abs(applied-0.4) > 1e-9 {
		t.Errorf("applied=%v err=%v, want 0.4", applied, err)
	}
}
