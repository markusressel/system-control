package daemon

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// FlagKind describes how a flag value is transported on the wire.
type FlagKind int

const (
	FlagString FlagKind = iota
	FlagBool
	// FlagList is a comma separated list, may be repeated on the command line.
	FlagList
)

// FlagDef describes a CLI flag of an Op. Definitions are derived from the cobra command
// that implements the op (see Registry.Bind), cobra is the single source of the CLI surface.
type FlagDef struct {
	Name    string
	Short   string
	Kind    FlagKind
	Default string

	// valid rejects values cobra would reject (e.g. a number flag), nil accepts everything.
	// A command line with an invalid value is left to cobra, which produces the error message.
	valid func(value string) error
}

// Args are the parsed arguments of one Op invocation.
type Args struct {
	// Flags contains a value for every FlagDef of the Op (defaults included).
	Flags map[string]string
	Pos   []string
	Watch bool
}

// Flag returns the raw value of a flag.
func (a Args) Flag(name string) string {
	return a.Flags[name]
}

// Arg returns the i-th positional argument, or an error if it is missing. Cobra validates the
// count of a command line, but an op must not rely on it to avoid an index panic.
func (a Args) Arg(i int) (string, error) {
	if i < 0 || i >= len(a.Pos) {
		return "", fmt.Errorf("missing argument %d", i+1)
	}
	return a.Pos[i], nil
}

// Int returns the value of a flag as integer.
func (a Args) Int(name string) (int, error) {
	v, err := strconv.Atoi(a.Flags[name])
	if err != nil {
		return 0, fmt.Errorf("invalid --%s %q: not an integer", name, a.Flags[name])
	}
	return v, nil
}

// Float returns the value of a flag as number.
func (a Args) Float(name string) (float64, error) {
	v, err := strconv.ParseFloat(a.Flags[name], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid --%s %q: not a number", name, a.Flags[name])
	}
	return v, nil
}

// Duration returns the value of a flag as duration.
func (a Args) Duration(name string) (time.Duration, error) {
	v, err := time.ParseDuration(a.Flags[name])
	if err != nil {
		return 0, fmt.Errorf("invalid --%s %q: not a duration", name, a.Flags[name])
	}
	return v, nil
}

// List returns the values of a FlagList flag.
func (a Args) List(name string) []string {
	v := a.Flags[name]
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}

// key identifies the state an invocation observes.
func (a Args) key(opName string) string {
	names := make([]string, 0, len(a.Flags))
	for k := range a.Flags {
		names = append(names, k)
	}
	sort.Strings(names)

	var sb strings.Builder
	sb.WriteString(opName)
	for _, k := range names {
		fmt.Fprintf(&sb, "\x00%s=%s", k, a.Flags[k])
	}
	for _, p := range a.Pos {
		fmt.Fprintf(&sb, "\x00%s", p)
	}
	return sb.String()
}

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
		if f.Name != "watch" {
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

type node struct {
	children map[string]*node
	op       *Op
}

// Registry holds all known ops.
type Registry struct {
	ops  map[string]*Op
	root *node
}

func NewRegistry() *Registry {
	return &Registry{ops: map[string]*Op{}, root: &node{children: map[string]*node{}}}
}

// Default is the registry the application's packages register their ops in.
var Default = NewRegistry()

// Register adds an op to the default registry. Intended to be called from init().
func Register(op Op) {
	Default.Register(op)
}

func (r *Registry) Register(op Op) {
	if _, exists := r.ops[op.Name]; exists {
		panic("daemon: duplicate op " + op.Name)
	}
	o := op
	o.stats = &opStats{}
	r.ops[o.Name] = &o

	n := r.root
	for _, word := range o.path() {
		child, ok := n.children[word]
		if !ok {
			child = &node{children: map[string]*node{}}
			n.children[word] = child
		}
		n = child
	}
	n.op = &o
}

// Lookup finds an op by name.
func (r *Registry) Lookup(name string) (*Op, bool) {
	op, ok := r.ops[name]
	return op, ok
}

// Ops returns all registered ops sorted by name.
func (r *Registry) Ops() []*Op {
	result := make([]*Op, 0, len(r.ops))
	for _, op := range r.ops {
		result = append(result, op)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

// MatchArgv resolves a raw command line (without the program name) to an op and its arguments.
// ok is false for everything the registry does not fully understand; the caller must then
// run the command through the regular CLI, which produces the proper help and error messages.
//
// The accepted shape is `<command words...> [flags and positional arguments...]`.
func (r *Registry) MatchArgv(argv []string) (op *Op, a Args, ok bool) {
	n := r.root
	i := 0
	for ; i < len(argv); i++ {
		child, found := n.children[argv[i]]
		if !found {
			break
		}
		n = child
	}
	if n.op == nil {
		return nil, Args{}, false
	}
	op = n.op
	if op.flags == nil {
		return nil, Args{}, false // not bound to a command
	}

	a = op.defaultArgs()
	rest := argv[i:]
	for j := 0; j < len(rest); j++ {
		tok := rest[j]

		switch {
		case tok == "--":
			a.Pos = append(a.Pos, rest[j+1:]...)
			j = len(rest)

		case strings.HasPrefix(tok, "--"):
			name, value, hasValue := strings.Cut(tok[2:], "=")
			def := op.flagByName(name)
			if def == nil {
				return nil, Args{}, false
			}
			consumed, good := applyFlag(&a, def, value, hasValue, rest[j+1:])
			if !good {
				return nil, Args{}, false
			}
			j += consumed

		case len(tok) > 1 && tok[0] == '-':
			short := tok[1:2]
			value, hasValue := tok[2:], len(tok) > 2
			if hasValue && value[0] == '=' {
				value = value[1:]
			}
			def := op.flagByShort(short)
			if def == nil {
				return nil, Args{}, false
			}
			// "-xyz" is only understood as "-x yz" for value flags. Anything else (e.g. combined
			// boolean shorts) is left to cobra.
			if def.Kind == FlagBool && hasValue {
				return nil, Args{}, false
			}
			consumed, good := applyFlag(&a, def, value, hasValue, rest[j+1:])
			if !good {
				return nil, Args{}, false
			}
			j += consumed

		default:
			a.Pos = append(a.Pos, tok)
		}
	}

	// the first positional would select another (sub)command, e.g. `brightness inc`
	if len(a.Pos) > 0 && op.subcommands[a.Pos[0]] {
		return nil, Args{}, false
	}
	if op.validate != nil && op.validate(a.Pos) != nil {
		return nil, Args{}, false
	}
	if a.Watch && (op.Watch == nil || !op.Watch(a)) {
		return nil, Args{}, false
	}

	return op, a, true
}

// applyFlag stores a flag value. consumed is the number of following tokens used as the value.
func applyFlag(a *Args, def *FlagDef, value string, hasValue bool, following []string) (consumed int, ok bool) {
	switch def.Kind {
	case FlagBool:
		if !hasValue {
			value = "true"
		} else if value != "true" && value != "false" {
			return 0, false
		}
		if def.Name == "watch" {
			a.Watch = value == "true"
		} else {
			a.Flags[def.Name] = value
		}
		return 0, true

	case FlagList:
		if !hasValue {
			if len(following) == 0 {
				return 0, false
			}
			value = following[0]
			consumed = 1
		}
		// A list flag replaces its default on first use and accumulates afterwards.
		if existing := a.Flags[def.Name]; existing != "" && existing != def.Default {
			value = existing + "," + value
		}
		a.Flags[def.Name] = value
		return consumed, true

	default:
		if !hasValue {
			if len(following) == 0 {
				return 0, false
			}
			value = following[0]
			consumed = 1
		}
		if def.valid != nil && def.valid(value) != nil {
			return 0, false
		}
		a.Flags[def.Name] = value
		return consumed, true
	}
}
