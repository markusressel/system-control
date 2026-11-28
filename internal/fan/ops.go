package fan

import (
	"context"
	"fmt"
	"time"

	"github.com/markusressel/system-control/internal/configuration"
	"github.com/markusressel/system-control/internal/daemon"
)

// The fan mode can be changed by hardware keys without any event, so it has to be polled.
const defaultPollInterval = 5 * time.Second

func init() {
	daemon.Register(daemon.Op{
		Name:  "fan.mode",
		Run:   runMode,
		Watch: daemon.Always,
		Poll:  configuration.PollInterval("fan", defaultPollInterval),
	})
}

func runMode(ctx context.Context, a daemon.Args) ([]string, error) {
	mode, err := ReadMode()
	if err != nil {
		return nil, err
	}

	switch modeType := a.Flag(TypeFlag); modeType {
	case TypeNumber:
		return []string{fmt.Sprint(mode)}, nil
	case TypeText:
		if modeName, ok := modeNames[mode]; ok {
			return []string{modeName}, nil
		}
		return []string{fmt.Sprintf("Unknown (%d)", mode)}, nil
	default:
		return nil, fmt.Errorf("invalid value for --%s: %q (expected %q or %q)", TypeFlag, modeType, TypeText, TypeNumber)
	}
}
