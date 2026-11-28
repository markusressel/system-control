package session

import (
	"path/filepath"

	"github.com/markusressel/system-control/internal/configuration"
)

// maxCommLength is the number of characters the kernel keeps of a process name in /proc/<pid>/comm.
const maxCommLength = 15

// defaultLockCommand is used when session.lock.command is not configured.
var defaultLockCommand = []string{"i3lock", "--nofork", "--show-failed-attempts", "-c130003"}

// lockCommand returns the configured locker command line, which must stay in the foreground
// until the session is unlocked.
func lockCommand(config configuration.SessionLockConfig) []string {
	if len(config.Command) > 0 {
		return config.Command
	}
	return defaultLockCommand
}

// lockCommandLine returns the program and arguments that run the locker with its environment.
func lockCommandLine(config configuration.SessionLockConfig) (string, []string) {
	command := lockCommand(config)
	if len(config.Environment) == 0 {
		return command[0], command[1:]
	}
	return "env", append(append([]string{}, config.Environment...), command...)
}

// lockProcessName returns the name of the locker process as it appears in /proc/<pid>/comm.
// This is how a locked session is detected and unlocked.
func lockProcessName(config configuration.SessionLockConfig) string {
	name := config.ProcessName
	if name == "" {
		name = filepath.Base(lockCommand(config)[0])
	}
	if len(name) > maxCommLength {
		name = name[:maxCommLength]
	}
	return name
}

func currentLockProcessName() string {
	return lockProcessName(configuration.Get().Session.Lock)
}
