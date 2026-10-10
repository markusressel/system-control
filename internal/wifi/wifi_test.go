package wifi

import (
	"context"
	"strings"
	"testing"

	"github.com/markusressel/system-control/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestSplitTerse(t *testing.T) {
	assert.Equal(t, []string{"*", "My: SSID", "AA:BB:CC:DD:EE:FF", ""},
		splitTerse(`*:My\: SSID:AA\:BB\:CC\:DD\:EE\:FF:`))
	assert.Equal(t, []string{`a\b`, "c"}, splitTerse(`a\\b:c`))
	assert.Equal(t, []string{""}, splitTerse(""))
}

func TestParseTerseTableRejectsWrongColumnCount(t *testing.T) {
	_, err := parseTerseTable("a:b:c", 2)
	assert.Error(t, err)

	rows, err := parseTerseTable("a:b\n\nc:d\n", 2)
	assert.NoError(t, err)
	assert.Equal(t, [][]string{{"a", "b"}, {"c", "d"}}, rows)
}

func TestParseNetworks(t *testing.T) {
	output := `*:Home 📶:AA\:BB\:CC\:DD\:EE\:FF:Infra:36:80 MHz:5180 MHz:540 Mbit/s:100:▂▄▆█:WPA2
 :Guest:11\:22\:33\:44\:55\:66:Infra:1:20 MHz:2412 MHz:130 Mbit/s:42:▂▄__:WPA2 WPA3`
	rows, err := parseTerseTable(output, 11)
	assert.NoError(t, err)

	networks := parseNetworks(rows)

	assert.Len(t, networks, 2)
	assert.True(t, networks[0].Connected)
	assert.Equal(t, "Home 📶", networks[0].SSID)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", networks[0].BSSID)
	assert.Equal(t, 36, networks[0].Channel)
	assert.Equal(t, 100, networks[0].Signal)
	assert.False(t, networks[1].Connected)
	assert.Equal(t, "WPA2 WPA3", networks[1].Security)
}

func TestParseStationDump(t *testing.T) {
	output := "Station 04:f0:21:32:3f:d2 (on wlo1)\n" +
		"\tinactive time:\t70 ms\n" +
		"\tassociated at [boottime]:\t20613.215s\n" +
		"Station 0a:00:00:00:00:01 (on wlo1)\n" +
		"\tsignal:  \t-60 [-60, -60] dBm\n"

	stations := ParseStationDump(output)

	assert.Len(t, stations, 2)
	assert.Equal(t, "04:f0:21:32:3f:d2", stations[0].MAC)
	assert.Equal(t, "wlo1", stations[0].Interface)
	assert.Equal(t, "70 ms", stations[0].Properties["inactive time"])
	assert.Equal(t, "20613.215s", stations[0].Properties["associated at [boottime]"])
	assert.Equal(t, "-60 [-60, -60] dBm", stations[1].Properties["signal"])
}

func TestParseStationDumpEmpty(t *testing.T) {
	assert.Empty(t, ParseStationDump(""))
	assert.Empty(t, ParseStationDump("\tstray: line\n"))
}

func TestParseLeases(t *testing.T) {
	text := "1731980000 04:f0:21:32:3f:d2 10.42.0.15 phone 01:04:f0:21:32:3f:d2\n\nshort line\n"

	assert.Equal(t, []HotspotLease{{MAC: "04:f0:21:32:3f:d2", IP: "10.42.0.15", Name: "phone"}}, ParseLeases(text))
}

func TestDisconnectOnlyDisconnectsConnectedWifiDevices(t *testing.T) {
	var commands []string
	restore := util.SetRunner(func(ctx context.Context, command string, args ...string) (string, error) {
		line := command + " " + strings.Join(args, " ")
		commands = append(commands, line)
		if strings.HasPrefix(line, "nmcli -t -f DEVICE") {
			return "wlo1:wifi:connected:full:none:/p:Home:uuid:/c\n" +
				"eno1:ethernet:connected:full:none:/p:Wired:uuid:/c\n" +
				"wlp2:wifi:disconnected:none:none::::", nil
		}
		return "", nil
	})
	defer restore()

	assert.NoError(t, Disconnect())
	assert.Contains(t, commands, "nmcli device disconnect wlo1")
	assert.NotContains(t, commands, "nmcli device disconnect eno1")
	assert.NotContains(t, commands, "nmcli device disconnect wlp2")
}
