package bluetooth

import (
	"github.com/google/uuid"
)

// BluetoothDevice represents a bluetooth device
type BluetoothDevice struct {
	Name              string // LG-TONE-FP9
	Address           string // B8:F8:BE:13:A4:72
	Alias             string // LG-TONE-FP9
	Class             string // 0x00240404 (2360324)
	Icon              string // audio-headset
	Paired            bool   // yes
	Bonded            bool   // yes
	Trusted           bool   // yes
	Blocked           bool   // no
	Connected         bool   // yes
	LegacyPairing     bool   // no
	UUIDs             map[string]uuid.UUID
	BatteryPercentage *int64 // 0x4b (75)
}

// simpleDeviceInfo is retained for compatibility with callers that may still use it
// (internal callers should prefer using Address strings directly).
type simpleDeviceInfo struct {
	Name    string
	Address string
}
