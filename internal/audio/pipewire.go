package audio

import (
	"github.com/markusressel/system-control/internal/util"
	"log"
	"strconv"
	"strings"
)

type PropertyFilter struct {
	key   string
	value string
}

func parsePipwireToMap(input string) []map[string]string {
	var result = make([]map[string]string, 0, 1000)

	lines := strings.Split(input, "\n")
	var objectMap map[string]string
	for _, line := range lines {
		if len(strings.TrimSpace(line)) <= 0 {
			continue
		}
		if strings.Contains(line, ",") && !strings.Contains(line, "=") {
			// this is the "header" of an object

			// create a new map for the current object and fill it
			objectMap = make(map[string]string)
			splits := strings.Split(line, ",")
			for _, item := range splits {
				item = strings.TrimSpace(item)
				splits := strings.SplitAfter(item, " ")
				objectMap[strings.TrimSpace(splits[0])] = strings.TrimSpace(splits[1])
			}
			result = append(result, objectMap)
		} else {
			// this is a property of an object

			splits := strings.SplitAfter(line, "=")

			key := strings.TrimRight(splits[0], "=")
			key = strings.TrimSpace(key)

			value := strings.TrimSpace(splits[1])
			value = strings.Trim(value, "\"")
			objectMap[key] = value
		}
	}

	return result
}

// FilterPipwireObjects filters the given pipewire object map based on the given function
func FilterPipwireObjects(vs []map[string]string, f func(map[string]string) bool) []map[string]string {
	vsf := make([]map[string]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// Switches the default sink to the target sink
// You need to get a sink name with "pw-cli ls Node"
// and look for the "node.name" property for a valid value.
func setDefaultSinkPipewire(sinkName string) (err error) {
	_, err = util.ExecCommand("pw-metadata", "0", "default.configured.audio.sink", `{ "name": "`+sinkName+`" }`)
	return err
}

// RotateActiveSinkPipewire switches the default sink and moves all existing sink inputs to the next available sink in the list
func RotateActiveSinkPipewire(reverse bool) {
	activeSink := FindActiveSinkPipewire("")

	objects := getPipewireObjects(
		PropertyFilter{"media.class", "Audio/Sink"},
	)

	for idx, item := range objects {
		if item["id"] == activeSink["id"] {
			offset := 1
			if reverse {
				offset = -1
			}
			nextIndex := (len(objects) + (idx + offset)) % len(objects)
			nextSink := objects[nextIndex]
			SwitchSinkPipewire(nextSink)
		}
	}

	log.Fatal("Active sink not found")
}

// SwitchSinkPipewire switches the default sink and moves all existing sink inputs to the target sink
func SwitchSinkPipewire(node map[string]string) {
	nodeName := node["node.name"]
	nodeId, err := strconv.Atoi(node["id"])
	if err != nil {
		log.Fatal(err)
	}
	err = setDefaultSinkPipewire(nodeName)
	if err != nil {
		log.Fatal(err)
	}

	var streams = getPipewireObjects(
		PropertyFilter{"media.class", "Stream/Output/Audio"},
	)
	for _, stream := range streams {
		moveStreamToNode(stream["id"], nodeId)
	}
}

func moveStreamToNode(streamId string, nodeId int) {
	_, err := util.ExecCommand("pw-metadata", streamId, "target.node", strconv.Itoa(nodeId))
	if err != nil {
		log.Fatal(err)
	}
}

// SetVolumePipewire sets the given volume to the given sink using pipewire
// volume in percent
func SetVolumePipewire(sinkId int, volume int) error {
	//objects := getPipewireObjects(
	//	PropertyFilter{"media.class", "Audio/Sink"},
	//)

	_, err := util.ExecCommand("pactl", "set-sink-volume", strconv.Itoa(sinkId), strconv.Itoa(volume)+"%")
	return err
}

// retrieve a list of pipewire objects
// optionally filtered
func getPipewireObjects(filters ...PropertyFilter) (objects []map[string]string) {
	result, err := util.ExecCommand("pw-cli", "ls")
	if err != nil {
		log.Fatal(err)
	}

	objects = parsePipwireToMap(result)
	objects = FilterPipwireObjects(objects, func(v map[string]string) bool {
		for _, filter := range filters {
			if v[filter.key] != filter.value {
				return false
			}
		}

		return true
	})

	return objects
}

// FindActiveSinkPipewire returns the index of the active sink
// or 0 if the given text is NOT found in the active sink
// or 1 if the given text IS found in the active sink
func FindActiveSinkPipewire(text string) map[string]string {
	// ignore case
	text = strings.ToLower(text)

	currentDefaultSinkName, err := util.ExecCommand("pactl", "get-default-sink")
	if err != nil {
		log.Fatal(err)
	}

	var activeSink map[string]string
	objects := getPipewireObjects(
		PropertyFilter{"media.class", "Audio/Sink"},
	)
	for _, item := range objects {
		if util.ContainsIgnoreCase(item["node.name"], currentDefaultSinkName) {
			activeSink = item
		}
	}

	return activeSink
}

// ContainsActiveSinkPipewire returns
// 0: if the given text is NOT found in the active sink
// 1: if the given text IS found in the active sink
func ContainsActiveSinkPipewire(text string) int {
	sink := FindActiveSinkPipewire(text)
	if sink == nil {
		return 0
	}

	if util.ContainsIgnoreCase(sink["node.name"], text) ||
		util.ContainsIgnoreCase(sink["node.description"], text) {
		return 1
	} else {
		return 0
	}
}

// FindSinkPipewire returns the index of a sink that contains the given text
func FindSinkPipewire(text string) map[string]string {
	objects := getPipewireObjects(
		PropertyFilter{"media.class", "Audio/Sink"},
	)
	for _, item := range objects {
		if util.ContainsIgnoreCase(item["node.description"], text) {
			return item
		}
	}

	return nil
}
