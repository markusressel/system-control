package pipewire

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
