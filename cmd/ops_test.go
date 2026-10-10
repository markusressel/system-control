package cmd

import (
	"strings"
	"testing"

	"github.com/markusressel/system-control/internal/daemon"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// findFlag looks up a flag the way cobra resolves it for a command: local first, then inherited
// persistent flags of the parents.
func findFlag(cmd *cobra.Command, name string) *pflag.Flag {
	for c := cmd; c != nil; c = c.Parent() {
		if f := c.Flags().Lookup(name); f != nil {
			return f
		}
		if f := c.PersistentFlags().Lookup(name); f != nil {
			return f
		}
	}
	return nil
}

// The daemon parses raw command lines using the flag definitions of an op, while the cobra
// command is what runs without a daemon. Both must describe the same command line.
func TestOpsMatchCobraCommands(t *testing.T) {
	for _, op := range daemon.Default.Ops() {
		t.Run(op.Name, func(t *testing.T) {
			words := strings.Split(op.Name, ".")
			cmd, rest, err := RootCmd.Find(words)
			if err != nil || len(rest) != 0 || cmd == RootCmd {
				t.Fatalf("no cobra command for op path %v (err=%v, rest=%v)", words, err, rest)
			}
			if cmd.Name() != words[len(words)-1] {
				t.Fatalf("op resolves to command %q", cmd.CommandPath())
			}

			for _, def := range op.Flags {
				f := findFlag(cmd, def.Name)
				if f == nil {
					t.Errorf("flag --%s missing on %s", def.Name, cmd.CommandPath())
					continue
				}
				if f.Shorthand != def.Short {
					t.Errorf("flag --%s: shorthand %q, op declares %q", def.Name, f.Shorthand, def.Short)
				}
				if (f.Value.Type() == "bool") != (def.Kind == daemon.FlagBool) {
					t.Errorf("flag --%s: cobra type %s does not match op kind %d", def.Name, f.Value.Type(), def.Kind)
				}
				if (f.Value.Type() == "stringSlice") != (def.Kind == daemon.FlagList) {
					t.Errorf("flag --%s: cobra type %s does not match op kind %d", def.Name, f.Value.Type(), def.Kind)
				}
				defValue := strings.Trim(f.DefValue, "[]")
				if defValue != def.Default {
					t.Errorf("flag --%s: default %q, op declares %q", def.Name, defValue, def.Default)
				}
			}

			if op.Watch != nil {
				f := findFlag(cmd, "watch")
				if f == nil || f.Shorthand != "w" || f.Value.Type() != "bool" {
					t.Errorf("op is watchable but %s has no --watch/-w flag", cmd.CommandPath())
				}
			}
		})
	}
}
