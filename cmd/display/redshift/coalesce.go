package redshift

import (
	"time"

	"github.com/markusressel/system-control/internal/daemon"
)

// stepAccumulator sums the steps of inc/dec ops, see coalescedStep.
var stepAccumulator daemon.Accumulator

// stepLockWait bounds how long a step waits for the redshift lock. A step that gives up is not
// lost, it stays accumulated for the next one.
var stepLockWait = 2 * time.Second

// coalescedStep applies a relative change that may arrive faster than it can be applied
// (scrolling the mouse wheel on a bar module, applying takes about 150ms).
//
// The change is accumulated under `key` first. Whoever gets the lock applies everything that
// accumulated, so a burst of N ticks results in N steps, in few applications, instead of
// the waiting ones being dropped.
func coalescedStep(key string, delta float64, apply func(total float64) ([]string, error)) ([]string, error) {
	stepAccumulator.Add(key, delta)

	unlock, ok, err := daemon.AcquireLockWithin(redshiftLockName, stepLockWait)
	if err != nil || !ok {
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
