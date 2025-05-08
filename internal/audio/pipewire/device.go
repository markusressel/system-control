package pipewire

import (
	"errors"
	"fmt"
	"github.com/markusressel/system-control/internal/util"
	"math"
	"strconv"
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

func (d InterfaceDevice) SetProfileByName(profileName string) error {
	profile, err := d.GetProfileIdByName(profileName)
	if err != nil {
		return err
	}

	_, err = util.ExecCommand(
		"pw-cli",
		"s",
		strconv.Itoa(d.Id),
		"Profile",
		fmt.Sprintf("{ index: %d, save: true }",
			profile.Index,
		),
	)
	return err
}

func (d InterfaceDevice) GetProfileIdByName(profileName string) (*DeviceProfile, error) {
	// search for exact match first
	for _, profile := range d.Info.Params.EnumProfile {
		if profile.Name == profileName || profile.Description == profileName {
			return &profile, nil
		}
	}

	// if no exact match, search for partial match
	for _, profile := range d.Info.Params.EnumProfile {
		if util.ContainsIgnoreCase(profile.Name, profileName) || util.ContainsIgnoreCase(profile.Description, profileName) {
			return &profile, nil
		}
	}

	return nil, errors.New("Profile not found: " + profileName)
}
