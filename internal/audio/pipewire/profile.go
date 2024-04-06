package pipewire

type DeviceProfile struct {
	Index       int           `json:"index"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Available   string        `json:"available"`
	Priority    int           `json:"priority"`
	Classes     []interface{} `json:"classes"`
	Save        bool          `json:"save,omitempty"`
}
