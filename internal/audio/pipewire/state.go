package pipewire

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/markusressel/system-control/internal/util"
	"log"
	"strings"
)

const (
	MediaClassAudioSink = "Audio/Sink"
)

type GraphState struct {
	Nodes            []InterfaceNode
	Factories        []InterfaceFactory
	Modules          []InterfaceModule
	Cores            []InterfaceCore
	Clients          []InterfaceClient
	Links            []InterfaceLink
	Ports            []InterfacePort
	Devices          []InterfaceDevice
	Profilers        []InterfaceProfiler
	Metadatas        []InterfaceMetadata
	SecurityContexts []InterfaceSecurityContext
}

func (state *GraphState) UnmarshalJSON(data []byte) error {

	var objectDataList []GraphObject
	if err := json.NewDecoder(strings.NewReader(string(data))).Decode(&objectDataList); err != nil {
		log.Fatalf("decode: %s", err)
	}

	for _, object := range objectDataList {
		switch object.CommonData.Type {
		case TypeNode:
			state.Nodes = append(state.Nodes, InterfaceNode{
				CommonData: object.CommonData,
				Info:       object.Info.(InterfaceNodeInfo),
			})
		case TypeFactory:
			state.Factories = append(state.Factories, InterfaceFactory{
				CommonData: object.CommonData,
				Info:       object.Info.(InterfaceFactoryInfo),
			})
		case TypeModule:
			state.Modules = append(state.Modules, InterfaceModule{
				CommonData: object.CommonData,
				Info:       object.Info.(InterfaceModuleInfo),
			})
		case TypeCore:
			state.Cores = append(state.Cores, InterfaceCore{
				CommonData: object.CommonData,
				Info:       object.Info.(InterfaceCoreInfo),
			})
		case TypeClient:
			state.Clients = append(state.Clients, InterfaceClient{
				CommonData: object.CommonData,
				Info:       object.Info.(InterfaceClientInfo),
			})
		case TypeLink:
			state.Links = append(state.Links, InterfaceLink{
				CommonData: object.CommonData,
				Info:       object.Info.(InterfaceLinkInfo),
			})
		case TypePort:
			state.Ports = append(state.Ports, InterfacePort{
				CommonData: object.CommonData,
				Info:       object.Info.(InterfacePortInfo),
			})
		case TypeDevice:
			state.Devices = append(state.Devices, InterfaceDevice{
				CommonData: object.CommonData,
				Info:       object.Info.(InterfaceDeviceInfo),
			})
		case TypeProfiler:
			var info InterfaceProfilerInfo = nil
			if object.Info != nil {
				info = object.Info.(InterfaceProfilerInfo)
			}
			state.Profilers = append(state.Profilers, InterfaceProfiler{
				CommonData: object.CommonData,
				Info:       info,
			})
		case TypeMetadata:
			var info InterfaceMetadataInfo = nil
			if object.Info != nil {
				info = object.Info.(InterfaceMetadataInfo)
			}
			state.Metadatas = append(state.Metadatas, InterfaceMetadata{
				CommonData: object.CommonData,
				Info:       info,
			})
		case TypeSecurityContext:
			var info InterfaceSecurityContextInfo = nil
			if object.Info != nil {
				info = object.Info.(InterfaceSecurityContextInfo)
			}
			state.SecurityContexts = append(state.SecurityContexts, InterfaceSecurityContext{
				CommonData: object.CommonData,
				Info:       info,
			})
		default:
			fmt.Println("Unknown type: ", object.Type)
		}
	}

	return nil
}

func (o *GraphObject) GetName() (string, error) {
	infoProps, ok := o.Info.(InterfaceNodeInfo)
	if !ok {
		return "", errors.New("invalid object type")
	}
	nodeName, ok := infoProps.Props["node.name"].(string)
	if !ok {
		return "", errors.New("node name not found")
	}
	return nodeName, nil
}

func (state *GraphState) IsMuted(sinkId int) (bool, error) {
	node, err := state.GetNodeById(sinkId)
	if err != nil {
		return false, err
	}
	muted := node.GetMuted()
	return muted, err
}

// SetMuted sets the given volume to the given sink using pipewire
// volume in percent
func (state *GraphState) SetMuted(deviceId int, muted bool) error {
	device, err := state.GetDeviceById(deviceId)
	if err != nil {
		return err
	}
	return device.SetMuted(muted)
}

// GetDefaultSinkNodeName returns the "node.name" value of the InterfaceNode that is
// currently used as the default "audio.sink".
func (state *GraphState) GetDefaultSinkNodeName() (string, error) {
	for _, item := range state.Metadatas {
		if item.Props["metadata.name"] != "default" {
			continue
		}

		for _, entry := range item.Metadata {
			if entry["key"] == "default.audio.sink" {
				return entry["value"].(map[string]interface{})["name"].(string), nil
			}
		}
	}

	return "", errors.New("default sink not found")
}

func (state *GraphState) GetDefaultSource() (string, error) {
	defaultSinkName, err := state.GetDefaultSinkNodeName()
	if err != nil {
		return "", err
	}

	node := state.FindNodesByName(defaultSinkName)
	if len(node) <= 0 {
		return "", errors.New("default sink not found")
	}
	if len(node) > 1 {
		return "", errors.New("multiple default sinks found")
	}
	return node[0].GetName()
}

func (state *GraphState) GetNodeById(id int) (InterfaceNode, error) {
	for _, node := range state.Nodes {
		if node.Id == id {
			return node, nil
		}
	}
	return InterfaceNode{}, errors.New("node not found")
}

func (state *GraphState) GetDeviceById(id int) (InterfaceDevice, error) {
	for _, device := range state.Devices {
		if device.Id == id {
			return device, nil
		}
	}
	return InterfaceDevice{}, errors.New("device not found")
}

func (state *GraphState) GetDeviceByName(name string) (InterfaceDevice, error) {
	for _, device := range state.Devices {
		infoProps := device.Info.Props
		deviceName := infoProps["device.name"].(string)
		deviceDescription := infoProps["device.description"].(string)
		if util.ContainsIgnoreCase(deviceName, name) || util.ContainsIgnoreCase(deviceDescription, name) {
			return device, nil
		}
	}
	return InterfaceDevice{}, errors.New("device not found")
}

func (state *GraphState) getClientById(id int) (InterfaceClient, error) {
	for _, client := range state.Clients {
		if client.Id == id {
			return client, nil
		}
	}
	return InterfaceClient{}, errors.New("client not found")
}

func (state *GraphState) GetLinkById(id int) (InterfaceLink, error) {
	for _, link := range state.Links {
		if link.Id == id {
			return link, nil
		}
	}
	return InterfaceLink{}, errors.New("link not found")
}

func (state *GraphState) GetPortById(id int) (InterfacePort, error) {
	for _, port := range state.Ports {
		if port.Id == id {
			return port, nil
		}
	}
	return InterfacePort{}, errors.New("port not found")
}

func (state *GraphState) GetFactoryById(id int) (InterfaceFactory, error) {
	for _, factory := range state.Factories {
		if factory.Id == id {
			return factory, nil
		}
	}
	return InterfaceFactory{}, errors.New("factory not found")
}

func (state *GraphState) GetModuleById(id int) (InterfaceModule, error) {
	for _, module := range state.Modules {
		if module.Id == id {
			return module, nil
		}
	}
	return InterfaceModule{}, errors.New("module not found")
}

func (state *GraphState) FindNodesByName(name string) []InterfaceNode {
	result := make([]InterfaceNode, 0)
	for _, node := range state.Nodes {
		nodeInfoProperties := node.Info.Props
		nodeName := nodeInfoProperties["node.name"].(string)
		nodeDescription, ok := nodeInfoProperties["node.description"].(string)
		if !ok {
			nodeDescription = ""
		}
		if util.ContainsIgnoreCase(nodeName, name) || util.ContainsIgnoreCase(nodeDescription, name) {
			result = append(result, node)
		}
	}
	return result
}

func (state *GraphState) GetPortByName(nodeName string, name string) (InterfacePort, error) {
	nodes := state.FindNodesByName(nodeName)
	if len(nodes) <= 0 {
		return InterfacePort{}, errors.New("node not found")
	}
	if len(nodes) > 1 {
		return InterfacePort{}, errors.New("ambiguous node name")
	}
	node := nodes[0]

	for _, port := range state.Ports {
		infoProps := port.Info
		if infoProps.Props["port.name"] == name && infoProps.Props["port.node"] == node.Info.Props["object.id"] {
			return port, nil
		}
	}
	return InterfacePort{}, errors.New("port not found")
}

// SetVolume sets the given volume to the given sink using pipewire
// volume in percent
func (state *GraphState) SetVolume(deviceId int, volume float64) error {
	node, err := state.GetDeviceById(deviceId)
	if err != nil {
		return err
	}
	return node.SetVolume(volume)
}

func (state *GraphState) GetNodesOfDevice(deviceId int) []InterfaceNode {
	result := make([]InterfaceNode, 0)
	for _, node := range state.Nodes {
		if node.Info.Props["device.id"] == deviceId {
			result = append(result, node)
		}
	}
	return result
}

func (state *GraphState) GetSinkNodes() []InterfaceNode {
	var result []InterfaceNode
	for _, node := range state.Nodes {
		mediaClass, ok := node.Info.Props["media.class"].(string)
		if !ok {
			continue
		}
		if mediaClass == MediaClassAudioSink {
			result = append(result, node)
		}
	}
	return result
}

func (state *GraphState) GetStreamNodes() []InterfaceNode {
	var result []InterfaceNode
	for _, node := range state.Nodes {
		mediaClass, ok := node.Info.Props["media.class"].(string)
		if !ok {
			continue
		}
		if mediaClass == "Stream/Output/Audio" {
			result = append(result, node)
		}
	}
	return result
}

func (state *GraphState) FindStreamNodes(name string) []InterfaceNode {
	result := make([]InterfaceNode, 0)
	streamNodes := state.GetStreamNodes()
	for _, node := range streamNodes {
		nodeInfoProperties := node.Info.Props
		nodeName := nodeInfoProperties["node.name"].(string)
		nodeDescription, ok := nodeInfoProperties["node.description"].(string)
		if !ok {
			nodeDescription = ""
		}
		if util.ContainsIgnoreCase(nodeName, name) || util.ContainsIgnoreCase(nodeDescription, name) {
			result = append(result, node)
		}
	}
	return result
}

// SwitchSinkTo switches the default sink to the given node and moves
// all existing streams on the currently active sink to the new default sink
func (state *GraphState) SwitchSinkTo(node InterfaceNode) error {
	nodeName, err := node.GetName()
	if err != nil {
		return err
	}

	streams := state.GetStreamNodes()

	// figure out id of target stream
	objectSerial, err := node.GetObjectSerial()
	if err != nil {
		return err
	}

	err = setDefaultSink(nodeName)
	if err != nil {
		return err
	}

	for _, stream := range streams {
		err = moveStreamToNode(stream.Id, node.Id, objectSerial)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetVolumeByName returns the volume of the with the given name.
// The name must be part of the "node.name" or "node.description" property.
// If the name is empty, the volume of the active sink is returned.
// The volume is returned as a float value in [0..1]
func (state *GraphState) GetVolumeByName(name string) (float64, error) {
	var node InterfaceNode
	if name == "" {
		activeSink, err := state.GetDefaultSinkNode()
		if err != nil {
			return -1, err
		}
		node = activeSink
	} else {
		nodes := state.FindNodesByName(name)
		if len(nodes) <= 0 {
			return -1, errors.New("node not found")
		}
		if len(nodes) > 1 {
			return -1, errors.New("ambiguous node name")
		}
		node = nodes[0]
	}
	channelVolumes := node.GetVolume()
	// use left channel for now
	return channelVolumes[0], nil
}

// GetDefaultSinkNode returns the index of the active device
func (state *GraphState) GetDefaultSinkNode() (InterfaceNode, error) {
	currentDefaultSinkName, err := util.ExecCommand("pactl", "get-default-sink")
	if err != nil {
		return InterfaceNode{}, err
	}
	nodes := state.FindNodesByName(currentDefaultSinkName)
	if len(nodes) <= 0 {
		return InterfaceNode{}, errors.New("node not found")
	}
	if len(nodes) > 1 {
		return InterfaceNode{}, errors.New("ambiguous node name")
	}
	return nodes[0], nil
}

// ContainsActiveSink returns
// 0: if the given text is NOT found in the active sink
// 1: if the given text IS found in the active sink
func (state *GraphState) ContainsActiveSink(text string) int {
	node, err := state.GetDefaultSinkNode()
	if err != nil {
		return 0
	}

	nodeName := node.Info.Props["node.name"].(string)
	nodeDescription := node.Info.Props["node.description"].(string)

	if util.ContainsIgnoreCase(nodeName, text) || util.ContainsIgnoreCase(nodeDescription, text) {
		return 1
	} else {
		return 0
	}
}

// SetDeviceProfile sets the given profile to the given device using pipewire
// profile is the name of the profile to set
// deviceId is the id of the device to set the profile for
func (state *GraphState) SetDeviceProfile(deviceId int, profile string) error {
	device, err := state.GetDeviceById(deviceId)
	if err != nil {
		return err
	}
	profileId, err := device.GetProfileIdByName(profile)
	if err != nil {
		return err
	}
	err = device.SetProfileByName(profileId.Name)
	if err != nil {
		return err
	}
	return nil
}

// GetVolume returns the volume of the active sink
// The volume is returned as a float value in [0..1]
func (state *GraphState) GetVolume() (float64, error) {
	return state.GetVolumeByName("")
}

// FindDevicesByName returns the first device that matches the given name.
func (state *GraphState) FindDevicesByName(searchTerm string) ([]InterfaceDevice, error) {
	var matches []InterfaceDevice
	for _, device := range state.Devices {
		infoProps := device.Info.Props
		deviceName := infoProps["device.name"].(string)
		deviceDescription := infoProps["device.description"].(string)
		if util.ContainsIgnoreCase(deviceName, searchTerm) || util.ContainsIgnoreCase(deviceDescription, searchTerm) {
			matches = append(matches, device)
		}
	}
	return matches, errors.New("device not found")
}

// FindDeviceByName returns the first device that matches the given name.
func (state *GraphState) FindDeviceByName(name string) (InterfaceDevice, error) {
	matches, err := state.FindDevicesByName(name)
	if err != nil {
		return InterfaceDevice{}, err
	}
	if len(matches) > 1 {
		return InterfaceDevice{}, fmt.Errorf("multiple devices found for name '%s'", name)
	}
	return matches[0], nil
}
