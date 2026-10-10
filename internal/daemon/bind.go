package daemon

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Bind connects every op to the cobra command at its CLI path and derives flags, positional
// argument validation and sub commands from it. This keeps the command tree the only
// declaration of the CLI surface: the daemon parses raw command lines exactly as cobra would.
//
// Flags of the root command (e.g. --configuration) are deliberately not understood, a command
// line using them runs through the regular CLI.
func (r *Registry) Bind(root *cobra.Command) error {
	for _, op := range r.ops {
		words := op.path()
		cmd, rest, err := root.Find(words)
		if err != nil || len(rest) != 0 || cmd == root || cmd.Name() != words[len(words)-1] {
			return fmt.Errorf("op %s: no command %q", op.Name, strings.Join(words, " "))
		}

		op.flags = []FlagDef{}
		// cobra merges inherited flags into cmd.Flags(), the root's must stay unknown
		seen := map[string]bool{}
		root.PersistentFlags().VisitAll(func(f *pflag.Flag) { seen[f.Name] = true })
		collect := func(set *pflag.FlagSet) {
			set.VisitAll(func(f *pflag.Flag) {
				if seen[f.Name] || f.Name == "help" {
					return
				}
				seen[f.Name] = true
				op.flags = append(op.flags, flagDef(f))
			})
		}
		collect(cmd.Flags())
		collect(cmd.PersistentFlags())
		for p := cmd.Parent(); p != nil && p.HasParent(); p = p.Parent() {
			collect(p.PersistentFlags())
		}

		if f := op.flagByName("watch"); op.Watch != nil && (f == nil || f.Kind != FlagBool) {
			return fmt.Errorf("op %s is watchable, but %q has no boolean --watch flag", op.Name, cmd.CommandPath())
		}

		op.validate = nil
		if cmd.Args != nil {
			c := cmd
			op.validate = func(pos []string) error { return c.Args(c, pos) }
		}

		op.subcommands = map[string]bool{}
		for _, sub := range cmd.Commands() {
			op.subcommands[sub.Name()] = true
			for _, alias := range sub.Aliases {
				op.subcommands[alias] = true
			}
		}
	}
	return nil
}

func flagDef(f *pflag.Flag) FlagDef {
	def := FlagDef{Name: f.Name, Short: f.Shorthand, Default: f.DefValue}
	switch f.Value.Type() {
	case "bool":
		def.Kind = FlagBool
	case "stringSlice":
		def.Kind = FlagList
		def.Default = strings.Trim(f.DefValue, "[]")
	case "int", "int8", "int16", "int32", "int64":
		def.valid = func(v string) error { _, err := strconv.ParseInt(v, 0, 64); return err }
	case "uint", "uint8", "uint16", "uint32", "uint64":
		def.valid = func(v string) error { _, err := strconv.ParseUint(v, 0, 64); return err }
	case "float32", "float64":
		def.valid = func(v string) error { _, err := strconv.ParseFloat(v, 64); return err }
	case "duration":
		def.valid = func(v string) error { _, err := time.ParseDuration(v); return err }
	}
	return def
}
