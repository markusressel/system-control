package daemon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

// MaxQueueWait is how long a mutating command may wait for its lock before it is
// considered stale (e.g. a backlog of scroll events) and dropped.
const MaxQueueWait = 200 * time.Millisecond

// LockTimeout bounds how long a command waits for a lock held by another process. A holder that
// is stuck (e.g. a hung external tool) must not block every later command forever.
const LockTimeout = 30 * time.Second

const lockRetryDelay = 5 * time.Millisecond

// AcquireLock takes the cross-process lock `name` and returns the function that releases it.
// The same lock is used by the daemon and by in-process CLI runs, so mutations are
// serialized no matter who executes them.
func AcquireLock(name string) (unlock func(), err error) {
	unlock, acquired, err := acquireLock(name, LockTimeout)
	if err == nil && !acquired {
		return nil, fmt.Errorf("could not acquire lock %s within %v", name, LockTimeout)
	}
	return unlock, err
}

// acquireLock waits at most `wait` for the lock. A lock that is still held afterwards is
// reported as acquired=false, not as an error.
func acquireLock(name string, wait time.Duration) (unlock func(), acquired bool, err error) {
	dir := RuntimeDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, false, err
	}

	fileLock := flock.New(filepath.Join(dir, name+".lock"))
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	locked, err := fileLock.TryLockContext(ctx, lockRetryDelay)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return nil, false, fmt.Errorf("could not acquire lock %s: %w", name, err)
	}
	if !locked {
		return nil, false, nil
	}
	return func() { _ = fileLock.Unlock() }, true, nil
}

// WithLock runs fn while holding the cross-process lock `name`.
//
// A command waits for the lock only until MaxQueueWait has passed since queuedAt. After that
// fn is not run and dropped is true, so a stuck lock holder cannot pile up waiting commands.
func WithLock(name string, queuedAt time.Time, fn func() error) (dropped bool, err error) {
	var unlock func()
	var acquired bool
	if wait := MaxQueueWait - time.Since(queuedAt); wait > 0 {
		unlock, acquired, err = acquireLock(name, wait)
		if err != nil {
			return false, err
		}
	}
	if !acquired {
		log.Printf("Dropping queued command %s: waited %v for the lock", name, time.Since(queuedAt))
		return true, nil
	}
	defer unlock()

	return false, fn()
}

// AcquireLockWithin is like AcquireLock, but gives up after `wait` and reports ok=false instead
// of an error. Ops that coalesce their work (see Accumulator) use it: giving up is harmless
// there, the work stays pending for whoever holds the lock next.
func AcquireLockWithin(name string, wait time.Duration) (unlock func(), ok bool, err error) {
	return acquireLock(name, wait)
}
