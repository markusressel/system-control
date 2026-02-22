package bluetooth

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	bt "tinygo.org/x/bluetooth"
)

type bluezAdapter struct {
	adapter    *bt.Adapter
	mu         sync.Mutex
	discovered map[string]BluetoothDevice
	scanning   bool
	// We store the context cancel func to stop the scan
	scanCancel context.CancelFunc
}

// NewBlueZAdapter returns an Adapter backed by tinygo.org/x/bluetooth (BlueZ on Linux).
func NewBlueZAdapter() *bluezAdapter {
	// Use the package-level variable directly
	ad := bt.DefaultAdapter
	_ = ad.Enable()

	return &bluezAdapter{
		adapter:    ad, // bt.DefaultAdapter is already *bt.Adapter
		discovered: make(map[string]BluetoothDevice),
	}
}

func (t *bluezAdapter) PowerOn() error {
	// powering on the adapter is platform specific; not supported here
	return ErrNotSupported
}

func (t *bluezAdapter) PowerOff() error {
	return ErrNotSupported
}

func (t *bluezAdapter) Scan(enable bool) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if enable {
		if t.scanning {
			return nil
		}

		// Use a context to handle the lifecycle of the scan
		ctx, cancel := context.WithCancel(context.Background())
		t.scanCancel = cancel
		t.scanning = true

		go func() {
			// NOTE: t.adapter.Scan is a blocking call
			err := t.adapter.Scan(func(adapter *bt.Adapter, device bt.ScanResult) {
				addr := device.Address.String()
				name := device.LocalName()
				if name == "" {
					name = "Unknown"
				}

				t.mu.Lock()
				t.discovered[addr] = BluetoothDevice{
					Name:    name,
					Address: addr,
					UUIDs:   make(map[string]uuid.UUID),
				}
				t.mu.Unlock()
			})

			if err != nil {
				// Handle scan error (e.g., adapter powered off during scan)
				t.mu.Lock()
				t.scanning = false
				t.mu.Unlock()
			}
		}()

		// StopScan will be invoked when ctx is cancelled
		go func() {
			<-ctx.Done()
			_ = t.adapter.StopScan()
		}()

		return nil
	}

	// Disable scanning
	if t.scanning && t.scanCancel != nil {
		t.scanCancel()
		t.scanning = false
	}
	return nil
}

func (t *bluezAdapter) ListDevices() ([]BluetoothDevice, error) {
	// Prefer BlueZ-managed devices via DBus; fall back to scan-based discovery
	if devs, err := listAllDevicesFromBlueZ(); err == nil && len(devs) > 0 {
		return devs, nil
	}

	// perform a short scan to collect nearby devices
	_ = t.Scan(true)
	// wait briefly to collect advertisements
	time.Sleep(2 * time.Second)
	_ = t.Scan(false)

	// Build result by enriching discovered advertisements with BlueZ info
	t.mu.Lock()
	addrs := make([]string, 0, len(t.discovered))
	for addr := range t.discovered {
		addrs = append(addrs, addr)
	}
	t.mu.Unlock()

	results := make([]BluetoothDevice, 0, len(addrs))
	for _, addr := range addrs {
		// Try BlueZ first
		if dev, err := getDeviceInfoFromBlueZ(addr); err == nil {
			results = append(results, dev)
			continue
		}
		// Fall back to advertisement info
		t.mu.Lock()
		if d, ok := t.discovered[addr]; ok {
			// Make sure UUIDs map is not nil
			if d.UUIDs == nil {
				d.UUIDs = make(map[string]uuid.UUID)
			}
			results = append(results, d)
		}
		t.mu.Unlock()
	}

	return results, nil
}

func (t *bluezAdapter) Info(address string) (BluetoothDevice, error) {
	// Try BlueZ first for rich info
	if dev, err := getDeviceInfoFromBlueZ(address); err == nil {
		return dev, nil
	}

	// Fall back to discovered advertisement info
	t.mu.Lock()
	defer t.mu.Unlock()
	if d, ok := t.discovered[address]; ok {
		if d.UUIDs == nil {
			d.UUIDs = make(map[string]uuid.UUID)
		}
		return d, nil
	}

	return BluetoothDevice{}, ErrNotSupported
}

func (t *bluezAdapter) Pair(address string) error       { return ErrNotSupported }
func (t *bluezAdapter) Connect(address string) error    { return ErrNotSupported }
func (t *bluezAdapter) Disconnect(address string) error { return ErrNotSupported }
func (t *bluezAdapter) DisconnectAll() error            { return ErrNotSupported }
func (t *bluezAdapter) Remove(address string) error     { return ErrNotSupported }
