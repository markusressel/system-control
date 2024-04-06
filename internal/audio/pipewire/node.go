package pipewire

type InterfaceNode struct {
	CommonData
	Info InterfaceNodeInfo
}

func (n InterfaceNode) GetMuted() bool {
	params := n.Info.Params
	props := params["Props"].([]interface{})
	firstProp := props[0].(map[string]interface{})
	muted := firstProp["mute"].(bool)
	return muted
}
