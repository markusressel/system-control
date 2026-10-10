package commandutil

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/markusressel/system-control/internal/daemon"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// OpRunE returns the RunE of a cobra command that implements the registered op `opName`.
//
// A running daemon gets the chance to handle the command line before cobra even starts (see
// cmd.Execute), so this is the in-process implementation: it runs the very same op.
func OpRunE(opName string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		queuedAt := time.Now()

		op, ok := daemon.Default.Lookup(opName)
		if !ok {
			return fmt.Errorf("op %s is not registered", opName)
		}
		a := ArgsFromCommand(cmd, args)

		lines, err := op.Execute(cmd.Context(), a, queuedAt)
		if err != nil {
			return err
		}
		for _, line := range lines {
			fmt.Println(line)
		}

		// A daemon, if there is one, did not handle this invocation and
		// has to learn about the changed state.
		daemon.Notify(op.Invalidates...)

		if a.Watch {
			return errors.New("daemon is not running, continuous watching is disabled")
		}
		return nil
	}
}

// ArgsFromCommand collects the arguments of a parsed cobra command for an op.
// Every flag of the command is included, ops pick the ones they know by name.
func ArgsFromCommand(cmd *cobra.Command, args []string) daemon.Args {
	a := daemon.Args{Flags: map[string]string{}, Pos: args}
	a.Watch, _ = cmd.Flags().GetBool("watch")

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "watch" {
			return
		}
		if f.Value.Type() == "stringSlice" {
			if values, err := cmd.Flags().GetStringSlice(f.Name); err == nil {
				a.Flags[f.Name] = strings.Join(values, ",")
				return
			}
		}
		a.Flags[f.Name] = f.Value.String()
	})
	return a
}
