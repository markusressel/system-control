package daemon

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

// EnvRuntimeDir overrides the directory of the daemon socket and lock files, e.g. to run a
// second daemon next to the real one. (Overriding XDG_RUNTIME_DIR itself would also move
// the PipeWire socket.)
const EnvRuntimeDir = "SYSTEM_CONTROL_RUNTIME_DIR"

// RuntimeDir returns the per-user directory that holds the daemon socket and lock files.
// $XDG_RUNTIME_DIR is honored so the daemon (started by systemd) and the CLI always agree.
func RuntimeDir() string {
	if dir := os.Getenv(EnvRuntimeDir); dir != "" {
		return dir
	}

	base := os.Getenv("XDG_RUNTIME_DIR")
	if base == "" {
		uid := fmt.Sprintf("%d", os.Getuid())
		if u, err := user.Current(); err == nil {
			uid = u.Uid
		}
		base = filepath.Join("/run/user", uid)
	}
	return filepath.Join(base, "system-control")
}

// SocketPath returns the path of the daemon's unix socket.
func SocketPath() string {
	return filepath.Join(RuntimeDir(), "ipc.sock")
}

// getuid is a variable for tests.
var getuid = os.Getuid

// notifySocketPath returns the socket of the daemon that has to learn about state a command
// changed. Under sudo that is the daemon of the invoking user, not one of root: sudo resets
// $XDG_RUNTIME_DIR. Only notifications go there, commands run as root are never handed to
// the user's daemon.
func notifySocketPath() string {
	if os.Getenv(EnvRuntimeDir) == "" && getuid() == 0 {
		if uid := os.Getenv("SUDO_UID"); uid != "" && uid != "0" {
			return filepath.Join("/run/user", uid, "system-control", "ipc.sock")
		}
	}
	return SocketPath()
}
