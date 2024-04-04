package audio

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/markusressel/system-control/internal/util"
	"log"
	"math"
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

// GetVolumePipewireByName returns the volume of the with the given name.
// The name must be part of the "node.name" or "node.description" property.
// If the name is empty, the volume of the active sink is returned.
// The volume is returned as a float value in [0..1]
func GetVolumePipewireByName(name string) (float64, error) {
	var sinkId int
	if name == "" {
		activeSink := GetActiveSinkPipewire()
		activeSinkId, err := strconv.Atoi(activeSink["id"])
		if err != nil {
			return -1, err
		}
		sinkId = activeSinkId
	} else {
		sink := FindSinkPipewire(name)
		if sink == nil {
			return -1, errors.New("Sink not found")
		}
		targetSinkId, err := strconv.Atoi(sink["id"])
		if err != nil {
			return -1, err
		}
		sinkId = targetSinkId
	}

	return GetVolumePipewireBySink(sinkId)
}

// GetVolumePipewireBySink returns the volume of the sink with the given sinkId
// The volume is returned as a float value in [0..1]
func GetVolumePipewireBySink(sinkId int) (float64, error) {
	nodeDetails, err := getNodeParams(sinkId)
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
	volume = math.Cbrt(volume)
	return volume, nil
}

// GetVolumePipewire returns the volume of the active sink
// The volume is returned as a float value in [0..1]
func GetVolumePipewire() (float64, error) {
	return GetVolumePipewireByName("")
}

func findParamProperty(details []PipewireObject, s string) (PipewireProperty, error) {
	for _, detail := range details {
		property, err := detail.findParamProperty1(s)
		if err != nil {
			return PipewireProperty{}, err
		}
		if property != nil {
			return *property, nil
		}
	}

	return PipewireProperty{}, errors.New("Unable to find property ")
}

func (p *PipewireObject) findParamProperty1(s string) (*PipewireProperty, error) {
	return p.GetProperty(s)
}

// SetVolumePipewire sets the given volume to the given sink using pipewire
// volume in percent
func SetVolumePipewire(deviceId int, volume float64) error {
	routes, err := getNodeRoutes(deviceId)
	if err != nil {
		return err
	}

	// TODO: find default route of this device, since it might not be the first one
	// TODO: which route is the correct one?

	for _, route := range routes {
		currentRoute := route

		indexProperty, err := currentRoute.findParamProperty1("index")
		if err != nil {
			continue
		}

		deviceProperty, err := currentRoute.findParamProperty1("device")
		if err != nil {
			continue
		}

		//objects := getPipewireObjects(
		//	PropertyFilter{"media.class", "Audio/Device"},
		//	PropertyFilter{"id", strconv.Itoa(deviceId)},
		//)

		// index and deviceId are properties of the route
		// index 2 is "analog-output-speaker" on M16
		// device 7 is a property of the route with index 2 on M16
		routeIndex := indexProperty.Value
		cardProfileDevice := deviceProperty.Value

		if volume < 0 {
			volume = 0
		} else if volume > 1 {
			volume = 1
		}
		volumeCubicRoot := math.Pow(volume, 3)

		muted := false
		save := true
		_, err = util.ExecCommand(
			"pw-cli",
			"set-param",
			strconv.Itoa(deviceId),
			"Route",
			fmt.Sprintf("{ index: %d, device: %d, props: { mute: %s, channelVolumes: [ %f, %f ] }, save: %s }",
				routeIndex,
				cardProfileDevice,
				strconv.FormatBool(muted),
				volumeCubicRoot,
				volumeCubicRoot,
				strconv.FormatBool(save),
			),
		)
		if err != nil {
			continue
		}
	}

	return err
}

// SetMutedPipewire sets the given volume to the given sink using pipewire
// volume in percent
func SetMutedPipewire(deviceId int, muted bool) error {
	routes, err := getNodeRoutes(deviceId)
	if err != nil {
		return err
	}

	// TODO: find default route of this device, since it might not be the first one
	// TODO: which route is the correct one?

	for _, route := range routes {
		currentRoute := route

		indexProperty, err := currentRoute.findParamProperty1("index")
		if err != nil {
			continue
		}

		deviceProperty, err := currentRoute.findParamProperty1("device")
		if err != nil {
			continue
		}

		//objects := getPipewireObjects(
		//	PropertyFilter{"media.class", "Audio/Device"},
		//	PropertyFilter{"id", strconv.Itoa(deviceId)},
		//)

		// index and deviceId are properties of the route
		// index 2 is "analog-output-speaker" on M16
		// device 7 is a property of the route with index 2 on M16
		routeIndex := indexProperty.Value
		cardProfileDevice := deviceProperty.Value

		save := true
		_, err = util.ExecCommand(
			"pw-cli",
			"set-param",
			strconv.Itoa(deviceId),
			"Route",
			fmt.Sprintf("{ index: %d, device: %d, props: { mute: %s, save: %s }",
				routeIndex,
				cardProfileDevice,
				strconv.FormatBool(muted),
				strconv.FormatBool(save),
			),
		)
		if err != nil {
			continue
		}
	}

	return err
}

// SetVolumePulseAudio sets the given volume to the given sink using PulseAudio
// volume in percent
func SetVolumePulseAudio(sinkId int, volume float64) error {
	_, err := util.ExecCommand(
		"pactl",
		"set-sink-volume",
		strconv.Itoa(sinkId),
		fmt.Sprint(volume),
	)
	return err
}

func GetSinkByName(name string) map[string]string {
	objects := getPipewireObjects(
		PropertyFilter{"media.class", "Audio/Sink"},
	)
	for _, item := range objects {
		if util.ContainsIgnoreCase(item["node.name"], name) || util.ContainsIgnoreCase(item["node.description"], name) {
			return item
		}
	}

	return nil
}

type Sink struct {
	properties map[string]string
}

// GetActiveSinkPipewire returns the index of the active sink
func GetActiveSinkPipewire() map[string]string {
	currentDefaultSinkName, err := util.ExecCommand("pactl", "get-default-sink")
	if err != nil {
		log.Fatal(err)
	}
	return GetSinkByName(currentDefaultSinkName)
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

// FindSinkPipewire returns the sink that contains the given text
func FindSinkPipewire(text string) map[string]string {
	objects := getPipewireObjects(
		PropertyFilter{"media.class", "Audio/Sink"},
	)
	for _, item := range objects {
		if util.ContainsIgnoreCase(item["node.name"], text) || util.ContainsIgnoreCase(item["node.description"], text) {
			return item
		}
	}

	return nil
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

			//PipewireObject{
			//	Properties: ObjectProperties{
			//
			//	},
			//}

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

func getNodeRoutes(nodeId int) ([]PipewireObject, error) {
	result, err := util.ExecCommand("pw-cli", "enum-params", strconv.Itoa(nodeId), "Route")
	if err != nil {
		return nil, err
	}

	// TODO: on my laptop there should be two routes

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

func (p *PipewireObject) GetProperty(name string) (*PipewireProperty, error) {
	for key, property := range p.Properties {
		keyParts := strings.Split(key, ":")
		if util.EqualsIgnoreCase(keyParts[len(keyParts)-1], name) {
			return &property, nil
		}
	}
	return nil, errors.New("Property not found")
}

type ObjectProperties map[string]PipewireProperty

func parsePipwireParamsToMap(input string) []PipewireObject {
	lines := strings.Split(input, "\n")
	result, _ := parsePipewireObjects(lines, -1)
	return result
}

func parsePipewireObjects(lines []string, endIndentation int) ([]PipewireObject, int) {
	var err error
	var consumedLines int
	result := make([]PipewireObject, 0, 1000)

	// TODO: this currently also seems to parse nested objects, which is
	//  not entirely wrong, but they nested objects should not be part of the result here,
	//  but property values of the parent object instead

	i := 0
	for i < len(lines) && util.CountLeadingSpace(lines[i]) > endIndentation {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "Object:") {
			objectParams := getPairsFromLine(strings.TrimPrefix(trimmedLine, "Object:"))
			objectSize, _ := strconv.Atoi(objectParams["size"])

			// properties
			objectIndentation := util.CountLeadingSpace(line)
			objectProperties := parsePipewireObjectProperties(lines[i+1:], objectIndentation)

			// construct object
			object := PipewireObject{
				Id:         objectParams["id"],
				Type:       objectParams["type"],
				Size:       objectSize,
				Properties: objectProperties,
			}

			consumedLines = i
			result = append(result, object)
		}

		i++
	}

	if err != nil {
		panic(err)
	}

	return result, consumedLines
}

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

func IsMutedPipewire(sinkId int) bool {
	nodeDetails, err := getNodeParams(sinkId)
	if err != nil {
		return false
	}
	property, err := findParamProperty(nodeDetails, "mute")
	if err != nil {
		return false
	}
	return property.Value.(bool)
}

func PwDump() PipewireState {
	result, err := util.ExecCommand("pw-dump")
	if err != nil {
		log.Fatal(err)
	}

	var objectDataList []PipewireObject1
	if err := json.NewDecoder(strings.NewReader(result)).Decode(&objectDataList); err != nil {
		log.Fatalf("decode: %s", err)
	}

	state := PipewireState{
		Objects: objectDataList,
	}

	return state
}

type PipewireState struct {
	Objects []PipewireObject1
}

type PipewireObject1 struct {
	Id          int                    `json:"id"`
	Type        string                 `json:"type"`
	Version     int                    `json:"version"`
	Permissions []string               `json:"permissions"`
	Info        PipewireObjectInfo     `json:"info,omitempty"`
	Props       map[string]interface{} `json:"props,omitempty"`
	Metadata    []interface{}          `json:"metadata,omitempty"`
}

func (o *PipewireObject1) UnmarshalJSON(data []byte) error {
	// Unmarshall common data
	temp := new(struct {
		Id          int                    `json:"id"`
		Type        string                 `json:"type"`
		Version     int                    `json:"version"`
		Permissions []string               `json:"permissions"`
		Info        json.RawMessage        `json:"info,omitempty"`
		Props       map[string]interface{} `json:"props,omitempty"`
		Metadata    []interface{}          `json:"metadata,omitempty"`
	})
	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}

	o.Id = temp.Id
	o.Type = temp.Type
	o.Version = temp.Version
	o.Permissions = temp.Permissions
	o.Props = temp.Props
	o.Metadata = temp.Metadata

	if temp.Info != nil {
		switch temp.Type {
		case "PipeWire:Interface:Node":
			info := PipewireInterfaceNode{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case "PipeWire:Interface:Factory":
			info := PipewireInterfaceFactory{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case "PipeWire:Interface:Module":
			info := PipewireInterfaceModule{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case "PipeWire:Interface:Core":
			info := PipewireInterfaceCore{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case "PipeWire:Interface:Client":
			info := PipewireInterfaceClient{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case "PipeWire:Interface:Link":
			info := PipewireInterfaceLink{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case "PipeWire:Interface:Port":
			info := PipewireInterfacePort{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case "PipeWire:Interface:Device":
			info := PipewireInterfaceDevice{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case "PipeWire:Interface:Profiler":
			info := PipewireInterfaceProfiler{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		default:
			fmt.Println("Unknown type: ", temp.Type)
			o.Info = nil
		}
	}

	return nil
}

type PipewireObjectInfo interface{}

// PipewireInterfaceNode Type: "PipeWire:Interface:Node"
type PipewireInterfaceNode struct {
	MaxInputPorts  int                    `json:"max-input-ports"`
	MaxOutputPorts int                    `json:"max-output-ports"`
	ChangeMask     []string               `json:"change-mask"`
	NInputPorts    int                    `json:"n-input-ports"`
	NOutputPorts   int                    `json:"n-output-ports"`
	State          string                 `json:"state"`
	Error          string                 `json:"error"`
	Props          map[string]interface{} `json:"props"`
}

// PipewireInterfaceFactory Type: "PipeWire:Interface:Factory"
type PipewireInterfaceFactory struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Version    int                    `json:"version"`
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
}

// PipewireInterfaceModule Type: "PipeWire:Interface:Module"
type PipewireInterfaceModule struct {
	Name       string                 `json:"name"`
	Filename   string                 `json:"filename"`
	Args       interface{}            `json:"args"`
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
}

// PipewireInterfaceCore Type: "PipeWire:Interface:Core"
type PipewireInterfaceCore struct {
	Cookie     int                    `json:"cookie"`
	UserName   string                 `json:"user-name"`
	HostName   string                 `json:"host-name"`
	Version    string                 `json:"version"`
	Name       string                 `json:"name"`
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
}

// PipewireInterfaceClient Type: "PipeWire:Interface:Client"
type PipewireInterfaceClient struct {
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
}

// PipewireInterfaceLink Type: "PipeWire:Interface:Link"
type PipewireInterfaceLink struct {
	OutputNodeId int         `json:"output-node-id"`
	OutputPortId int         `json:"output-port-id"`
	InputNodeId  int         `json:"input-node-id"`
	InputPortId  int         `json:"input-port-id"`
	ChangeMask   []string    `json:"change-mask"`
	State        string      `json:"state"`
	Error        interface{} `json:"error"`
	Format       struct {
		MediaType    string `json:"mediaType"`
		MediaSubtype string `json:"mediaSubtype"`
		Format       string `json:"format"`
	} `json:"format"`
	Props map[string]interface{} `json:"props"`
}

// PipewireInterfacePort Type: "PipeWire:Interface:Port"
type PipewireInterfacePort struct {
	Direction  string                 `json:"direction"`
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
	Params     map[string]interface{} `json:"params"`
}

// PipewireInterfaceDevice Type: "PipeWire:Interface:Device"
type PipewireInterfaceDevice struct {
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
	Params     map[string]interface{} `json:"params"`
}

// PipewireInterfaceProfiler Type: "PipeWire:Interface:Profiler"
type PipewireInterfaceProfiler struct {
}
