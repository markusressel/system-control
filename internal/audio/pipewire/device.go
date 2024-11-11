package pipewire

import (
	"github.com/markusressel/system-control/internal/util"
	"math"
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
		if util.EqualsIgnoreCase(route.Direction, "output") {
			outputRoutes = append(outputRoutes, route)
		}
	}
	return outputRoutes
}

func (d InterfaceDevice) SetMuted(muted bool) error {
	outputRoutes := d.Info.Params.GetOutputRoutes()

	for _, route := range outputRoutes {
		err := route.SetProps(d.Id, map[string]interface{}{
			"mute": muted,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (d InterfaceDevice) SetVolume(volume float64) error {
	outputRoutes := d.Info.Params.GetOutputRoutes()

	if volume < 0 {
		volume = 0
	} else if volume > 1 {
		volume = 1
	}
	volumeCubicRoot := math.Pow(volume, 3)

	for _, route := range outputRoutes {
		err := route.SetProps(d.Id, map[string]interface{}{
			"muted":          false,
			"channelVolumes": []float64{volumeCubicRoot, volumeCubicRoot},
			"save":           true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
