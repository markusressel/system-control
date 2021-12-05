package internal

import (
	"os/user"
	path2 "path"
)

var currentUser, _ = user.Current()
var ConfigBaseDir = path2.Join(currentUser.HomeDir, ".config", "system-control")
