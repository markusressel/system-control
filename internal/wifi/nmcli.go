package wifi

import (
	"github.com/markusressel/system-control/internal/util"
	"strconv"
	"strings"
)

type NetworkDevice struct {
	Name            string // DEVICE
	Type            string // TYPE
	State           string // STATE
	IP4Connectivity string // IP4-CONNECTIVITY
	IP6Connectivity string // IP6-CONNECTIVITY
	DBUSPath        string // DBUS-PATH
	Connection      string // CONNECTION
	CONUUID         string // CON-UUID
	CONPath         string // CON-PATH
}

// WiFiNetwork represents a WiFi network
type WiFiNetwork struct {
	Connected bool
	BSSID     string
	SSID      string
	Mode      string
	Channel   int
	Bandwidth string
	Frequency string
	Rate      string
	Signal    int
	Bars      string
	Security  string
}

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

// Disconnect from the currently connected WiFi network, if any
func Disconnect() error {
	connectedNetwork, err := GetConnectedNetwork()
	if err != nil {
		return err
	}

	if connectedNetwork == nil {
		return nil
	}

	_, err = util.ExecCommand(
		"nmcli",
		"connection",
		"down",
		connectedNetwork.SSID,
	)
	return err
}

func GetNetworkDevices() ([]NetworkDevice, error) {
	output, err := util.ExecCommand(
		"nmcli",
		"-f",
		"DEVICE,TYPE,STATE,IP4-CONNECTIVITY,IP6-CONNECTIVITY,DBUS-PATH,CONNECTION,CON-UUID,CON-PATH",
		"device",
	)
	if err != nil {
		return nil, err
	}

	devices, err := util.ParseTable(
		output,
		util.DefaultColumnHeaderRegexPattern,
		func(row []string) NetworkDevice {
			return NetworkDevice{
				Name:            strings.TrimSpace(row[0]),
				Type:            strings.TrimSpace(row[1]),
				State:           strings.TrimSpace(row[2]),
				IP4Connectivity: strings.TrimSpace(row[3]),
				IP6Connectivity: strings.TrimSpace(row[4]),
				DBUSPath:        strings.TrimSpace(row[5]),
				Connection:      strings.TrimSpace(row[6]),
				CONUUID:         strings.TrimSpace(row[7]),
				CONPath:         strings.TrimSpace(row[8]),
			}
		})

	return devices, err
}

// GetNetworks returns a list of all known WiFi networks
func GetNetworks() ([]WiFiNetwork, error) {
	output, err := util.ExecCommand(
		"nmcli",
		"-f",
		"in-use,ssid,bssid,mode,chan,bandwidth,freq,rate,signal,bars,security",
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
			channelInt, _ := strconv.Atoi(strings.TrimSpace(row[4]))
			signalInt, _ := strconv.Atoi(strings.TrimSpace(row[8]))
			return WiFiNetwork{
				Connected: strings.Contains(row[0], "*"),
				SSID:      strings.TrimSpace(row[1]),
				BSSID:     strings.TrimSpace(row[2]),
				Mode:      strings.TrimSpace(row[3]),
				Channel:   channelInt,
				Bandwidth: strings.TrimSpace(row[5]),
				Frequency: strings.TrimSpace(row[6]),
				Rate:      strings.TrimSpace(row[7]),
				Signal:    signalInt,
				Bars:      strings.TrimSpace(row[9]),
				Security:  strings.TrimSpace(row[10]),
			}
		})

	return wifiNetworks, err
}

// GetConnectedNetwork returns the currently connected WiFi network, if any
func GetConnectedNetwork() (*WiFiNetwork, error) {
	networks, err := GetNetworks()
	if err != nil {
		return nil, err
	}

	for _, network := range networks {
		if network.Connected {
			return &network, nil
		}
	}

	return nil, nil
}

// OpenManageGui opens the network manager GUI
func OpenManageGui() error {
	err := util.ExecCommandAndFork(
		"nm-connection-editor",
	)
	return err
}

// TurnOnWifiAdapter turns on the WiFi adapter
func TurnOnWifiAdapter() error {
	_, err := util.ExecCommand(
		"nmcli",
		"radio",
		"wifi",
		"on",
	)
	return err
}

// TurnOffWifiAdapter turns off the WiFi adapter
func TurnOffWifiAdapter() error {
	_, err := util.ExecCommand(
		"nmcli",
		"radio",
		"wifi",
		"off",
	)
	return err
}
