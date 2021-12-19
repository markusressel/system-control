package internal

import (
	"os/user"
	"path"
)

var currentUser, _ = user.Current()
var ConfigBaseDir = path.Join(currentUser.HomeDir, ".config", "system-control")
