package pipewire

import (
	"github.com/markusressel/system-control/internal/util"
	"strconv"
)

// RotateActiveSinkPipewire switches the default sink and moves all existing sink inputs to the next available sink in the list
func RotateActiveSinkPipewire(reverse bool) error {
	state := PwDump()
	allSinks := state.GetSinkNodes()
	activeNode, err := state.GetDefaultNode()
	if err != nil {
		return err
	}

	indexOfActiveSink := -1
	for idx, sink := range allSinks {
		if sink.Id == activeNode.Id {
			indexOfActiveSink = idx
			break
		}
	}

	var indexOfNextSink = indexOfActiveSink
	if reverse {
		indexOfNextSink = len(allSinks) + (indexOfActiveSink-1)%(len(allSinks))
	} else {
		indexOfNextSink = (indexOfActiveSink + 1) % (len(allSinks))
	}

	nextSink := allSinks[indexOfNextSink]
	return state.SwitchSinkTo(nextSink)
}

func moveStreamToNode(streamId int, nodeId int) error {
	_, err := util.ExecCommand(
		"pw-metadata",
		strconv.Itoa(streamId),
		"target.node", strconv.Itoa(nodeId),
	)
	return err
}

// Switches the default sink to the target sink
// You need to get a sink name with "pw-cli ls Node"
// and look for the "node.name" property for a valid value.
func setDefaultSink(sinkName string) (err error) {
	_, err = util.ExecCommand("pw-metadata", "0", "default.configured.audio.sink", `{ "name": "`+sinkName+`" }`)
	return err
}
