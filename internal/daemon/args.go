package daemon

import (
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

// WatchFlag is the boolean flag that makes a command watch its value (Args.Watch).
const WatchFlag = "watch"
