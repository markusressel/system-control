package daemon

import (
	"context"
)

type StateUpdate struct {
	// The unique file path relative to the tmpfs base, e.g., "audio/volume_device-Speakers"
	Path string
	// The new value to write to the file
	Value string
}

// Module represents a subsystem that can provide state updates to the daemon.
type Module interface {
	// Name returns the module name (e.g., "audio")
	Name() string

	// Track asks the module to start pushing updates for a specific command and flag combination.
	// Modules that support 'Push' will block on event listeners.
	// Modules that only support 'Pull' will use a ticker based on configuration.
	Track(ctx context.Context, command string, flags map[string]string, updates chan<- StateUpdate) error

	// Wakeup hints the module to immediately poll its state if it relies on polling.
	Wakeup()
}
