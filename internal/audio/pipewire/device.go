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

type DeviceProfile struct {
	Index       int           `json:"index"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Available   string        `json:"available"`
	Priority    int           `json:"priority"`
	Classes     []interface{} `json:"classes"`
	Save        bool          `json:"save,omitempty"`
}

type DeviceRoute struct {
	Index       int                    `json:"index"`
	Direction   string                 `json:"direction"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Priority    int                    `json:"priority"`
	Available   string                 `json:"available"`
	Info        []interface{}          `json:"info"`
	Profiles    []int                  `json:"profiles"`
	Device      int                    `json:"device"`
	Props       map[string]interface{} `json:"props"`
	Save        bool                   `json:"save,omitempty"`
	Devices     []int                  `json:"devices"`
	Profile     int                    `json:"profile"`
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
