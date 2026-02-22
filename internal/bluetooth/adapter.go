package bluetooth

import "errors"

// Adapter defines the operations required by the rest of the application.
// Implementations may use tinygo or shelling out to bluetoothctl.
type Adapter interface {
	PowerOn() error
	PowerOff() error
	Scan(enable bool) error
	ListDevices() ([]BluetoothDevice, error)
	ListConnected() ([]BluetoothDevice, error)
	ListPaired() ([]BluetoothDevice, error)
	Info(address string) (BluetoothDevice, error)
	Pair(address string) error
	Connect(address string) error
	Disconnect(address string) error
	DisconnectAll() error
	Remove(address string) error
}

var (
	impl Adapter
	// ErrNoAdapter is returned when no adapter implementation is configured.
	ErrNoAdapter = errors.New("no bluetooth adapter available")
	// ErrNotSupported is returned when an adapter doesn't support a specific operation.
	ErrNotSupported = errors.New("operation not supported by adapter")
)

// init sets a sensible default adapter so existing behavior remains unchanged.
func init() {
	// default to tinygo-backed adapter
	impl = NewTinyGoAdapter()
}

// SetAdapter allows overriding the adapter implementation (useful for tests).
func SetAdapter(a Adapter) {
	impl = a
}

// UseBluetoothCtl forces selecting the bluetoothctl adapter. Kept for compatibility but no-op now.
func UseBluetoothCtl() error {
	// bluetoothctl adapter was removed; return not supported
	return ErrNotSupported
}

// Public wrappers delegating to the selected adapter. These preserve the original
// package-level API so cmd/* packages don't need changes.

func TurnOnBluetoothAdapter() error {
	if impl == nil {
		return ErrNoAdapter
	}
	return impl.PowerOn()
}

func TurnOffBluetoothAdapter() error {
	if impl == nil {
		return ErrNoAdapter
	}
	return impl.PowerOff()
}

func SetBluetoothScan(enable bool) error {
	if impl == nil {
		return ErrNoAdapter
	}
	return impl.Scan(enable)
}

func GetBluetoothDevices() ([]BluetoothDevice, error) {
	if impl == nil {
		return nil, ErrNoAdapter
	}
	return impl.ListDevices()
}

func GetBluetoothDeviceInfo(device simpleDeviceInfo) (BluetoothDevice, error) {
	if impl == nil {
		return BluetoothDevice{}, ErrNoAdapter
	}
	return impl.Info(device.Address)
}

func PairBluetoothDevice(device BluetoothDevice) error {
	if impl == nil {
		return ErrNoAdapter
	}
	return impl.Pair(device.Address)
}

func ConnectToBluetoothDevice(device BluetoothDevice) error {
	if impl == nil {
		return ErrNoAdapter
	}
	return impl.Connect(device.Address)
}

func DisconnectAllBluetoothDevices() error {
	if impl == nil {
		return ErrNoAdapter
	}
	return impl.DisconnectAll()
}

func DisconnectBluetoothDevice(device BluetoothDevice) error {
	if impl == nil {
		return ErrNoAdapter
	}
	return impl.Disconnect(device.Address)
}

func RemoveBluetoothDevice(device BluetoothDevice) error {
	if impl == nil {
		return ErrNoAdapter
	}
	return impl.Remove(device.Address)
}
