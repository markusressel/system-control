package pipewire

type PipewireInterfaceNode struct {
	CommonData
	Info PipewireInterfaceNodeInfo
}

func (n PipewireInterfaceNode) GetMuted() bool {
	params := n.Info.Params
	props := params["Props"].([]interface{})
	firstProp := props[0].(map[string]interface{})
	muted := firstProp["mute"].(bool)
	return muted
}
