package daemon

import "sync"

// Accumulator sums up relative changes (e.g. the ticks of a mouse wheel) of ops that are slower
// than the rate at which they are triggered.
//
// An op adds its change, takes the lock, and applies whatever has accumulated so far in one go.
// A caller that finds nothing left was served by the previous lock holder. No tick is lost, and
// no backlog builds up, unlike dropping the commands that waited too long.
type Accumulator struct {
	mu      sync.Mutex
	pending map[string]float64
}

// Add records a change for the given key.
func (a *Accumulator) Add(key string, delta float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.pending == nil {
		a.pending = map[string]float64{}
	}
	a.pending[key] += delta
}

// Take returns the change accumulated for the key and resets it.
func (a *Accumulator) Take(key string) float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	delta := a.pending[key]
	delete(a.pending, key)
	return delta
}
