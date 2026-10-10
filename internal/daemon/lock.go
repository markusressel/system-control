package daemon

import (
	"context"
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
	dir := RuntimeDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	fileLock := flock.New(filepath.Join(dir, name+".lock"))
	ctx, cancel := context.WithTimeout(context.Background(), LockTimeout)
	defer cancel()

	locked, err := fileLock.TryLockContext(ctx, lockRetryDelay)
	if err != nil {
		return nil, fmt.Errorf("could not acquire lock %s: %w", name, err)
	}
	if !locked {
		return nil, fmt.Errorf("could not acquire lock %s", name)
	}
	return func() { _ = fileLock.Unlock() }, nil
}

// WithLock runs fn while holding the cross-process lock `name`.
//
// If more than MaxQueueWait passed between queuedAt and acquiring the lock, fn is not run
// and dropped is true.
func WithLock(name string, queuedAt time.Time, fn func() error) (dropped bool, err error) {
	unlock, err := AcquireLock(name)
	if err != nil {
		return false, err
	}
	defer unlock()

	if time.Since(queuedAt) > MaxQueueWait {
		log.Printf("Dropping queued command %s: waited %v for the lock", name, time.Since(queuedAt))
		return true, nil
	}

	return false, fn()
}
