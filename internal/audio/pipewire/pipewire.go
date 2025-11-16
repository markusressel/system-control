package pipewire

import (
	"fmt"
	"strconv"

	"github.com/markusressel/system-control/internal/util"
)

// RotateActiveSinkPipewire switches the default sink and moves all existing sink inputs to the next available sink in the list
func RotateActiveSinkPipewire(reverse bool) error {
	state := PwDump()
	allSinks := state.GetSinkNodes()
	activeNode, err := state.GetDefaultSinkNode()
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
		indexOfNextSink = (len(allSinks) + (indexOfActiveSink - 1)) % (len(allSinks))
	} else {
		indexOfNextSink = (indexOfActiveSink + 1) % (len(allSinks))
	}

	nextSink := allSinks[indexOfNextSink]
	return state.SwitchSinkTo(nextSink)
}

func moveStreamToNode(streamId int, nodeId int, objectId int) error {
	_, err := util.ExecCommand(
		"pw-metadata",
		strconv.Itoa(streamId),
		"target.node", strconv.Itoa(nodeId),
	)
	if err != nil {
		return err
	}

	_, err = util.ExecCommand(
		"pw-metadata",
		strconv.Itoa(streamId),
		"target.object", strconv.Itoa(objectId),
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

func WpCtlSetVolume(id int, volume float64) error {
	formattedVolume := fmt.Sprintf("%.3f", volume)
	return runWpCtl("set-volume", strconv.Itoa(id), formattedVolume)
}

func WpCtlSetMute(id int, mute bool) error {
	muteParam := "0"
	if mute {
		muteParam = "1"
	}
	return runWpCtl("set-mute", strconv.Itoa(id), muteParam)
}

func WpCtlToggleMute(id int) error {
	return runWpCtl("set-mute", strconv.Itoa(id), "toggle")
}

func runWpCtl(args ...string) error {
	_, err := util.ExecCommand(
		"wpctl", args...,
	)
	return err
}
