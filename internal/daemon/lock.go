package daemon

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

// MaxQueueWait is how long a mutating command may wait for its lock before it is
// considered stale (e.g. a backlog of scroll events) and dropped.
const MaxQueueWait = 200 * time.Millisecond

// WithLock runs fn while holding the cross-process lock `name`. The same lock is used by
// the daemon and by in-process CLI runs, so mutations are serialized no matter who executes them.
//
// If more than MaxQueueWait passed between queuedAt and acquiring the lock, fn is not run
// and dropped is true.
func WithLock(name string, queuedAt time.Time, fn func() error) (dropped bool, err error) {
	dir := RuntimeDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return false, err
	}

	fileLock := flock.New(filepath.Join(dir, name+".lock"))
	if err := fileLock.Lock(); err != nil {
		return false, err
	}
	defer fileLock.Unlock()

	if time.Since(queuedAt) > MaxQueueWait {
		log.Printf("Dropping queued command %s: waited %v for the lock", name, time.Since(queuedAt))
		return true, nil
	}

	return false, fn()
}
