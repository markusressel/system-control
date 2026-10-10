package daemon

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/markusressel/system-control/internal/util"
)

// companionEnv lists variables that belong to a required one and are imported with it.
var companionEnv = map[string][]string{
	"DISPLAY": {"XAUTHORITY"},
}

var importMu sync.Mutex

// ensureEnvironment makes sure the variables in Op.Requires are set in a daemon process.
//
// The daemon is started by the user service manager, possibly before the session imported
// DISPLAY and friends into it. Missing variables are then taken from the manager's environment.
// If a variable is still missing the op can't run here: ErrUnhandled makes the client, which
// has the session environment, run it. In the CLI the environment is the caller's own.
func (op *Op) ensureEnvironment() error {
	if !Running() {
		return nil
	}
	missing := op.missingEnvironment()
	if len(missing) == 0 {
		return nil
	}

	importMu.Lock()
	defer importMu.Unlock()
	importSessionEnvironment(op.Requires)

	if missing = op.missingEnvironment(); len(missing) > 0 {
		return fmt.Errorf("%w: %s not set", ErrUnhandled, strings.Join(missing, ", "))
	}
	return nil
}

func (op *Op) missingEnvironment() (missing []string) {
	for _, name := range op.Requires {
		if os.Getenv(name) == "" {
			missing = append(missing, name)
		}
	}
	return missing
}

// importSessionEnvironment copies the given variables, and their companions, of the user
// service manager into this process. Variables that are already set are kept.
func importSessionEnvironment(names []string) {
	wanted := map[string]bool{}
	for _, name := range names {
		wanted[name] = true
		for _, c := range companionEnv[name] {
			wanted[c] = true
		}
	}

	out, err := util.ExecCommand("systemctl", "--user", "show-environment")
	if err != nil {
		return
	}
	for _, line := range strings.Split(out, "\n") {
		name, value, ok := strings.Cut(line, "=")
		if ok && wanted[name] && os.Getenv(name) == "" {
			_ = os.Setenv(name, value)
		}
	}
}
