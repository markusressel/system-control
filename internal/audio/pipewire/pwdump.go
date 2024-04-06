package pipewire

import (
	"encoding/json"
	"github.com/markusressel/system-control/internal/util"
	"log"
	"strings"
)

func PwDump() PipewireState {
	result, err := util.ExecCommand("pw-dump")
	if err != nil {
		log.Fatal(err)
	}

	state, err := parsePwDumpToState(result)
	if err != nil {
		log.Fatal(err)
	}

	return state
}

func parsePwDumpToState(pwDump string) (PipewireState, error) {
	var state PipewireState
	if err := json.NewDecoder(strings.NewReader(pwDump)).Decode(&state); err != nil {
		return state, err
	}
	return state, nil
}
