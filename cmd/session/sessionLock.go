package session

import (
	"fmt"
	"time"

	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

const (
	lockSessionUser    = "markus"
	lockSessionDisplay = ":0"
)

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Lock the current desktop session",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return sessionLockScript()
	},
}

func init() {
	Command.AddCommand(lockCmd)
}

func sessionLockScript() error {
	locked, err := isProcessRunning("i3lock")
	if err != nil {
		return err
	}
	if locked {
		return nil
	}

	if err := setSessionDPMSTimeout(15, 15, 15); err != nil {
		return err
	}
	defer restoreSessionDPMSTimeout()

	// Allow releasing the lock keybind before forcing displays off.
	time.Sleep(500 * time.Millisecond)

	if err := forceSessionDPMS("suspend"); err != nil {
		return err
	}
	if err := forceSessionDPMS("standby"); err != nil {
		return err
	}
	if err := forceSessionDPMS("off"); err != nil {
		return err
	}

	if err := runSessionLock(); err != nil {
		return err
	}

	return nil
}

func runCommandAsSessionUser(command string, args ...string) error {
	fullArgs := []string{"-u", lockSessionUser, "env", "DISPLAY=" + lockSessionDisplay, command}
	fullArgs = append(fullArgs, args...)

	_, err := util.ExecCommand("sudo", fullArgs...)
	if err != nil {
		return fmt.Errorf("failed to execute %s: %w", command, err)
	}

	return nil
}

func setSessionDPMSTimeout(standby int, suspend int, off int) error {
	if err := runCommandAsSessionUser("xset", "s", "off"); err != nil {
		return err
	}
	return runCommandAsSessionUser("xset", "dpms", fmt.Sprintf("%d", standby), fmt.Sprintf("%d", suspend), fmt.Sprintf("%d", off))
}

func restoreSessionDPMSTimeout() {
	_ = runCommandAsSessionUser("xset", "s", "off")
	_ = runCommandAsSessionUser("xset", "dpms", "0", "0", "0")
}

func forceSessionDPMS(mode string) error {
	return runCommandAsSessionUser("xset", "dpms", "force", mode)
}

func runSessionLock() error {
	return runCommandAsSessionUser(
		"xss-lock",
		"--transfer-sleep-lock",
		"--",
		"i3lock",
		"--nofork",
		"--show-failed-attempts",
		"--ignore-empty-password",
		"-c130003",
	)
}
