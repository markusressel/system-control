package daemon

import (
	"strings"
)

// MatchArgv resolves a raw command line (without the program name) to an op and its arguments.
// ok is false for everything the registry does not fully understand; the caller must then
// run the command through the regular CLI, which produces the proper help and error messages.
//
// The accepted shape is `<command words...> [flags and positional arguments...]`.
func (r *Registry) MatchArgv(argv []string) (op *Op, a Args, ok bool) {
	op, rest := r.lookup(argv)
	if op == nil || op.flags == nil { // unknown, or not bound to a command
		return nil, Args{}, false
	}
	a, ok = op.parseArgs(rest)
	if !ok || !op.accepts(a) {
		return nil, Args{}, false
	}
	return op, a, true
}

// lookup follows the command words of argv through the command tree. It returns the op of the
// deepest command reached (nil if it has none) and the arguments after the command words.
func (r *Registry) lookup(argv []string) (op *Op, rest []string) {
	n := r.root
	i := 0
	for ; i < len(argv); i++ {
		child, found := n.children[argv[i]]
		if !found {
			break
		}
		n = child
	}
	return n.op, argv[i:]
}

// parseArgs parses flags and positional arguments like cobra does, as far as it understands them.
func (op *Op) parseArgs(rest []string) (a Args, ok bool) {
	a = op.defaultArgs()
	for j := 0; j < len(rest); j++ {
		tok := rest[j]
		if tok == "--" {
			a.Pos = append(a.Pos, rest[j+1:]...)
			break
		}
		if len(tok) < 2 || tok[0] != '-' {
			a.Pos = append(a.Pos, tok)
			continue
		}
		def, value, hasValue, ok := op.parseFlagToken(tok)
		if !ok {
			return Args{}, false
		}
		consumed, ok := applyFlag(&a, def, value, hasValue, rest[j+1:])
		if !ok {
			return Args{}, false
		}
		j += consumed
	}
	return a, true
}

// parseFlagToken splits a "--name[=value]" or "-n[value]" token into its flag and value.
func (op *Op) parseFlagToken(tok string) (def *FlagDef, value string, hasValue bool, ok bool) {
	if name, ok := strings.CutPrefix(tok, "--"); ok {
		name, value, hasValue = strings.Cut(name, "=")
		def = op.flagByName(name)
		return def, value, hasValue, def != nil
	}

	value, hasValue = tok[2:], len(tok) > 2
	if hasValue && value[0] == '=' {
		value = value[1:]
	}
	def = op.flagByShort(tok[1:2])
	// "-xyz" is only understood as "-x yz" for value flags. Anything else (e.g. combined
	// boolean shorts) is left to cobra.
	if def == nil || (def.Kind == FlagBool && hasValue) {
		return nil, "", false, false
	}
	return def, value, hasValue, true
}

// accepts checks parsed arguments against what the command allows.
func (op *Op) accepts(a Args) bool {
	// the first positional would select another (sub)command, e.g. `brightness inc`
	if len(a.Pos) > 0 && op.subcommands[a.Pos[0]] {
		return false
	}
	if op.validate != nil && op.validate(a.Pos) != nil {
		return false
	}
	return !a.Watch || (op.Watch != nil && op.Watch(a))
}

// applyFlag stores a flag value. consumed is the number of following tokens used as the value.
func applyFlag(a *Args, def *FlagDef, value string, hasValue bool, following []string) (consumed int, ok bool) {
	if def.Kind == FlagBool {
		if !hasValue {
			value = "true"
		} else if value != "true" && value != "false" {
			return 0, false
		}
		if def.Name == WatchFlag {
			a.Watch = value == "true"
		} else {
			a.Flags[def.Name] = value
		}
		return 0, true
	}

	if !hasValue {
		if len(following) == 0 {
			return 0, false
		}
		value, consumed = following[0], 1
	}
	switch {
	case def.Kind == FlagList:
		// A list flag replaces its default on first use and accumulates afterwards.
		if existing := a.Flags[def.Name]; existing != "" && existing != def.Default {
			value = existing + "," + value
		}
	case def.valid != nil && def.valid(value) != nil:
		return 0, false
	}
	a.Flags[def.Name] = value
	return consumed, true
}
