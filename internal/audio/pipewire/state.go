package pipewire

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

type GraphState struct {
	Nodes     []InterfaceNode
	Factories []InterfaceFactory
	Modules   []InterfaceModule
	Cores     []InterfaceCore
	Clients   []InterfaceClient
	Links     []InterfaceLink
	Ports     []InterfacePort
	Devices   []InterfaceDevice
	Profilers []InterfaceProfiler
	Metadatas []InterfaceMetadata
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
		default:
			fmt.Println("Unknown type: ", object.Type)
		}
	}

	return nil
}

func (state *GraphState) IsMuted(sinkId int) (bool, error) {
	node, err := state.GetNodeById(sinkId)
	if err != nil {
		return false, err
	}
	muted := node.GetMuted()
	return muted, err
}

func (state *GraphState) SetMuted(deviceId int, muted bool) error {
	device, err := state.GetDeviceById(deviceId)
	if err != nil {
		return err
	}
	return device.SetMuted(muted)
}

func (state *GraphState) GetDefaultSink() (string, error) {
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
	defaultSinkName, err := state.GetDefaultSink()
	if err != nil {
		return "", err
	}

	node, err := state.GetNodeByName(defaultSinkName)
	if err != nil {
		return "", err
	}
	return node.GetName()
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

func (state *GraphState) GetNodeByName(name string) (InterfaceNode, error) {
	for _, node := range state.Nodes {
		nodeInfoProperties := node.Info.Props
		if nodeInfoProperties["node.name"] == name {
			objectId := nodeInfoProperties["object.id"].(float64)
			deviceId := nodeInfoProperties["device.id"].(float64)
			clientId := nodeInfoProperties["client.id"].(float64)
			cardProfileDevice := nodeInfoProperties["card.profile.device"].(float64)
			deviceRoutes := nodeInfoProperties["device.routes"].(float64)
			fmt.Println("Found node: ", name, " with id: ", objectId, " device id: ", deviceId, " client id: ", clientId, " card profile device: ", cardProfileDevice, " device routes: ", deviceRoutes)
			return node, nil
		}
	}
	return InterfaceNode{}, errors.New("node not found")
}

func (state *GraphState) GetPortByName(nodeName string, name string) (InterfacePort, error) {
	node, err := state.GetNodeByName(nodeName)
	if err != nil {
		return InterfacePort{}, err
	}
	for _, port := range state.Ports {
		infoProps := port.Info
		if infoProps.Props["port.name"] == name && infoProps.Props["port.node"] == node.Info.Props["object.id"] {
			return port, nil
		}
	}
	return InterfacePort{}, errors.New("port not found")
}

func (state *GraphState) SetVolume(deviceId int, volume float64) error {
	node, err := state.GetDeviceById(deviceId)
	if err != nil {
		return err
	}
	return node.SetVolume(volume)
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
