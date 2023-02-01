package audio

import (
	"errors"
	"fmt"
	"github.com/markusressel/system-control/internal/util"
	"log"
	"strconv"
	"strings"
)

type PropertyFilter struct {
	key   string
	value string
}

// RotateActiveSinkPipewire switches the default sink and moves all existing sink inputs to the next available sink in the list
func RotateActiveSinkPipewire(reverse bool) {
	activeSink := GetActiveSinkPipewire()

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

func GetVolumePipewire() (float64, error) {
	activeSink := GetActiveSinkPipewire()

	activeSinkId, err := strconv.Atoi(activeSink["id"])
	if err != nil {
		return -1, err
	}
	nodeDetails, err := getNodeParams(activeSinkId)
	if err != nil {
		return -1, err
	}
	property, err := findParamProperty(nodeDetails, "channelVolumes")
	if err != nil {
		return -1, err
	}

	var volume float64

	value := property.Value
	switch value.(type) {
	case []interface{}:
		value = value.([]interface{})[0]
	}

	switch value.(type) {
	case int:
		volume = float64(value.(int))
	case int32, int64:
		volume = value.(float64)
	case float64, float32:
		typedValue := value.(float64)
		volume = typedValue
	default:
		volume, err = strconv.ParseFloat(fmt.Sprint(value), 64)
	}
	if err != nil {
		return -1, err
	}
	return volume, nil
}

func findParamProperty(details []PipewireObject, s string) (PipewireProperty, error) {
	for _, detail := range details {
		for key, property := range detail.Properties {
			if util.ContainsIgnoreCase(key, s) {
				return property, nil
			}
		}
	}

	return PipewireProperty{}, errors.New("Unable to find property ")
}

// SetVolumePipewire sets the given volume to the given sink using pipewire
// volume in percent
func SetVolumePipewire(sinkId int, volume float64) error {
	//objects := getPipewireObjects(
	//	PropertyFilter{"media.class", "Audio/Sink"},
	//)

	_, err := util.ExecCommand(
		"pactl",
		"set-sink-volume",
		strconv.Itoa(sinkId),
		fmt.Sprint(volume),
	)
	return err
}

// GetActiveSinkPipewire returns the index of the active sink
func GetActiveSinkPipewire() map[string]string {
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
	sink := GetActiveSinkPipewire()
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

func SetMutedPipewire(sinkId int, channel string, muted bool) error {
	var targetState int
	if muted {
		targetState = 1
	} else {
		targetState = 0
	}
	_, err := util.ExecCommand("pactl", "set-sink-mute", strconv.Itoa(sinkId), strconv.Itoa(targetState))
	return err
}

// Switches the default sink to the target sink
// You need to get a sink name with "pw-cli ls Node"
// and look for the "node.name" property for a valid value.
func setDefaultSinkPipewire(sinkName string) (err error) {
	_, err = util.ExecCommand("pw-metadata", "0", "default.configured.audio.sink", `{ "name": "`+sinkName+`" }`)
	return err
}

// retrieve a list of pipewire objects
// optionally filtered
func getPipewireObjects(filters ...PropertyFilter) (objects []map[string]string) {
	result, err := util.ExecCommand("pw-cli", "list-objects")
	if err != nil {
		log.Fatal(err)
	}

	objects = parsePipwireObjectsToMap(result)
	objects = filterPipwireObjects(objects, func(v map[string]string) bool {
		for _, filter := range filters {
			if v[filter.key] != filter.value {
				return false
			}
		}

		return true
	})

	return objects
}

func parsePipwireObjectsToMap(input string) []map[string]string {
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

// filterPipwireObjects filters the given pipewire object map based on the given function
func filterPipwireObjects(vs []map[string]string, f func(map[string]string) bool) []map[string]string {
	vsf := make([]map[string]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func getNodeParams(nodeId int) ([]PipewireObject, error) {
	result, err := util.ExecCommand("pw-cli", "enum-params", strconv.Itoa(nodeId), "Props")
	if err != nil {
		return nil, err
	}
	params := parsePipwireParamsToMap(result)
	return params, nil
}

type PipewireObject struct {
	Size       int
	Type       string
	Id         string
	Properties ObjectProperties
}

type PipewireProperty struct {
	Key   string
	Flags string
	Value interface{}
}

type ObjectProperties map[string]PipewireProperty

func parsePipwireParamsToMap(input string) []PipewireObject {
	result := make([]PipewireObject, 0, 1000)

	lines := strings.Split(input, "\n")

	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "Object:") {
			objectParams := getPairsFromLine(strings.TrimPrefix(trimmedLine, "Object:"))
			objectSize, _ := strconv.Atoi(objectParams["size"])

			// properties
			objectIndentation := util.CountLeadingSpace(line)
			objectProperties := parsePipewireObjectProperties(lines[i+1:], objectIndentation)

			// construct object
			pipewireObject := PipewireObject{
				Id:         objectParams["id"],
				Type:       objectParams["type"],
				Size:       objectSize,
				Properties: objectProperties,
			}
			result = append(result, pipewireObject)
		}

		i++
	}

	return result
}

func parsePipewireObjectProperties(lines []string, endIndentation int) ObjectProperties {
	result := make(ObjectProperties)

	i := 0
	for i < len(lines) && util.CountLeadingSpace(lines[i]) > endIndentation {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "Prop:") {
			propertyIndentation := util.CountLeadingSpace(line)

			propertyParams := getPairsFromLine(strings.TrimPrefix(trimmedLine, "Prop:"))

			// has a key, and properties in the following lines
			propertyKey := propertyParams["key"]
			propertyFlags := propertyParams["flags"]
			propertyValue, consumedLines := parsePipewireObjectPropertyValue(lines[i+1:], propertyIndentation)
			i = i + consumedLines

			property := PipewireProperty{
				Key:   propertyKey,
				Flags: propertyFlags,
				Value: propertyValue,
			}

			result[propertyKey] = property
		}

		i++
	}

	return result
}

func parsePipewireObjectPropertyValue(lines []string, endIndentation int) (value interface{}, consumedLines int) {
	consumedLines = 0
	var err error
	for consumedLines < len(lines) && util.CountLeadingSpace(lines[consumedLines]) > endIndentation {
		line := lines[consumedLines]
		trimmedLine := strings.TrimSpace(line)
		propertyIndentation := util.CountLeadingSpace(line)

		keyValue := strings.SplitN(trimmedLine, " ", 2)
		key, rawValue := strings.TrimSpace(keyValue[0]), strings.TrimSpace(keyValue[1])

		if key == "Bool" {
			value, err = strconv.ParseBool(rawValue)
			consumedLines = 1
			break
		} else if key == "Int" {
			value, err = strconv.ParseInt(rawValue, 10, 32)
			consumedLines = 1
			break
		} else if key == "Long" {
			value, err = strconv.ParseInt(rawValue, 10, 64)
			consumedLines = 1
			break
		} else if key == "Float" {
			value, err = strconv.ParseFloat(rawValue, 64)
			consumedLines = 1
			break
		} else if key == "String" {
			value = rawValue[1 : len(rawValue)-1]
			consumedLines = 1
			break
		} else if key == "Array:" {
			// TODO:
			_value, subConsumedLines := parsePipewireObjectPropertyValueArray(lines[consumedLines+1:len(lines)-1], propertyIndentation)
			value = _value
			consumedLines = consumedLines + subConsumedLines
			break
		} else if key == "Struct:" {
			// TODO:
			_value, subConsumedLines := parsePipewireObjectPropertyValueStruct(lines[consumedLines+1:len(lines)-1], propertyIndentation)
			value = _value
			consumedLines = consumedLines + subConsumedLines
			break
		} else {
			log.Printf("Ignored line: %s", line)
			consumedLines++
		}
	}

	if err != nil {
		panic(err)
	}

	return value, consumedLines
}

func parsePipewireObjectPropertyValueArray(lines []string, endIndentation int) (value []interface{}, consumedLines int) {
	consumedLines = 0
	for consumedLines < len(lines) && util.CountLeadingSpace(lines[consumedLines]) > endIndentation {
		//line := lines[consumedLines]
		//trimmedLine := strings.TrimSpace(line)
		//
		//getPairsFromLine(trimmedLine)

		subValue, subConsumedLines := parsePipewireObjectPropertyValue(lines[consumedLines+1:len(lines)-1], endIndentation)
		consumedLines = consumedLines + subConsumedLines

		value = append(value, subValue)
		consumedLines++
	}

	return value, consumedLines
}

func parsePipewireObjectPropertyValueStruct(lines []string, endIndentation int) (value map[string]interface{}, consumedLines int) {
	// TODO:
	return map[string]interface{}{}, 0
}

func getPairsFromLine(line string) map[string]string {
	result := make(map[string]string)
	objectParams := strings.Split(line, ",")
	for _, item := range objectParams {
		item = strings.TrimSpace(item)
		splits := strings.SplitAfter(item, " ")
		result[strings.TrimSpace(splits[0])] = strings.TrimSpace(splits[1])
	}
	return result
}
