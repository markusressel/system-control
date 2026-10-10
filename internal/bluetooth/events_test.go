package bluetooth

import (
	"testing"

	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
)

func TestIsRelevant(t *testing.T) {
	changed := func(iface string) *dbus.Signal {
		return &dbus.Signal{Name: "org.freedesktop.DBus.Properties.PropertiesChanged", Body: []interface{}{iface}}
	}

	assert.True(t, isRelevant(changed(deviceInterface)))
	assert.True(t, isRelevant(changed(batteryInterface)))
	assert.False(t, isRelevant(changed("org.bluez.MediaTransport1")))
	assert.False(t, isRelevant(&dbus.Signal{Name: "org.freedesktop.DBus.Properties.PropertiesChanged"}))
	assert.True(t, isRelevant(&dbus.Signal{Name: "org.freedesktop.DBus.ObjectManager.InterfacesAdded"}))
}
