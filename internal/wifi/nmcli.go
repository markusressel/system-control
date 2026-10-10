package wifi

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/markusressel/system-control/internal/util"
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

// Connection represents a network connection
type Connection struct {
	Name           string // NAME
	UUID           string // UUID
	Type           string // TYPE
	Timestamp      string // TIMESTAMP
	TimestampReal  string // TIMESTAMP-REAL
	Autoconnect    string // AUTOCONNECT
	AutoconnectPri string // AUTOCONNECT-PRIORITY
	Readonly       string // READONLY
	DBUSPath       string // DBUS-PATH
	Active         string // ACTIVE
	Device         string // DEVICE
	State          string // STATE
	ActivePath     string // ACTIVE-PATH
	Slave          string // SLAVE
	Filename       string // FILENAME
}

const (
	connectionFields = "NAME,UUID,TYPE,TIMESTAMP,TIMESTAMP-REAL,AUTOCONNECT,AUTOCONNECT-PRIORITY,READONLY,DBUS-PATH,ACTIVE,DEVICE,STATE,ACTIVE-PATH,SLAVE,FILENAME"
	deviceFields     = "DEVICE,TYPE,STATE,IP4-CONNECTIVITY,IP6-CONNECTIVITY,DBUS-PATH,CONNECTION,CON-UUID,CON-PATH"
	networkFields    = "IN-USE,SSID,BSSID,MODE,CHAN,BANDWIDTH,FREQ,RATE,SIGNAL,BARS,SECURITY"
)

// splitTerse splits a line of `nmcli -t` output. Fields are separated by ':',
// a literal ':' or '\' in a value is escaped with a backslash.
func splitTerse(line string) []string {
	fields := []string{}
	var current strings.Builder
	escaped := false
	for _, r := range line {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == ':':
			fields = append(fields, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	fields = append(fields, current.String())
	return fields
}

// parseTerseTable splits nmcli terse output into rows of exactly the given number of columns.
func parseTerseTable(output string, columns int) ([][]string, error) {
	rows := [][]string{}
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		row := splitTerse(line)
		if len(row) != columns {
			return nil, fmt.Errorf("unexpected nmcli output, expected %d columns but got %d: %q", columns, len(row), line)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func nmcliTable(fields string, args ...string) ([][]string, error) {
	cmdArgs := append([]string{"-t", "-f", fields}, args...)
	output, err := util.ExecCommand("nmcli", cmdArgs...)
	if err != nil {
		return nil, err
	}
	return parseTerseTable(output, strings.Count(fields, ",")+1)
}

// Connect to a known WiFi network
func Connect(name string) error {
	connections, err := GetConnections()
	if err != nil {
		return err
	}
	knownConnection := false
	for _, connection := range connections {
		if connection.Name == name {
			knownConnection = true
			break
		}
	}

	if knownConnection {
		_, err = util.ExecCommand("nmcli", "connection", "up", name, "--ask")
	} else {
		_, err = util.ExecCommand("nmcli", "device", "wifi", "connect", name, "--ask")
	}

	return err
}

// Disconnect from the currently connected WiFi network, if any
func Disconnect() error {
	devices, err := GetNetworkDevices()
	if err != nil {
		return err
	}

	var errs []error
	for _, device := range devices {
		if device.Type != "wifi" || !strings.HasPrefix(device.State, "connected") {
			continue
		}
		_, err := util.ExecCommand("nmcli", "device", "disconnect", device.Name)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func GetConnections() ([]Connection, error) {
	rows, err := nmcliTable(connectionFields, "connection", "show")
	if err != nil {
		return nil, err
	}
	return parseConnections(rows), nil
}

func parseConnections(rows [][]string) []Connection {
	connections := make([]Connection, 0, len(rows))
	for _, row := range rows {
		connections = append(connections, Connection{
			Name:           row[0],
			UUID:           row[1],
			Type:           row[2],
			Timestamp:      row[3],
			TimestampReal:  row[4],
			Autoconnect:    row[5],
			AutoconnectPri: row[6],
			Readonly:       row[7],
			DBUSPath:       row[8],
			Active:         row[9],
			Device:         row[10],
			State:          row[11],
			ActivePath:     row[12],
			Slave:          row[13],
			Filename:       row[14],
		})
	}
	return connections
}

func GetNetworkDevices() ([]NetworkDevice, error) {
	rows, err := nmcliTable(deviceFields, "device")
	if err != nil {
		return nil, err
	}
	return parseNetworkDevices(rows), nil
}

func parseNetworkDevices(rows [][]string) []NetworkDevice {
	devices := make([]NetworkDevice, 0, len(rows))
	for _, row := range rows {
		devices = append(devices, NetworkDevice{
			Name:            row[0],
			Type:            row[1],
			State:           row[2],
			IP4Connectivity: row[3],
			IP6Connectivity: row[4],
			DBUSPath:        row[5],
			Connection:      row[6],
			CONUUID:         row[7],
			CONPath:         row[8],
		})
	}
	return devices
}

// GetNetworks returns a list of all visible WiFi networks
func GetNetworks() ([]WiFiNetwork, error) {
	rows, err := nmcliTable(networkFields, "device", "wifi", "list")
	if err != nil {
		return nil, err
	}
	return parseNetworks(rows), nil
}

func parseNetworks(rows [][]string) []WiFiNetwork {
	networks := make([]WiFiNetwork, 0, len(rows))
	for _, row := range rows {
		channel, _ := strconv.Atoi(row[4])
		signal, _ := strconv.Atoi(row[8])
		networks = append(networks, WiFiNetwork{
			Connected: row[0] == "*",
			SSID:      row[1],
			BSSID:     row[2],
			Mode:      row[3],
			Channel:   channel,
			Bandwidth: row[5],
			Frequency: row[6],
			Rate:      row[7],
			Signal:    signal,
			Bars:      row[9],
			Security:  row[10],
		})
	}
	return networks
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
	return util.ExecCommandAndFork("nm-connection-editor")
}

// TurnOnWifiAdapter turns on the WiFi adapter
func TurnOnWifiAdapter() error {
	_, err := util.ExecCommand("nmcli", "radio", "wifi", "on")
	return err
}

// TurnOffWifiAdapter turns off the WiFi adapter
func TurnOffWifiAdapter() error {
	_, err := util.ExecCommand("nmcli", "radio", "wifi", "off")
	return err
}

// GetWifiInterface returns the name of the first WiFi device.
func GetWifiInterface() (string, error) {
	devices, err := GetNetworkDevices()
	if err != nil {
		return "", err
	}
	for _, device := range devices {
		if device.Type == "wifi" {
			return device.Name, nil
		}
	}
	return "", errors.New("no WiFi device found")
}
