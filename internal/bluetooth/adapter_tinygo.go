package bluetooth

// tinyGoAdapter is a placeholder implementation that will be replaced with a
// real implementation using tinygo.org/x/bluetooth. For now it returns
// ErrNotSupported for operations that are not implemented yet so callers can
// fall back to the bluetoothctl adapter if desired.
type tinyGoAdapter struct{}

func (t *tinyGoAdapter) PowerOn() error {
	return ErrNotSupported
}

func (t *tinyGoAdapter) PowerOff() error {
	return ErrNotSupported
}

func (t *tinyGoAdapter) Scan(enable bool) error {
	return ErrNotSupported
}

func (t *tinyGoAdapter) ListDevices() ([]BluetoothDevice, error) {
	// not implemented yet
	return []BluetoothDevice{}, nil
}

func (t *tinyGoAdapter) ListConnected() ([]BluetoothDevice, error) {
	return nil, ErrNotSupported
}

func (t *tinyGoAdapter) ListPaired() ([]BluetoothDevice, error) {
	return nil, ErrNotSupported
}

func (t *tinyGoAdapter) Info(address string) (BluetoothDevice, error) {
	return BluetoothDevice{}, ErrNotSupported
}

func (t *tinyGoAdapter) Pair(address string) error {
	return ErrNotSupported
}

func (t *tinyGoAdapter) Connect(address string) error {
	return ErrNotSupported
}

func (t *tinyGoAdapter) Disconnect(address string) error {
	return ErrNotSupported
}

func (t *tinyGoAdapter) DisconnectAll() error {
	return ErrNotSupported
}

func (t *tinyGoAdapter) Remove(address string) error {
	return ErrNotSupported
}
