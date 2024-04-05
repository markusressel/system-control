package pipewire

import (
	"encoding/json"
	"fmt"
)

const (
	TypeNode     = "PipeWire:Interface:Node"
	TypeFactory  = "PipeWire:Interface:Factory"
	TypeModule   = "PipeWire:Interface:Module"
	TypeCore     = "PipeWire:Interface:Core"
	TypeClient   = "PipeWire:Interface:Client"
	TypeLink     = "PipeWire:Interface:Link"
	TypePort     = "PipeWire:Interface:Port"
	TypeDevice   = "PipeWire:Interface:Device"
	TypeProfiler = "PipeWire:Interface:Profiler"
	TypeMetadata = "PipeWire:Interface:Metadata"
)

type PipewireState struct {
	Objects   []PipewireStateObject
	Nodes     []PipewireInterfaceNode
	Factories []PipewireStateObject
	Modules   []PipewireStateObject
	Cores     []PipewireStateObject
	Clients   []PipewireStateObject
	Links     []PipewireStateObject
	Ports     []PipewireStateObject
	Devices   []PipewireStateObject
	Profilers []PipewireStateObject
	Metadatas []PipewireStateObject
}

type CommonData struct {
	Id          int                      `json:"id"`
	Type        string                   `json:"type"`
	Version     int                      `json:"version"`
	Permissions []string                 `json:"permissions"`
	Props       map[string]interface{}   `json:"props,omitempty"`
	Metadata    []map[string]interface{} `json:"metadata,omitempty"`
}

type PipewireGraphObject struct {
	CommonData CommonData
	Info       PipewireObjectInfo
}

type PipewireStateObject struct {
	Id          int                      `json:"id"`
	Type        string                   `json:"type"`
	Version     int                      `json:"version"`
	Permissions []string                 `json:"permissions"`
	Info        PipewireObjectInfo       `json:"info,omitempty"`
	Props       map[string]interface{}   `json:"props,omitempty"`
	Metadata    []map[string]interface{} `json:"metadata,omitempty"`
}

func (o *PipewireStateObject) UnmarshalJSON(data []byte) error {
	// Unmarshall common data
	temp := new(struct {
		Id          int                      `json:"id"`
		Type        string                   `json:"type"`
		Version     int                      `json:"version"`
		Permissions []string                 `json:"permissions"`
		Info        json.RawMessage          `json:"info,omitempty"`
		Props       map[string]interface{}   `json:"props,omitempty"`
		Metadata    []map[string]interface{} `json:"metadata,omitempty"`
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
		case TypeNode:
			info := PipewireInterfaceNodeInfo{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case TypeFactory:
			info := PipewireInterfaceFactoryInfo{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case TypeModule:
			info := PipewireInterfaceModuleInfo{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case TypeCore:
			info := PipewireInterfaceCoreInfo{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case TypeClient:
			info := PipewireInterfaceClientInfo{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case TypeLink:
			info := PipewireInterfaceLinkInfo{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case TypePort:
			info := PipewireInterfacePortInfo{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case TypeDevice:
			info := PipewireInterfaceDeviceInfo{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case TypeProfiler:
			info := PipewireInterfaceProfilerInfo{}
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

// PipewireInterfaceNodeInfo Type: "PipeWire:Interface:Node"
type PipewireInterfaceNodeInfo struct {
	MaxInputPorts  int                    `json:"max-input-ports"`
	MaxOutputPorts int                    `json:"max-output-ports"`
	ChangeMask     []string               `json:"change-mask"`
	NInputPorts    int                    `json:"n-input-ports"`
	NOutputPorts   int                    `json:"n-output-ports"`
	State          string                 `json:"state"`
	Error          string                 `json:"error"`
	Props          map[string]interface{} `json:"props"`
}

type PipewireInterfaceNode struct {
	CommonData
	Info PipewireInterfaceNodeInfo
}

func (n PipewireInterfaceNode) GetName() (string, error) {
	nodeName, ok := n.Info.Props["node.name"].(string)
	if !ok {
		return "", fmt.Errorf("could not get node name")
	}
	return nodeName, nil
}

// PipewireInterfaceFactoryInfo Type: "PipeWire:Interface:Factory"
type PipewireInterfaceFactoryInfo struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Version    int                    `json:"version"`
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
}

// PipewireInterfaceModuleInfo Type: "PipeWire:Interface:Module"
type PipewireInterfaceModuleInfo struct {
	Name       string                 `json:"name"`
	Filename   string                 `json:"filename"`
	Args       interface{}            `json:"args"`
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
}

// PipewireInterfaceCoreInfo Type: "PipeWire:Interface:Core"
type PipewireInterfaceCoreInfo struct {
	Cookie     int                    `json:"cookie"`
	UserName   string                 `json:"user-name"`
	HostName   string                 `json:"host-name"`
	Version    string                 `json:"version"`
	Name       string                 `json:"name"`
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
}

// PipewireInterfaceClientInfo Type: "PipeWire:Interface:Client"
type PipewireInterfaceClientInfo struct {
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
}

// PipewireInterfaceLinkInfo Type: "PipeWire:Interface:Link"
type PipewireInterfaceLinkInfo struct {
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

// PipewireInterfacePortInfo Type: "PipeWire:Interface:Port"
type PipewireInterfacePortInfo struct {
	Direction  string                 `json:"direction"`
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
	Params     map[string]interface{} `json:"params"`
}

// PipewireInterfaceDeviceInfo Type: "PipeWire:Interface:Device"
type PipewireInterfaceDeviceInfo struct {
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
	Params     map[string]interface{} `json:"params"`
}

// PipewireInterfaceProfilerInfo Type: "PipeWire:Interface:Profiler"
type PipewireInterfaceProfilerInfo struct {
}
