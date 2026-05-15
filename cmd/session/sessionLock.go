package session

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/markusressel/system-control/internal/util"
	"github.com/spf13/cobra"
)

const (
	lockSessionUser    = "markus"
	lockSessionDisplay = ":0"
	lockDetachedFlag   = "detached"
	lockRunnerFlag     = "internal-detached-lock-runner"
)

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Lock the current desktop session",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		isDetachedRunner, err := cmd.Flags().GetBool(lockRunnerFlag)
		if err != nil {
			return err
		}
		useDetached, err := cmd.Flags().GetBool(lockDetachedFlag)
		if err != nil {
			return err
		}

		if useDetached && !isDetachedRunner {
			return startDetachedLockProcess()
		}

		return sessionLockScript()
	},
}

func init() {
	Command.AddCommand(lockCmd)
	lockCmd.Flags().Bool(lockDetachedFlag, false, "run lock command in a detached background process")
	lockCmd.Flags().Bool(lockRunnerFlag, false, "internal flag used to run detached lock process")
	_ = lockCmd.Flags().MarkHidden(lockRunnerFlag)
}

func startDetachedLockProcess() error {
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	childArgs := append([]string{}, os.Args[1:]...)
	childArgs = append(childArgs, "--"+lockRunnerFlag)

	childProcess := exec.Command(executablePath, childArgs...)
	childProcess.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	childProcess.Stdin = nil
	childProcess.Stdout = io.Discard
	childProcess.Stderr = io.Discard

	if err := childProcess.Start(); err != nil {
		return fmt.Errorf("failed to start detached lock process: %w", err)
	}

	if err := childProcess.Process.Release(); err != nil {
		return fmt.Errorf("failed to release detached lock process: %w", err)
	}

	return nil
}

func sessionLockScript() (err error) {
	locked, err := isProcessRunning("i3lock")
	if err != nil {
		return err
	}
	if locked {
		return nil
	}

	if err := setSessionScreenTimeout(10, 10); err != nil {
		return err
	}
	if err := setSessionDPMSTimeout(15, 15, 15); err != nil {
		return err
	}
	defer func() {
		restoreErr := restoreSessionDPMSTimeout()
		if restoreErr == nil {
			return
		}

		if err == nil {
			err = restoreErr
			return
		}

		err = fmt.Errorf("%w; additionally failed to restore session DPMS timeout: %v", err, restoreErr)
	}()

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
	envArgs := []string{"DISPLAY=" + lockSessionDisplay}
	if currentPath := os.Getenv("PATH"); currentPath != "" {
		envArgs = append(envArgs, "PATH="+currentPath)
	}

	fullArgs := []string{"-u", lockSessionUser, "env"}
	fullArgs = append(fullArgs, envArgs...)
	fullArgs = append(fullArgs, command)
	fullArgs = append(fullArgs, args...)

	_, err := util.ExecCommand("sudo", fullArgs...)
	if err != nil {
		return fmt.Errorf("failed to execute %s: %w", command, err)
	}

	return nil
}

func setSessionDPMSTimeout(standby int, suspend int, off int) error {
	return runCommandAsSessionUser("xset", "dpms", fmt.Sprintf("%d", standby), fmt.Sprintf("%d", suspend), fmt.Sprintf("%d", off))
}

func restoreSessionDPMSTimeout() error {
	var restoreErr error

	if err := runCommandAsSessionUser("xset", "s", "0", "0"); err != nil {
		restoreErr = errors.Join(restoreErr, err)
	}
	if err := runCommandAsSessionUser("xset", "dpms", "0", "0", "0"); err != nil {
		restoreErr = errors.Join(restoreErr, err)
	}

	return restoreErr
}

func setSessionScreenTimeout(timeout int, cycle int) error {
	return runCommandAsSessionUser("xset", "s", fmt.Sprintf("%d", timeout), fmt.Sprintf("%d", cycle))
}

func forceSessionDPMS(mode string) error {
	return runCommandAsSessionUser("xset", "dpms", "force", mode)
}

func runSessionLock() error {
	return runCommandAsSessionUser(
		"i3lock",
		"--nofork",
		"--show-failed-attempts",
		"-c130003",
	)
}
