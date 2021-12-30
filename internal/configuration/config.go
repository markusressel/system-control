package configuration

import (
	"os"
	"os/user"
	"path"
)

var currentUser, _ = user.Current()
var BaseDir = path.Join(currentUser.HomeDir, ".config", "system-control")

func init() {
	_ = ensureConfigDirExists()
}

func ensureConfigDirExists() error {
	return os.MkdirAll(BaseDir, 0755)
}
