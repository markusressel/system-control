package daemon

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"time"
)

// StatusInfo describes a running daemon. It is what `daemon status` renders, and the payload of
// the `STATUS json` request.
type StatusInfo struct {
	PID       int       `json:"pid"`
	Socket    string    `json:"socket"`
	StartedAt time.Time `json:"startedAt"`
	Ops       int       `json:"ops"`

	Watches    []WatchStatus `json:"watches"`
	Executions []ExecStatus  `json:"executions"`
	// Details are lines of event sources, see Engine.AddStatus.
	Details []string `json:"details,omitempty"`
}

// WatchStatus is one watched state, shared by all of its subscribers.
type WatchStatus struct {
	Op          string   `json:"op"`
	Args        string   `json:"args"`
	Subscribers int      `json:"subscribers"`
	Value       []string `json:"value,omitempty"`
	Error       string   `json:"error,omitempty"`
}

// ExecStatus summarizes the executions of one op.
type ExecStatus struct {
	Op          string        `json:"op"`
	Calls       int           `json:"calls"`
	Failed      int           `json:"failed"`
	Dropped     int           `json:"dropped"`
	Unhandled   int           `json:"unhandled"`
	Avg         time.Duration `json:"avg"`
	Max         time.Duration `json:"max"`
	LastError   string        `json:"lastError,omitempty"`
	LastErrorAt time.Time     `json:"lastErrorAt,omitzero"`
}

func (s *opStats) snapshot(name string) (ExecStatus, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.calls == 0 && s.unhandled == 0 {
		return ExecStatus{}, false
	}
	st := ExecStatus{
		Op: name, Calls: s.calls, Failed: s.failed, Dropped: s.dropped, Unhandled: s.unhandled,
		Max: s.slowest, LastError: s.lastErr, LastErrorAt: s.lastErrAt,
	}
	if s.calls > 0 {
		st.Avg = s.total / time.Duration(s.calls)
	}
	return st, true
}

// watches lists the watched entries.
func (s *store) watches() []WatchStatus {
	s.mu.Lock()
	entries := make([]*entry, 0, len(s.entries))
	for _, e := range s.entries {
		entries = append(entries, e)
	}
	s.mu.Unlock()

	result := make([]WatchStatus, 0, len(entries))
	for _, e := range entries {
		e.mu.Lock()
		w := WatchStatus{
			Op:          e.op.Name,
			Args:        e.describeArgs(),
			Subscribers: len(e.subs),
			Value:       slices.Clone(e.last),
			Error:       e.lastErr,
		}
		e.mu.Unlock()
		result = append(result, w)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Op != result[j].Op {
			return result[i].Op < result[j].Op
		}
		return result[i].Args < result[j].Args
	})
	return result
}

// describeArgs lists the arguments that differ from the defaults of the op.
func (e *entry) describeArgs() string {
	defaults := e.op.defaultArgs().Flags

	var parts []string
	for name, value := range e.args.Flags {
		if value != defaults[name] {
			parts = append(parts, fmt.Sprintf("%s=%s", name, value))
		}
	}
	sort.Strings(parts)
	parts = append(parts, e.args.Pos...)
	return strings.Join(parts, " ")
}

// statusInfo collects the state of the engine.
func (e *Engine) statusInfo() StatusInfo {
	info := StatusInfo{
		PID:        os.Getpid(),
		Socket:     SocketPath(),
		StartedAt:  e.startedAt,
		Ops:        len(e.registry.Ops()),
		Watches:    e.store.watches(),
		Executions: []ExecStatus{},
	}
	for _, op := range e.registry.Ops() {
		if op.stats == nil {
			continue
		}
		if st, ok := op.stats.snapshot(op.Name); ok {
			info.Executions = append(info.Executions, st)
		}
	}
	for _, f := range e.statuses {
		info.Details = append(info.Details, f()...)
	}
	return info
}
