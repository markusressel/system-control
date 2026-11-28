package session

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"syscall"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/util"
)

const (
	// LockDetachedFlag is the name of the flag of `session lock` that locks in a detached background process.
	LockDetachedFlag = "detached"
	// LockRunnerFlag is the name of the hidden flag of `session lock` that the detached background process runs with.
	LockRunnerFlag = "internal-detached-lock-runner"
)

// Session is the desktop session that is locked and unlocked. Empty fields are resolved
// when they are used: User falls back to $SUDO_USER, else the current user, Display to
// $DISPLAY, else ":0".
type Session struct {
	User    string
	Display string
}

// StartDetachedLockProcess runs the current command line again with LockRunnerFlag in a new
// session, without waiting for it.
func StartDetachedLockProcess() error {
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	childArgs := append([]string{}, os.Args[1:]...)
	childArgs = append(childArgs, "--"+LockRunnerFlag)

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

// Lock switches the displays off and runs the screen locker until the session is unlocked.
// It does nothing if the session is already locked.
func (s Session) Lock() (err error) {
	locked, err := isProcessRunning(currentLockProcessName())
	if err != nil {
		return err
	}
	if locked {
		return nil
	}

	if err := s.setScreenTimeout(10, 10); err != nil {
		return err
	}
	if err := s.setDPMSTimeout(15, 15, 15); err != nil {
		return err
	}
	defer func() {
		restoreErr := s.restoreDPMSTimeout()
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

	if err := s.forceDPMS("suspend"); err != nil {
		return err
	}
	if err := s.forceDPMS("standby"); err != nil {
		return err
	}
	if err := s.forceDPMS("off"); err != nil {
		return err
	}

	if err := s.runLock(); err != nil {
		return err
	}

	return nil
}

// Unlock terminates the screen locker and restores the DPMS timeouts.
func (s Session) Unlock() error {
	terminateErr := terminateProcessesByName(currentLockProcessName())
	restoreErr := s.restoreDPMSTimeout()

	if terminateErr != nil && restoreErr != nil {
		return fmt.Errorf("failed to unlock session: %w; additionally failed to restore session DPMS timeout: %v", terminateErr, restoreErr)
	}

	if terminateErr != nil {
		return terminateErr
	}

	return restoreErr
}

// sessionCommand builds the command line that runs a command as the session user on the
// session display. sudo is only used if the current user is not the session user.
func sessionCommand(sessionUser string, currentUser string, display string, path string, command string, args ...string) (string, []string) {
	envArgs := []string{"DISPLAY=" + display}
	if path != "" {
		envArgs = append(envArgs, "PATH="+path)
	}

	var fullArgs []string
	name := "env"
	if sessionUser != currentUser {
		name = "sudo"
		fullArgs = append(fullArgs, "-u", sessionUser, "env")
	}
	fullArgs = append(fullArgs, envArgs...)
	fullArgs = append(fullArgs, command)
	fullArgs = append(fullArgs, args...)
	return name, fullArgs
}

// runCommandAsSessionUser runs a short lived command in the desktop session.
func (s Session) runCommandAsSessionUser(command string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.runInSessionContext(ctx, util.ExecCommandContext, command, args...)
}

type contextRunner func(ctx context.Context, command string, args ...string) (string, error)

func (s Session) runInSessionContext(ctx context.Context, run contextRunner, command string, args ...string) error {
	currentUser := ""
	if u, err := user.Current(); err == nil {
		currentUser = u.Username
	}
	name, fullArgs := sessionCommand(s.resolveUser(), currentUser, s.resolveDisplay(), os.Getenv("PATH"), command, args...)

	if _, err := run(ctx, name, fullArgs...); err != nil {
		return fmt.Errorf("failed to execute %s: %w", command, err)
	}
	return nil
}

// resolveUser returns the user whose desktop session is controlled: the --user flag,
// else the user that invoked sudo, else the current user.
func (s Session) resolveUser() string {
	if s.User != "" {
		return s.User
	}
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		return sudoUser
	}
	if u, err := user.Current(); err == nil {
		return u.Username
	}
	return ""
}

// resolveDisplay returns the X display of the session: the --display flag, else $DISPLAY, else ":0".
func (s Session) resolveDisplay() string {
	if s.Display != "" {
		return s.Display
	}
	if display := os.Getenv("DISPLAY"); display != "" {
		return display
	}
	return ":0"
}

func (s Session) setDPMSTimeout(standby int, suspend int, off int) error {
	return s.runCommandAsSessionUser("xset", "dpms", fmt.Sprintf("%d", standby), fmt.Sprintf("%d", suspend), fmt.Sprintf("%d", off))
}

func (s Session) restoreDPMSTimeout() error {
	var restoreErr error

	if err := s.runCommandAsSessionUser("xset", "s", "0", "0"); err != nil {
		restoreErr = errors.Join(restoreErr, err)
	}
	if err := s.runCommandAsSessionUser("xset", "dpms", "0", "0", "0"); err != nil {
		restoreErr = errors.Join(restoreErr, err)
	}

	return restoreErr
}

func (s Session) setScreenTimeout(timeout int, cycle int) error {
	return s.runCommandAsSessionUser("xset", "s", fmt.Sprintf("%d", timeout), fmt.Sprintf("%d", cycle))
}

func (s Session) forceDPMS(mode string) error {
	return s.runCommandAsSessionUser("xset", "dpms", "force", mode)
}

// runLock runs the screen locker until the session is unlocked. It must not be
// bound by a timeout: killing the locker would unlock the session.
func (s Session) runLock() error {
	command, args := lockCommandLine(configuration.Get().Session.Lock)
	return s.runInSessionContext(context.Background(), util.ExecCommandContext, command, args...)
}
