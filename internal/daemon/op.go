package daemon

import (
	"context"
	"fmt"
	"sort"
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

// FlagDef describes a CLI flag of an Op. The daemon parses raw argv with these definitions,
// the cobra command that implements the same Op must declare the same flags
// (verified by a test in package cmd).
type FlagDef struct {
	Name    string
	Short   string
	Kind    FlagKind
	Default string
	// Mutating flags turn an otherwise read-only invocation into a mutation
	// that the daemon does not handle. Setting one makes the CLI run locally.
	Mutating bool
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
type Op struct {
	// Name is the dotted CLI path, e.g. "audio.volume.inc" for `system-control audio volume inc`.
	Name   string
	Flags  []FlagDef
	MinPos int
	MaxPos int

	// Lock names a cross-process lock held while Run executes (mutating ops).
	Lock string

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
}

func (op *Op) path() []string {
	return strings.Split(op.Name, ".")
}

func (op *Op) flagByName(name string) *FlagDef {
	for i := range op.Flags {
		if op.Flags[i].Name == name {
			return &op.Flags[i]
		}
	}
	return nil
}

func (op *Op) flagByShort(short string) *FlagDef {
	for i := range op.Flags {
		if op.Flags[i].Short == short {
			return &op.Flags[i]
		}
	}
	return nil
}

// DefaultArgs returns Args with all flag defaults filled in.
func (op *Op) DefaultArgs() Args {
	a := Args{Flags: make(map[string]string, len(op.Flags))}
	for _, f := range op.Flags {
		a.Flags[f.Name] = f.Default
	}
	return a
}

// Execute runs the op once, honoring its lock. queuedAt is the time the user issued the command.
// A command that was dropped for waiting too long for its lock returns no lines and no error.
func (op *Op) Execute(ctx context.Context, a Args, queuedAt time.Time) ([]string, error) {
	if op.Lock == "" {
		return op.Run(ctx, a)
	}
	var lines []string
	_, err := WithLock(op.Lock, queuedAt, func() error {
		var err error
		lines, err = op.Run(ctx, a)
		return err
	})
	return lines, err
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

	a = op.DefaultArgs()
	rest := argv[i:]
	for j := 0; j < len(rest); j++ {
		tok := rest[j]

		switch {
		case tok == "--":
			a.Pos = append(a.Pos, rest[j+1:]...)
			j = len(rest)

		case strings.HasPrefix(tok, "--"):
			name, value, hasValue := strings.Cut(tok[2:], "=")
			if name == "watch" && op.Watch != nil {
				if !hasValue || value == "true" {
					a.Watch = true
				} else if value != "false" {
					return nil, Args{}, false
				}
				continue
			}
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
			if short == "w" && op.Watch != nil && !hasValue {
				a.Watch = true
				continue
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

	if len(a.Pos) < op.MinPos || len(a.Pos) > op.MaxPos {
		return nil, Args{}, false
	}
	if a.Watch && (op.Watch == nil || !op.Watch(a)) {
		return nil, Args{}, false
	}
	for _, f := range op.Flags {
		if f.Mutating && a.Flags[f.Name] != f.Default {
			return nil, Args{}, false
		}
	}

	return op, a, true
}

// applyFlag stores a flag value. consumed is the number of following tokens used as the value.
func applyFlag(a *Args, def *FlagDef, value string, hasValue bool, following []string) (consumed int, ok bool) {
	switch def.Kind {
	case FlagBool:
		if !hasValue {
			a.Flags[def.Name] = "true"
		} else if value == "true" || value == "false" {
			a.Flags[def.Name] = value
		} else {
			return 0, false
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
		a.Flags[def.Name] = value
		return consumed, true
	}
}
