package pipewire

import (
	"fmt"
	"github.com/markusressel/system-control/internal/util"
	"math"
	"strconv"
	"strings"
)

type InterfaceDevice struct {
	CommonData
	Info InterfaceDeviceInfo
}

// InterfaceDeviceInfo Type: "PipeWire:Interface:Device"
type InterfaceDeviceInfo struct {
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

func (d InterfaceDevice) SetParameter(params map[string]interface{}) error {
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

func (d InterfaceDevice) SetMuted(muted bool) error {
	return d.SetParameter(map[string]interface{}{
		"mute": muted,
		"save": true,
	})
}

func (d InterfaceDevice) SetVolume(volume float64) error {
	if volume < 0 {
		volume = 0
	} else if volume > 1 {
		volume = 1
	}
	volumeCubicRoot := math.Pow(volume, 3)

	return d.SetParameter(map[string]interface{}{
		"muted":          false,
		"channelVolumes": []float64{volumeCubicRoot, volumeCubicRoot},
		"save":           true,
	})
}
