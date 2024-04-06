package pipewire

type PipewireInterfaceNode struct {
	CommonData
	Info PipewireInterfaceNodeInfo
}

func (node *PipewireInterfaceNode) GetMuted() bool {
	params := node.Info.Params
	props := params["Props"].([]interface{})
	firstProp := props[0].(map[string]interface{})
	muted := firstProp["mute"].(bool)
	return muted
}
