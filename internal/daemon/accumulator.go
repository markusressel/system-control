package daemon

import "sync"

// Accumulator sums up relative changes (e.g. the ticks of a mouse wheel) of ops that are slower
// than the rate at which they are triggered.
//
// An op adds its change, takes the lock, and applies whatever has accumulated so far in one go.
// A caller that finds nothing left was served by the previous lock holder. No tick is lost, and
// no backlog builds up, unlike dropping the commands that waited too long. A caller that can't
// get the lock withdraws its change, so it is not applied later as a surprise.
type Accumulator struct {
	mu      sync.Mutex
	pending map[string]float64
	// taken counts the Takes per key, so a withdrawal can tell whether its change was applied
	taken map[string]uint64
}

// AccumulatorTicket identifies a change added to an Accumulator, see Withdraw.
type AccumulatorTicket struct {
	key   string
	delta float64
	taken uint64
}

// Add records a change for the given key.
func (a *Accumulator) Add(key string, delta float64) AccumulatorTicket {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.pending == nil {
		a.pending = map[string]float64{}
		a.taken = map[string]uint64{}
	}
	a.pending[key] += delta
	return AccumulatorTicket{key: key, delta: delta, taken: a.taken[key]}
}

// Take returns the change accumulated for the key and resets it.
func (a *Accumulator) Take(key string) float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	delta := a.pending[key]
	delete(a.pending, key)
	if a.taken != nil {
		a.taken[key]++
	}
	return delta
}

// Withdraw removes a change that was not taken yet. It returns false if it was already taken,
// i.e. applied by someone else.
func (a *Accumulator) Withdraw(ticket AccumulatorTicket) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.taken[ticket.key] != ticket.taken {
		return false
	}
	a.pending[ticket.key] -= ticket.delta
	return true
}
