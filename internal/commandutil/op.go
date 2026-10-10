package commandutil

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/markusressel/system-control/internal/daemon"
	"github.com/spf13/cobra"
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
		a := ArgsFromCommand(cmd, args, op)

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
func ArgsFromCommand(cmd *cobra.Command, args []string, op *daemon.Op) daemon.Args {
	a := op.DefaultArgs()
	a.Pos = args
	a.Watch, _ = cmd.Flags().GetBool("watch")

	for _, def := range op.Flags {
		flag := cmd.Flags().Lookup(def.Name)
		if flag == nil {
			continue
		}
		if def.Kind == daemon.FlagList {
			values, err := cmd.Flags().GetStringSlice(def.Name)
			if err == nil {
				a.Flags[def.Name] = strings.Join(values, ",")
				continue
			}
		}
		a.Flags[def.Name] = flag.Value.String()
	}
	return a
}
