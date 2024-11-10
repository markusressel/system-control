package wifi

import (
	"github.com/markusressel/system-control/internal/util"
	"strings"
)

// Connect to a known WiFi network
func Connect(name string) error {
	_, err := util.ExecCommand(
		"nmcli",
		"connection",
		"up",
		name,
	)
	return err
}

type WiFiNetwork struct {
	Connected bool
	BSSID     string
	SSID      string
	Mode      string
	Channel   string
	Bandwidth string
	Frequency string
	Rate      string
	Signal    string
	Bars      string
	Security  string
}

func GetNetworks() ([]WiFiNetwork, error) {
	output, err := util.ExecCommand(
		"nmcli",
		"-f",
		"in-use,ssid,bssid,mode,chan,bandwidth,freq,rate,signal,bars,security",
		// WPA-FLAGS  RSN-FLAGS                     DEVICE  ACTIVE  IN->
		"device",
		"wifi",
		"list",
	)
	if err != nil {
		return nil, err
	}

	wifiNetworks, err := util.ParseTable(
		output,
		util.DefaultColumnHeaderRegexPattern,
		func(row []string) WiFiNetwork {
			return WiFiNetwork{
				Connected: strings.Contains(row[0], "*"),
				SSID:      strings.TrimSpace(row[1]),
				BSSID:     strings.TrimSpace(row[2]),
				Mode:      strings.TrimSpace(row[3]),
				Channel:   strings.TrimSpace(row[4]),
				Bandwidth: strings.TrimSpace(row[5]),
				Frequency: strings.TrimSpace(row[6]),
				Rate:      strings.TrimSpace(row[7]),
				Signal:    strings.TrimSpace(row[8]),
				Bars:      strings.TrimSpace(row[9]),
				Security:  strings.TrimSpace(row[10]),
			}
		})

	return wifiNetworks, err
}

func OpenManageGui() error {
	_, err := util.ExecCommand(
		"nm-connection-editor",
	)
	return err
}
