package pipewire

import (
	"fmt"
	"github.com/markusressel/system-control/internal/util"
	"strconv"
	"strings"
)

type PipewireInterfaceDevice struct {
	CommonData
	Info PipewireInterfaceDeviceInfo
}

// PipewireInterfaceDeviceInfo Type: "PipeWire:Interface:Device"
type PipewireInterfaceDeviceInfo struct {
	ChangeMask []string               `json:"change-mask"`
	Props      map[string]interface{} `json:"props"`
	Params     DeviceInfoParams       `json:"params"`
}

type DeviceInfoParams struct {
	EnumProfile []DeviceProfile `json:"EnumProfile"`
	Profile     []DeviceProfile `json:"Profile"`
	EnumRoute   []DeviceRoute   `json:"EnumRoute"`
	Route       []DeviceRoute   `json:"Route"`
}

func (i DeviceInfoParams) GetOutputRoutes() []DeviceRoute {
	var outputRoutes []DeviceRoute
	for _, route := range i.Route {
		if strings.ToLower(route.Direction) == "output" {
			outputRoutes = append(outputRoutes, route)
		}
	}
	return outputRoutes
}

func (d PipewireInterfaceDevice) SetParameter(params map[string]interface{}) error {
	outputRoutes := d.Info.Params.GetOutputRoutes()

	formattedParams := ""
	for key, value := range params {
		formattedParams += fmt.Sprintf("%v: %v, ", key, value)
	}
	formattedParams = strings.TrimRight(formattedParams, ", ")

	for _, route := range outputRoutes {
		_, err := util.ExecCommand(
			"pw-cli",
			"set-param",
			strconv.Itoa(d.Id),
			"Route",
			fmt.Sprintf("{ index: %d, device: %d, props: { %s }",
				route.Index,
				route.Device,
				formattedParams,
			),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d PipewireInterfaceDevice) SetMuted(muted bool) error {
	return d.SetParameter(map[string]interface{}{
		"mute": muted, "save": true,
	})
}
