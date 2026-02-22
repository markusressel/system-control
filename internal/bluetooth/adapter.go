package bluetooth

import "errors"

var (
	// global bluez adapter instance
	bluez *bluezAdapter

	// ErrNotSupported is returned when an adapter doesn't support a specific operation.
	ErrNotSupported = errors.New("operation not supported by adapter")
)

func init() {
	bluez = NewBlueZAdapter()
}

func TurnOnBluetoothAdapter() error {
	return bluez.PowerOn()
}

func TurnOffBluetoothAdapter() error {
	return bluez.PowerOff()
}

func SetBluetoothScan(enable bool) error {
	return bluez.Scan(enable)
}

func GetBluetoothDevices() ([]BluetoothDevice, error) {
	return bluez.ListDevices()
}

func GetBluetoothDeviceInfo(device simpleDeviceInfo) (BluetoothDevice, error) {
	return bluez.Info(device.Address)
}

func PairBluetoothDevice(device BluetoothDevice) error {
	return bluez.Pair(device.Address)
}

func ConnectToBluetoothDevice(device BluetoothDevice) error {
	return bluez.Connect(device.Address)
}

func DisconnectAllBluetoothDevices() error {
	return bluez.DisconnectAll()
}

func DisconnectBluetoothDevice(device BluetoothDevice) error {
	return bluez.Disconnect(device.Address)
}

func RemoveBluetoothDevice(device BluetoothDevice) error {
	return bluez.Remove(device.Address)
}
