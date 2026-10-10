package pipewire

import (
	"encoding/json"
	"strings"

	"github.com/markusressel/system-control/internal/util"
)

// pwDumpDirect queries the full graph by running pw-dump.
func pwDumpDirect() (GraphState, error) {
	result, err := util.ExecCommand("pw-dump")
	if err != nil {
		return GraphState{}, err
	}

	return parsePwDumpToState(result)
}

func parsePwDumpToState(pwDump string) (GraphState, error) {
	var state GraphState
	if err := json.NewDecoder(strings.NewReader(pwDump)).Decode(&state); err != nil {
		return state, err
	}
	return state, nil
}
