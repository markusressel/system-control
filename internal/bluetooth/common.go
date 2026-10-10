package bluetooth

import (
	"github.com/google/uuid"
)

// BluetoothDevice represents a bluetooth device
type BluetoothDevice struct {
	Name              string // LG-TONE-FP9
	Address           string // B8:F8:BE:13:A4:72
	Alias             string // LG-TONE-FP9
	Class             string // 0x00240404
	Icon              string // audio-headset
	Paired            bool   // yes
	Bonded            bool   // yes
	Trusted           bool   // yes
	Blocked           bool   // no
	Connected         bool   // yes
	LegacyPairing     bool   // no
	UUIDs             map[string]uuid.UUID
	BatteryPercentage *int64 // 75
}

type BluetoothDeviceList []BluetoothDevice

func (r BluetoothDeviceList) Len() int { return len(r) }

func (r BluetoothDeviceList) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

func (r BluetoothDeviceList) Less(i, j int) bool { return r[i].Name < r[j].Name }
