package pipewire

import (
	"encoding/json"
	"fmt"
	"github.com/markusressel/system-control/internal/util"
	"log"
	"strings"
)

func PwDump() PipewireState {
	result, err := util.ExecCommand("pw-dump")
	if err != nil {
		log.Fatal(err)
	}

	var objectDataList []PipewireStateObject
	if err := json.NewDecoder(strings.NewReader(result)).Decode(&objectDataList); err != nil {
		log.Fatalf("decode: %s", err)
	}

	nodes := extractNodes(objectDataList)

	state := PipewireState{
		Objects:   objectDataList,
		Nodes:     nodes,
		Factories: filterByType(objectDataList, TypeFactory),
		Modules:   filterByType(objectDataList, TypeModule),
		Cores:     filterByType(objectDataList, TypeCore),
		Clients:   filterByType(objectDataList, TypeClient),
		Links:     filterByType(objectDataList, TypeLink),
		Ports:     filterByType(objectDataList, TypePort),
		Devices:   filterByType(objectDataList, TypeDevice),
		Profilers: filterByType(objectDataList, TypeProfiler),
		Metadatas: filterByType(objectDataList, TypeMetadata),
	}

	defaultSinkName, err := state.GetDefaultSink()
	fmt.Println("Default sink: ", defaultSinkName)

	defaultSourceName, err := state.GetDefaultSource()
	fmt.Println("Default source: ", defaultSourceName)

	//port, err := state.GetPortByType("PipeWire:Interface:Port", "Audio/Source")

	return state
}

func extractNodes(list []PipewireStateObject) []PipewireInterfaceNode {
	var nodes []PipewireInterfaceNode
	for _, object := range filterByType(list, TypeNode) {
		node := PipewireInterfaceNode{
			CommonData: CommonData{
				Id:          object.Id,
				Type:        object.Type,
				Version:     object.Version,
				Permissions: object.Permissions,
				Props:       object.Props,
				Metadata:    object.Metadata,
			},
			Info: object.Info.(PipewireInterfaceNodeInfo),
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func filterByType(list []PipewireStateObject, t string) []PipewireStateObject {
	result := []PipewireStateObject{}
	for _, item := range list {
		if item.Type == t {
			result = append(result, item)
		}
	}
	return result
}
