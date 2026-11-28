package redshift

import (
	"errors"
	"time"

	"github.com/markusressel/system-control/internal/daemon"
)

// stepAccumulator sums the steps of inc/dec ops, see coalescedStep.
var stepAccumulator daemon.Accumulator

// stepLockWait bounds how long a step waits for the redshift lock. A step that gives up is
// dropped (ErrStepDropped), unless a lock holder applied it in the meantime.
var stepLockWait = 2 * time.Second

// ErrStepDropped is returned for a step that could not get the redshift lock in time.
var ErrStepDropped = errors.New("redshift is busy, step dropped")

// coalescedStep applies a relative change that may arrive faster than it can be applied
// (scrolling the mouse wheel on a bar module, applying takes about 150ms).
//
// The change is accumulated under `key` first. Whoever gets the lock applies everything that
// accumulated, so a burst of N ticks results in N steps, in few applications, instead of
// the waiting ones being dropped.
func coalescedStep(key string, delta float64, apply func(total float64) ([]string, error)) ([]string, error) {
	ticket := stepAccumulator.Add(key, delta)

	unlock, ok, err := daemon.AcquireLockWithin(LockName, stepLockWait)
	if err != nil || !ok {
		if !stepAccumulator.Withdraw(ticket) {
			// a lock holder applied it while we waited
			return nil, err
		}
		if err == nil {
			err = ErrStepDropped
		}
		return nil, err
	}
	defer unlock()

	total := stepAccumulator.Take(key)
	if total == 0 {
		// applied by the holder we waited for
		return nil, nil
	}
	return apply(total)
}
