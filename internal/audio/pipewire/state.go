package pipewire

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

type PipewireState struct {
	Nodes     []PipewireInterfaceNode
	Factories []PipewireInterfaceFactory
	Modules   []PipewireInterfaceModule
	Cores     []PipewireInterfaceCore
	Clients   []PipewireInterfaceClient
	Links     []PipewireInterfaceLink
	Ports     []PipewireInterfacePort
	Devices   []PipewireInterfaceDevice
	Profilers []PipewireInterfaceProfiler
	Metadatas []PipewireInterfaceMetadata
}

func (state *PipewireState) UnmarshalJSON(data []byte) error {

	var objectDataList []PipewireStateObject
	if err := json.NewDecoder(strings.NewReader(string(data))).Decode(&objectDataList); err != nil {
		log.Fatalf("decode: %s", err)
	}

	for _, object := range objectDataList {
		switch object.CommonData.Type {
		case TypeNode:
			state.Nodes = append(state.Nodes, PipewireInterfaceNode{
				CommonData: object.CommonData,
				Info:       object.Info.(PipewireInterfaceNodeInfo),
			})
		case TypeFactory:
			state.Factories = append(state.Factories, PipewireInterfaceFactory{
				CommonData: object.CommonData,
				Info:       object.Info.(PipewireInterfaceFactoryInfo),
			})
		case TypeModule:
			state.Modules = append(state.Modules, PipewireInterfaceModule{
				CommonData: object.CommonData,
				Info:       object.Info.(PipewireInterfaceModuleInfo),
			})
		case TypeCore:
			state.Cores = append(state.Cores, PipewireInterfaceCore{
				CommonData: object.CommonData,
				Info:       object.Info.(PipewireInterfaceCoreInfo),
			})
		case TypeClient:
			state.Clients = append(state.Clients, PipewireInterfaceClient{
				CommonData: object.CommonData,
				Info:       object.Info.(PipewireInterfaceClientInfo),
			})
		case TypeLink:
			state.Links = append(state.Links, PipewireInterfaceLink{
				CommonData: object.CommonData,
				Info:       object.Info.(PipewireInterfaceLinkInfo),
			})
		case TypePort:
			state.Ports = append(state.Ports, PipewireInterfacePort{
				CommonData: object.CommonData,
				Info:       object.Info.(PipewireInterfacePortInfo),
			})
		case TypeDevice:
			state.Devices = append(state.Devices, PipewireInterfaceDevice{
				CommonData: object.CommonData,
				Info:       object.Info.(PipewireInterfaceDeviceInfo),
			})
		case TypeProfiler:
			var info PipewireInterfaceProfilerInfo = nil
			if object.Info != nil {
				info = object.Info.(PipewireInterfaceProfilerInfo)
			}
			state.Profilers = append(state.Profilers, PipewireInterfaceProfiler{
				CommonData: object.CommonData,
				Info:       info,
			})
		case TypeMetadata:
			var info PipewireInterfaceMetadataInfo = nil
			if object.Info != nil {
				info = object.Info.(PipewireInterfaceMetadataInfo)
			}
			state.Metadatas = append(state.Metadatas, PipewireInterfaceMetadata{
				CommonData: object.CommonData,
				Info:       info,
			})
		default:
			fmt.Println("Unknown type: ", object.Type)
		}
	}

	return nil
}

func (state PipewireState) IsMuted(sinkId int) (bool, error) {
	node, err := state.GetNodeById(sinkId)
	if err != nil {
		return false, err
	}
	muted := node.GetMuted()
	return muted, err
}

func (state PipewireState) SetMuted(deviceId int, muted bool) error {
	device, err := state.GetDeviceById(deviceId)
	if err != nil {
		return err
	}
	return device.SetMuted(muted)
}

func (state PipewireState) GetDefaultSink() (string, error) {
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

func (state PipewireState) GetDefaultSource() (string, error) {
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

func (state PipewireState) GetNodeById(id int) (PipewireInterfaceNode, error) {
	for _, node := range state.Nodes {
		if node.Id == id {
			return node, nil
		}
	}
	return PipewireInterfaceNode{}, errors.New("node not found")
}

func (state PipewireState) GetDeviceById(id int) (PipewireInterfaceDevice, error) {
	for _, device := range state.Devices {
		if device.Id == id {
			return device, nil
		}
	}
	return PipewireInterfaceDevice{}, errors.New("device not found")
}

func (state PipewireState) getClientById(id int) (PipewireInterfaceClient, error) {
	for _, client := range state.Clients {
		if client.Id == id {
			return client, nil
		}
	}
	return PipewireInterfaceClient{}, errors.New("client not found")
}

func (state PipewireState) GetLinkById(id int) (PipewireInterfaceLink, error) {
	for _, link := range state.Links {
		if link.Id == id {
			return link, nil
		}
	}
	return PipewireInterfaceLink{}, errors.New("link not found")
}

func (state PipewireState) GetPortById(id int) (PipewireInterfacePort, error) {
	for _, port := range state.Ports {
		if port.Id == id {
			return port, nil
		}
	}
	return PipewireInterfacePort{}, errors.New("port not found")
}

func (state PipewireState) GetFactoryById(id int) (PipewireInterfaceFactory, error) {
	for _, factory := range state.Factories {
		if factory.Id == id {
			return factory, nil
		}
	}
	return PipewireInterfaceFactory{}, errors.New("factory not found")
}

func (state PipewireState) GetModuleById(id int) (PipewireInterfaceModule, error) {
	for _, module := range state.Modules {
		if module.Id == id {
			return module, nil
		}
	}
	return PipewireInterfaceModule{}, errors.New("module not found")
}

func (state PipewireState) GetNodeByName(name string) (PipewireInterfaceNode, error) {
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
	return PipewireInterfaceNode{}, errors.New("node not found")
}

func (state PipewireState) GetPortByName(nodeName string, name string) (PipewireInterfacePort, error) {
	node, err := state.GetNodeByName(nodeName)
	if err != nil {
		return PipewireInterfacePort{}, err
	}
	for _, port := range state.Ports {
		infoProps := port.Info
		if infoProps.Props["port.name"] == name && infoProps.Props["port.node"] == node.Info.Props["object.id"] {
			return port, nil
		}
	}
	return PipewireInterfacePort{}, errors.New("port not found")
}

func (o *PipewireStateObject) GetName() (string, error) {
	infoProps, ok := o.Info.(PipewireInterfaceNodeInfo)
	if !ok {
		return "", errors.New("invalid object type")
	}
	nodeName, ok := infoProps.Props["node.name"].(string)
	if !ok {
		return "", errors.New("node name not found")
	}
	return nodeName, nil
}
