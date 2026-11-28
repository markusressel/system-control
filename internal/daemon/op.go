package daemon

import (
	"context"
	"strings"
	"time"
)

// Op is a single CLI operation that can run in the daemon or in-process, with one implementation.
//
// An op is bound to the cobra command at its CLI path, which declares flags, arguments and help.
type Op struct {
	// Name is the dotted CLI path, e.g. "audio.volume.inc" for `system-control audio volume inc`.
	Name string

	// Lock names a cross-process lock held while Run executes (mutating ops).
	Lock string

	// LockFor decides per invocation, for ops that both query and mutate depending on their
	// arguments. It returns the lock to hold, or "" for none, and takes precedence over Lock.
	LockFor func(a Args) string

	// Requires lists environment variables the op needs (e.g. "DISPLAY" for X11 tools). A daemon
	// that lacks one takes it from the user service manager, and if that has none either, the
	// client runs the command itself.
	Requires []string

	// Run executes the op and returns the lines to print.
	Run func(ctx context.Context, a Args) ([]string, error)

	// Watch reports whether this invocation can be watched (Run must be a side effect free query).
	// nil means the op does not support --watch.
	Watch func(a Args) bool

	// OutputFlags names flags that only change how the output is printed (e.g. a format string).
	// Watchers that differ only in these share one query, and Render applies them for each.
	OutputFlags []string

	// Render turns the lines of the shared query into the lines for one watcher, using the
	// OutputFlags of its arguments. nil leaves the lines as they are.
	Render func(a Args, lines []string) []string

	// Poll returns an interval at which watched state is re-queried, for sources that emit no events.
	// nil or <= 0 disables polling.
	Poll func() time.Duration

	// Invalidates lists name prefixes of watchable ops whose state may change after Run.
	Invalidates []string

	// set by Register, nil for an op that is not registered
	stats *opStats

	// set by Registry.Bind
	flags       []FlagDef
	validate    func(pos []string) error
	subcommands map[string]bool
}

// canonical returns the arguments with output flags reset to their defaults, so that watchers
// differing only in these map to the same state.
func (op *Op) canonical(a Args) Args {
	if len(op.OutputFlags) == 0 {
		return a
	}
	flags := make(map[string]string, len(a.Flags))
	for k, v := range a.Flags {
		flags[k] = v
	}
	defaults := op.defaultArgs().Flags
	for _, name := range op.OutputFlags {
		if _, ok := flags[name]; ok {
			flags[name] = defaults[name]
		}
	}
	a.Flags = flags
	return a
}

// render applies the output flags of one watcher to the lines of the shared state.
func (op *Op) render(a Args, lines []string) []string {
	if op.Render == nil {
		return lines
	}
	return op.Render(a, lines)
}

func (op *Op) path() []string {
	return strings.Split(op.Name, ".")
}

func (op *Op) flagByName(name string) *FlagDef {
	for i := range op.flags {
		if op.flags[i].Name == name {
			return &op.flags[i]
		}
	}
	return nil
}

func (op *Op) flagByShort(short string) *FlagDef {
	for i := range op.flags {
		if op.flags[i].Short == short {
			return &op.flags[i]
		}
	}
	return nil
}

func (op *Op) defaultArgs() Args {
	a := Args{Flags: make(map[string]string, len(op.flags))}
	for _, f := range op.flags {
		if f.Name != WatchFlag {
			a.Flags[f.Name] = f.Default
		}
	}
	return a
}

// Execute runs the op once, honoring its lock. queuedAt is the time the user issued the command.
// A command that was dropped for waiting too long for its lock returns no lines and no error.
func (op *Op) Execute(ctx context.Context, a Args, queuedAt time.Time) ([]string, error) {
	start := time.Now()
	lines, dropped, err := op.execute(ctx, a, queuedAt)
	if op.stats != nil {
		op.stats.record(time.Since(start), dropped, err)
	}
	return lines, err
}

func (op *Op) execute(ctx context.Context, a Args, queuedAt time.Time) (lines []string, dropped bool, err error) {
	if err := op.ensureEnvironment(); err != nil {
		return nil, false, err
	}
	lock := op.Lock
	if op.LockFor != nil {
		lock = op.LockFor(a)
	}
	if lock == "" {
		lines, err = op.Run(ctx, a)
		return lines, false, err
	}
	dropped, err = WithLock(lock, queuedAt, func() error {
		var err error
		lines, err = op.Run(ctx, a)
		return err
	})
	return lines, dropped, err
}

// Always is the Op.Watch of ops whose every invocation can be watched.
func Always(Args) bool { return true }
