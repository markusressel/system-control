package pipewire

import (
	"github.com/markusressel/system-control/internal/util"
	"strconv"
	"strings"
)

type PropertyFilter struct {
	key   string
	value string
}

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

func parsePipewireObjectProperties(lines []string, endIndentation int) ObjectProperties {
	result := make(ObjectProperties)

	//result = parsePipewireObjectPropertyValue(lines, endIndentation)

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
		} else if key == "Object:" {
			_value, subConsumedLines := parsePipewireObject(lines[consumedLines:len(lines)-1], propertyIndentation)
			value = _value
			consumedLines = consumedLines + subConsumedLines
			break
		} else {
			// log.Printf("Ignored line: %s", line)
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
		subValue, subConsumedLines := parsePipewireObjectPropertyValue(lines[consumedLines:len(lines)-1], endIndentation)
		consumedLines = consumedLines + subConsumedLines

		value = append(value, subValue)
	}

	return value, consumedLines
}

func parsePipewireObjectPropertyValueStruct(lines []string, endIndentation int) (value map[string]interface{}, consumedLines int) {
	// TODO:
	return map[string]interface{}{}, 0
}

func parsePipewireObject(lines []string, endIndentation int) (value PipewireObject, consumedLines int) {
	i := 0
	for i < len(lines) && util.CountLeadingSpace(lines[i]) > endIndentation {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		objectParams := getPairsFromLine(strings.TrimPrefix(trimmedLine, "Object:"))
		objectSize, _ := strconv.Atoi(objectParams["size"])

		// properties
		objectIndentation := util.CountLeadingSpace(line)
		objectProperties := parsePipewireObjectProperties(lines[i+1:], objectIndentation)

		// construct object
		value = PipewireObject{
			Id:         objectParams["id"],
			Type:       objectParams["type"],
			Size:       objectSize,
			Properties: objectProperties,
		}

	}
	return value, consumedLines
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
