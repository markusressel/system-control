package pipewire

import "math"

type InterfaceNode struct {
	CommonData
	Info InterfaceNodeInfo
}

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

func (n InterfaceNode) GetParentDevice() (InterfaceDevice, error) {
	state := PwDump()
	deviceId := n.Info.GetDeviceID()
	return state.GetDeviceById(deviceId)
}

func (n InterfaceNode) GetMuted() bool {
	params := n.Info.Params
	props := params["Props"].([]interface{})
	firstProp := props[0].(map[string]interface{})
	muted := firstProp["mute"].(bool)
	return muted
}

func (n InterfaceNode) GetVolume() []float64 {
	nodeInfoParamPropObjects := n.Info.Params["Props"].([]interface{})
	nodeInfoParamProps := nodeInfoParamPropObjects[0].(map[string]interface{})
	channelVolumes := nodeInfoParamProps["channelVolumes"].([]interface{})
	result := make([]float64, len(channelVolumes))
	// convert to 0-1 range
	for i, channel := range channelVolumes {
		result[i] = math.Cbrt(channel.(float64))
	}
	return result
}

func (i InterfaceNodeInfo) GetObjectID() int {
	objectId := i.Props["object.id"].(float64)
	return int(objectId)
}

func (i InterfaceNodeInfo) GetDeviceID() int {
	deviceId := i.Props["device.id"].(float64)
	return int(deviceId)
}

func (i InterfaceNodeInfo) GetClientID() int {
	clientId := i.Props["client.id"].(float64)
	return int(clientId)
}

func (i InterfaceNodeInfo) GetCardProfileDevice() int {
	cardProfileDevice := i.Props["card.profile.device"].(float64)
	return int(cardProfileDevice)
}

func (i InterfaceNodeInfo) GetDeviceRoutes() int {
	deviceRoutes := i.Props["device.routes"].(float64)
	return int(deviceRoutes)
}
