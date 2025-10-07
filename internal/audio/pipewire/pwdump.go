package pipewire

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/markusressel/system-control/internal/util"
)

func PwDump() GraphState {
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

func parsePwDumpToState(pwDump string) (GraphState, error) {
	var state GraphState
	if err := json.NewDecoder(strings.NewReader(pwDump)).Decode(&state); err != nil {
		return state, err
	}
	return state, nil
}
