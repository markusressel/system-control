package touchpad

import (
	"strconv"

	"github.com/markusressel/system-control/internal/util"
)

// SetArg parses a boolean given on the command line ("true", "0", ...) and enables or
// disables the touchpad accordingly.
func SetArg(arg string) error {
	enabled, err := strconv.ParseBool(arg)
	if err != nil {
		return err
	}
	return util.SetTouchpadEnabled(enabled)
}

// Toggle enables the touchpad if it is disabled and disables it otherwise.
func Toggle() error {
	enabled, err := util.IsTouchpadEnabled()
	if err != nil {
		return err
	}
	return util.SetTouchpadEnabled(!enabled)
}
