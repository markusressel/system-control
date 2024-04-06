package pipewire

import (
	"encoding/json"
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
