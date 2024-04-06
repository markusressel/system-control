package pipewire

import (
	"fmt"
	"github.com/markusressel/system-control/internal/util"
	"strconv"
	"strings"
)

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

func (r DeviceRoute) SetParameter(deviceId int, params map[string]interface{}) error {
	formattedParams := ""
	for key, value := range params {
		formattedParams += fmt.Sprintf("%v: %v, ", key, value)
	}
	formattedParams = strings.TrimRight(formattedParams, ", ")

	_, err := util.ExecCommand(
		"pw-cli",
		"set-param",
		strconv.Itoa(deviceId),
		"Route",
		fmt.Sprintf("{ index: %d, device: %d, props: { %s }",
			r.Index,
			r.Device,
			formattedParams,
		),
	)
	return err
}
