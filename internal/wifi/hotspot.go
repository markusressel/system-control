package wifi

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/markusressel/system-control/internal/util"
)

// TurnOnHotspot creates and activates a hotspot connection with the given name.
// If password is empty NetworkManager generates one.
func TurnOnHotspot(
	name string,
	wifiInterface string,
	ssid string,
	password string,
) error {
	args := []string{
		"device", "wifi", "hotspot",
		"ifname", wifiInterface,
		"con-name", name,
		"ssid", ssid,
	}
	if password != "" {
		args = append(args, "password", password)
	}

	_, err := util.ExecCommand("nmcli", args...)
	return err
}

func TurnOffHotspot(name string) error {
	_, err := util.ExecCommand("nmcli", "connection", "down", name)
	return err
}

type HotspotLease struct {
	IP   string
	Name string
	MAC  string
}

// GetConnectedHotspotDevices returns a list of devices currently connected to the hotspot connection with the given name.
// It does this by parsing the output of "iw dev <interface> station dump" and matching it with the leases in the
// dnsmasq lease file for the given interface. Only devices that have a matching station info entry
// (i.e. are currently connected) are returned.
func GetConnectedHotspotDevices(wifiInterface string, connectionName string) ([]HotspotLease, error) {
	result := make([]HotspotLease, 0)

	wifiHotspotIsUp, _ := IsHotspotUp(connectionName)
	if !wifiHotspotIsUp {
		return result, nil
	}

	output, err := util.ExecCommand("iw", "dev", wifiInterface, "station", "dump")
	if err != nil {
		return result, err
	}
	stationInfoList := ParseStationDump(output)

	leaseFilePath := fmt.Sprintf("/var/lib/NetworkManager/dnsmasq-%s.leases", wifiInterface)
	text, err := util.ReadTextFromFile(leaseFilePath)
	if err != nil {
		return result, err
	}

	for _, lease := range ParseLeases(text) {
		for _, station := range stationInfoList {
			if strings.EqualFold(station.MAC, lease.MAC) {
				result = append(result, lease)
				break
			}
		}
	}

	return result, nil
}

// ParseLeases parses a dnsmasq lease file, whose lines have the format:
//
//	<expiry> <mac> <ip> <hostname> <client id>
func ParseLeases(text string) []HotspotLease {
	leases := []HotspotLease{}
	for _, line := range strings.Split(text, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		leases = append(leases, HotspotLease{MAC: fields[1], IP: fields[2], Name: fields[3]})
	}
	return leases
}

type StationInfo struct {
	MAC        string
	Interface  string
	Properties map[string]string
}

var (
	stationHeaderRegex = regexp.MustCompile(`^Station\s+([0-9a-fA-F:]+)\s+\(on\s+(.+)\)`)
)

// ParseStationDump parses the output of `iw dev <interface> station dump`.
// Example:
//
//	Station 04:f0:21:32:3f:d2 (on wlo1)
//		inactive time:	70 ms
//		rx bytes:	12257886
//		signal:  	-60 [-60, -60] dBm
//		associated at [boottime]:	20613.215s
func ParseStationDump(output string) []StationInfo {
	result := []StationInfo{}

	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "\t") {
			match := stationHeaderRegex.FindStringSubmatch(line)
			if match == nil {
				continue
			}
			result = append(result, StationInfo{
				MAC:        match[1],
				Interface:  match[2],
				Properties: map[string]string{},
			})
			continue
		}

		if len(result) == 0 {
			continue
		}
		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		result[len(result)-1].Properties[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	return result
}

// IsHotspotUp checks if the hotspot connection with the given name is currently active.
func IsHotspotUp(connectionName string) (bool, error) {
	connections, err := GetConnections()
	if err != nil {
		return false, err
	}
	for _, connection := range connections {
		if connection.Name == connectionName {
			return connection.Active == "yes", nil
		}
	}
	return false, nil
}
