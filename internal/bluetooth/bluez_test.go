package bluetooth

import (
	"testing"

	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
)

func variant(v interface{}) dbus.Variant { return dbus.MakeVariant(v) }

func fakeObjects() managedObjects {
	return managedObjects{
		"/org/bluez/hci0": {
			adapterInterface: {"Powered": variant(true)},
		},
		"/org/bluez/hci0/dev_B8_F8_BE_13_A4_72": {
			deviceInterface: {
				"Address":   variant("B8:F8:BE:13:A4:72"),
				"Alias":     variant("LG-TONE-FP9"),
				"Class":     variant(uint32(0x240404)),
				"Paired":    variant(true),
				"Connected": variant(true),
				"UUIDs":     variant([]string{"0000110b-0000-1000-8000-00805f9b34fb", "not-a-uuid"}),
			},
			batteryInterface: {"Percentage": variant(byte(75))},
		},
		"/org/bluez/hci0/dev_00_11_22_33_44_55": {
			deviceInterface: {
				"Address": variant("00:11:22:33:44:55"),
				"Name":    variant("Mouse"),
			},
		},
	}
}

func TestDeviceFromInterfaces(t *testing.T) {
	objects := fakeObjects()

	dev, ok := deviceFromInterfaces(objects["/org/bluez/hci0/dev_B8_F8_BE_13_A4_72"])

	assert.True(t, ok)
	assert.Equal(t, "B8:F8:BE:13:A4:72", dev.Address)
	assert.Equal(t, "LG-TONE-FP9", dev.Name, "falls back to alias")
	assert.Equal(t, "0x00240404", dev.Class)
	assert.True(t, dev.Paired)
	assert.True(t, dev.Connected)
	assert.False(t, dev.Trusted)
	assert.Len(t, dev.UUIDs, 1)
	if assert.NotNil(t, dev.BatteryPercentage) {
		assert.Equal(t, int64(75), *dev.BatteryPercentage)
	}
}

func TestDeviceFromInterfacesIgnoresOtherObjects(t *testing.T) {
	_, ok := deviceFromInterfaces(fakeObjects()["/org/bluez/hci0"])
	assert.False(t, ok)
}

func TestDefaultAdapterPath(t *testing.T) {
	path, err := defaultAdapterPath(fakeObjects())
	assert.NoError(t, err)
	assert.Equal(t, dbus.ObjectPath("/org/bluez/hci0"), path)

	_, err = defaultAdapterPath(managedObjects{})
	assert.ErrorIs(t, err, ErrNoAdapter)
}

func TestFindDevicePath(t *testing.T) {
	path, err := findDevicePath(fakeObjects(), "00:11:22:33:44:55")
	assert.NoError(t, err)
	assert.Equal(t, dbus.ObjectPath("/org/bluez/hci0/dev_00_11_22_33_44_55"), path)

	_, err = findDevicePath(fakeObjects(), "FF:FF:FF:FF:FF:FF")
	assert.Error(t, err)
}
