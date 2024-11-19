package wifi

import (
	"fmt"
	"github.com/markusressel/system-control/internal/util"
	"regexp"
	"strings"
)

/**


# Monitor Hotspot connections
function hotspot-connections -d "List devices connected to local Hotpost"
  printf  "# %-20s %-30s %-20s\n" "IP address" "lease name" "MAC address"

  # set -x leasefile "/var/lib/misc/dnsmasq.leases"
  set -x leasefile "/var/lib/NetworkManager/dnsmasq-wlp8s0.leases"
  set -x leases (sudo cat "$leasefile" | cut -f 2,3,4 -s -d" ")
  # list all wireless network interfaces (for MAC80211 driver; see wiki article for alternative commands)
  # ven out by dnsmasq, save it.
  for interface in (iw dev | grep I/etc/sudoers)
    set -x ip "UNKN"
    set -x host ""

    for lease in $leases
      if test (echo $lease | grep $mac)
        # echo yes
        # ... show the mac address:
        set -x ip (echo $lease | cut -f 2 -s -d" ")
        set -x host (echo $lease | cut -f 3 -s -d" ")
        printf "  %-20s %-30s %-20s\n" $ip $host $mac
      end
    end
  end
end

# Block/Unblock specific hotspot client
function block-hotspot-client -d "Block/Unblock specific hotspot client" -a action name
  set -x leasefile "/var/lib/NetworkManager/dnsmasq-wlp8s0.leases"
  set -x leases (sudo cat "$leasefile" | cut -f 2,3,4 -s -d" ")

  for lease in $leases
    if test (echo $lease | grep $name)
      set -x mac  (echo $lease | cut -f 1 -s -d" ")
      set -x ip   (echo $lease | cut -f 2 -s -d" ")
      set -x host (echo $lease | cut -f 3 -s -d" ")
      if set -q ip
        if test $action = "block"
          sudo iptables -I INPUT --source $ip --jump DROP
          sudo iptables -I FORWARD --source $ip --jump DROP
          echo "Blocking '$host'"
        else
          sudo iptables -D INPUT --source $ip --jump DROP
          sudo iptables -D FORWARD --source $ip --jump DROP
          echo "Unblocking '$host'"
        end
        return
      end
    end
  end
  echo "Client $name not found"
end

# Limit bandwith of connected hotspot clients
# https://www.linux.com/tutorials/bandwidth-monitoring-iptables/
function limit-hotspot-client -d "Block/Unblock specific hotspot client" -a name down up
  set -x leasefile "/var/lib/NetworkManager/dnsmasq-wlp8s0.leases"
  set -x leases (sudo cat "$leasefile" | cut -f 2,3,4 -s -d" ")

  # TODO: find dhcp range in dnsmasq config
  set -x dhcp_range ps aux | grep $leasefile | grep -oP '\--dhcp-range=([\d\.,]+)' | grep -oP '[\d\.]*'

  for lease in $leases
    if test (echo $lease | grep $name)
      set -x mac  (echo $lease | cut -f 1 -s -d" ")
      set -x ip   (echo $lease | cut -f 2 -s -d" ")
      set -x host (echo $lease | cut -f 3 -s -d" ")
      if set -q ip
        # create iptables chain for this client
        set -x ip_u (string replace -a "." "_" $ip)
        set -x table_name "HOTSPOT-CLIENT-$ip_u"

        if test $down -lt 0
          sudo iptables -D FORWARD -d $ip -j $table_name
          sudo iptables -D $table_name -d $ip
        end
        if test $up -lt 0
          sudo iptables -D FORWARD -s $ip -j $table_name
          sudo iptables -D $table_name -s $ip
        end

        # create custom chain
        if test $up -lt 0 && test $down -lt 0
          sudo iptables -X "HOTSPOT-CLIENT-$ip_u"
        else
          sudo iptables -N "HOTSPOT-CLIENT-$ip_u"
        end

        # setup rules to limit bandwith
        # download
        if test $down -ge 0
          sudo iptables -I FORWARD -d $ip -j $table_name
          sudo iptables -I $table_name -d $ip
        end

        # upload
        if test $up -ge 0
          sudo iptables -I FORWARD -s $ip -j $table_name
          sudo iptables -I $table_name -s $ip
        end
        return
      end
    end
  end
  echo "Client $name not found"
  return
end

*/

func TurnOnHotspot(name string) error {
	// try this: nmcli d wifi hotspot ifname <wifi_iface> ssid <ssid> password <password>

	return Connect(name)
}

func TurnOffHotspot(name string) error {
	_, err := util.ExecCommand(
		"nmcli",
		"connection",
		"down",
		name,
	)

	return err
}

type HotspotLease struct {
	IP   string
	Name string
	MAC  string
}

// GetConn
func GetConnectedHotspotDevices(ssid string) ([]HotspotLease, error) {
	result := make([]HotspotLease, 0)

	wifiInterface := "wlo1"

	wifiHotspotIsUp, _ := IsHotspotUp(ssid)
	if !wifiHotspotIsUp {
		return result, nil
	}

	output, err := util.ExecCommand(
		"iw",
		"dev",
		wifiInterface,
		"station",
		"dump",
	)
	if err != nil {
		return result, err
	}
	stationInfoList := ParseStationDump(output)

	leaseFilePath := fmt.Sprintf("/var/lib/NetworkManager/dnsmasq-%s.leases", wifiInterface)
	text, err := util.ReadTextFromFile(leaseFilePath)
	if err != nil {
		return result, err
	}
	hotspotLeases, err := util.ParseDelimitedTable(text, " ", func(row []string) HotspotLease {
		return HotspotLease{
			MAC:  row[1],
			IP:   row[2],
			Name: row[3],
		}
	})
	if err != nil {
		return result, err
	}

	for _, hotspotLease := range hotspotLeases {
		matchingStationInfo := util.FilterFunc(stationInfoList, func(e StationInfo) bool {
			return e.MAC == hotspotLease.MAC
		})
		if len(matchingStationInfo) > 0 {
			result = append(result, hotspotLease)
		} else {
			// Did not find matching StationInfo for mac %s, assuming its not connected.
		}
	}

	//// list all wireless network interfaces (for MAC80211 driver; see wiki article for alternative commands)
	//// ven out by dnsmasq, save it.
	//output, err := util.ExecCommand("iw", "dev")
	//if err != nil {
	//	return result, err
	//}
	//outputLines := strings.Split(output, "\n")
	//interfaces := util.FilterFunc(outputLines, func(e string) bool {
	//	return util.ContainsIgnoreCase(e, "/etc/sudoers")
	//})
	//if len(interfaces) == 0 {
	//	return result, nil
	//}
	//interfaceA := interfaces[0]
	//// TODO: filter output to get the interface
	//
	//// interfaces := "iw dev | grep I/etc/sudoers"
	//ip := "UNKN"
	//host := ""
	//

	return result, nil
}

type StationInfo struct {
	MAC        string
	Interface  string
	Properties map[string]string
}

// ParseStationDump
// Example:
// Station 04:f0:21:32:3f:d2 (on wlo1)
//
//	inactive time:	70 ms
//	rx bytes:	12257886
//	rx packets:	12577
//	tx bytes:	664158
//	tx packets:	3441
//	tx retries:	976
//	tx failed:	0
//	beacon loss:	0
//	beacon rx:	2325
//	rx drop misc:	207
//	signal:  	-60 [-60, -60] dBm
//	signal avg:	-59 dBm
//	beacon signal avg:	-58 dBm
//	tx bitrate:	650.0 MBit/s VHT-MCS 7 80MHz short GI VHT-NSS 2
//	tx duration:	0 us
//	rx bitrate:	650.0 MBit/s VHT-MCS 7 80MHz short GI VHT-NSS 2
//	rx duration:	0 us
//	authorized:	yes
//	authenticated:	yes
//	associated:	yes
//	preamble:	long
//	WMM/WME:	yes
//	MFP:		yes
//	TDLS peer:	no
//	DTIM period:	2
//	beacon interval:100
//	short slot time:yes
//	connected time:	233 seconds
//	associated at [boottime]:	20613.215s
//	associated at:	1731975464052 ms
//	current time:	1731975697273 ms
func ParseStationDump(output string) []StationInfo {
	lines := strings.Split(output, "\n")
	//currentIndent := 0

	var lastStation *StationInfo = nil
	result := []StationInfo{}
	for _, line := range lines {
		if !strings.HasPrefix(line, "\t") {
			parsed := parseStationDumpEntryHeader(line)
			lastStation = &parsed
			result = append(result, parsed)
		} else {
			result[len(result)-1] = parseStationDumpEntryProperty(*lastStation, line)
		}
	}

	return result
}

func parseStationDumpEntryProperty(info StationInfo, line string) StationInfo {
	parts := strings.Split(line, ":")
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	info.Properties[key] = value
	return info
}

func parseStationDumpEntryHeader(line string) StationInfo {
	// example line: Station 04:f0:21:32:3f:d2 (on wlo1)
	macRegexPattern := " ([0-9a-fA-F:]+) "
	macRegex := regexp.MustCompile(macRegexPattern)

	mac := macRegex.FindString(line)
	mac = strings.TrimSpace(mac)

	interfaceNameRegexPattern := "\\(on (.+)\\)"
	interfaceNameRegex := regexp.MustCompile(interfaceNameRegexPattern)

	interfaceName := interfaceNameRegex.FindString(line)
	interfaceName = strings.TrimSpace(interfaceName)

	return StationInfo{
		MAC:        mac,
		Interface:  interfaceName,
		Properties: map[string]string{},
	}
}

// IsHotspotUp checks if the given hotspot is currently running
func IsHotspotUp(ssid string) (bool, error) {
	networkInfo, err := GetNetworks()
	if err != nil {
		return false, err
	}
	networkInfo = util.FilterFunc(networkInfo, func(e WiFiNetwork) bool {
		return e.SSID == ssid
	})
	if len(networkInfo) <= 0 {
		return false, nil
	} else {
		return networkInfo[0].Connected, nil
	}
}
