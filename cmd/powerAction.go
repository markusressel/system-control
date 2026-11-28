package cmd

import (
	"fmt"
	"time"

	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

var (
	powerActionTimeout time.Duration
	powerActionForce   bool
)

// addPowerActionFlags adds the flags of commands that close all windows before a power action.
func addPowerActionFlags(command *cobra.Command) {
	command.Flags().DurationVar(&powerActionTimeout, "timeout", 30*time.Second, "How long to wait for windows to close")
	command.Flags().BoolVar(&powerActionForce, "force", false, "Continue even if some windows did not close in time")
}

// gracefulPowerAction closes all windows and then runs the given command (e.g. "reboot").
// If a window refuses to close (e.g. a "save changes?" dialog) nothing is done,
// unless --force is given.
func gracefulPowerAction(command string) error {
	remaining, err := util.CloseAllWindows(powerActionTimeout)
	if err != nil {
		return err
	}
	if len(remaining) > 0 && !powerActionForce {
		return fmt.Errorf("windows did not close within %v: %s (use --force to continue anyway)",
			powerActionTimeout, util.DescribeWindows(remaining))
	}

	_, err = util.ExecCommand(command)
	return err
}
