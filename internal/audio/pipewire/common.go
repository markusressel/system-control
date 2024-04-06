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

type CommonData struct {
	Id          int                      `json:"id"`
	Type        string                   `json:"type"`
	Version     int                      `json:"version"`
	Permissions []string                 `json:"permissions"`
	Props       map[string]interface{}   `json:"props,omitempty"`
	Metadata    []map[string]interface{} `json:"metadata,omitempty"`
}

type GraphObject struct {
	CommonData
	Info GraphObjectInfo `json:"info,omitempty"`
}

func (o *GraphObject) UnmarshalJSON(data []byte) error {
	// Unmarshall common data
	temp := new(struct {
		CommonData
		Info json.RawMessage `json:"info,omitempty"`
	})
	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}

	o.CommonData = temp.CommonData

	if temp.Info != nil {
		switch temp.Type {
		case TypeNode:
			info := InterfaceNodeInfo{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case TypeFactory:
			info := InterfaceFactoryInfo{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case TypeModule:
			info := InterfaceModuleInfo{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case TypeCore:
			info := InterfaceCoreInfo{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case TypeClient:
			info := InterfaceClientInfo{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case TypeLink:
			info := InterfaceLinkInfo{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case TypePort:
			info := InterfacePortInfo{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case TypeDevice:
			info := InterfaceDeviceInfo{}
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case TypeProfiler:
			info := InterfaceProfilerInfo(&map[string]interface{}{})
			err := json.Unmarshal(temp.Info, &info)
			if err != nil {
				return err
			}
			o.Info = info
		case TypeMetadata:
			info := InterfaceMetadataInfo{}
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

type GraphObjectInfo interface{}

// InterfaceNodeInfo Type: "PipeWire:Interface:Node"
type InterfaceNodeInfo struct {
	MaxInputPorts  int                    `json:"max-input-ports"`
	MaxOutputPorts int                    `json:"max-output-ports"`
	ChangeMask     []string               `json:"change-mask"`
	NInputPorts    int                    `json:"n-input-ports"`
	NOutputPorts   int                    `json:"n-output-ports"`
	State          string                 `json:"state"`
	Error          string                 `json:"error"`
	Props          map[string]interface{} `json:"props"`
	Params         map[string]interface{} `json:"params"`
}

type InterfaceFactory struct {
	CommonData
	Info InterfaceFactoryInfo
}

type InterfaceModule struct {
	CommonData
	Info InterfaceModuleInfo
}

type InterfaceCore struct {
	CommonData
	Info InterfaceCoreInfo
}

type InterfaceClient struct {
	CommonData
	Info InterfaceClientInfo
}

type InterfaceLink struct {
	CommonData
	Info InterfaceLinkInfo
}

type InterfacePort struct {
	CommonData
	Info InterfacePortInfo
}

type InterfaceProfiler struct {
	CommonData
	Info InterfaceProfilerInfo
}

type InterfaceMetadata struct {
	CommonData
	Info map[string]interface{}
}

func (n InterfaceNode) GetName() (string, error) {
	nodeName, ok := n.Info.Props["node.name"].(string)
	if !ok {
		return "", fmt.Errorf("could not get node name")
	}
	return nodeName, nil
}

// InterfaceFactoryInfo Type: "PipeWire:Interface:Factory"
type InterfaceFactoryInfo struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Version    int                    `json:"version"`
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
}

// InterfaceModuleInfo Type: "PipeWire:Interface:Module"
type InterfaceModuleInfo struct {
	Name       string                 `json:"name"`
	Filename   string                 `json:"filename"`
	Args       interface{}            `json:"args"`
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
}

// InterfaceCoreInfo Type: "PipeWire:Interface:Core"
type InterfaceCoreInfo struct {
	Cookie     int                    `json:"cookie"`
	UserName   string                 `json:"user-name"`
	HostName   string                 `json:"host-name"`
	Version    string                 `json:"version"`
	Name       string                 `json:"name"`
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
}

// InterfaceClientInfo Type: "PipeWire:Interface:Client"
type InterfaceClientInfo struct {
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
}

// InterfaceLinkInfo Type: "PipeWire:Interface:Link"
type InterfaceLinkInfo struct {
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

// InterfacePortInfo Type: "PipeWire:Interface:Port"
type InterfacePortInfo struct {
	Direction  string                 `json:"direction"`
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
	Params     map[string]interface{} `json:"params"`
}

// InterfaceProfilerInfo Type: "PipeWire:Interface:Profiler"
type InterfaceProfilerInfo *map[string]interface{}

type InterfaceMetadataInfo map[string]interface{}
