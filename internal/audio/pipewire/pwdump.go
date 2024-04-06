package pipewire

import (
	"encoding/json"
	"fmt"
	"github.com/markusressel/system-control/internal/util"
	"log"
	"strings"
)

func PwDump() PipewireState {
	result, err := util.ExecCommand("pw-dump")
	if err != nil {
		log.Fatal(err)
	}

	var state PipewireState
	if err := json.NewDecoder(strings.NewReader(result)).Decode(&state); err != nil {
		log.Fatalf("decode: %s", err)
	}

	defaultSinkName, err := state.GetDefaultSink()
	fmt.Println("Default sink: ", defaultSinkName)

	defaultSourceName, err := state.GetDefaultSource()
	fmt.Println("Default source: ", defaultSourceName)

	//port, err := state.GetPortByType("PipeWire:Interface:Port", "Audio/Source")

	return state
}
