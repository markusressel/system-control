package daemon

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"
)

// Source feeds change notifications into the engine, e.g. from an event stream.
// refresh re-queries all watched ops whose name starts with one of the given prefixes.
// Sources must return promptly; long running work belongs in a goroutine stopped by ctx.
type Source func(ctx context.Context, refresh func(prefixes ...string))

// Engine serves ops over the daemon socket.
type Engine struct {
	registry *Registry
	store    *store
	sources  []Source
	server   *ipcServer
}

func NewEngine(registry *Registry) *Engine {
	e := &Engine{
		registry: registry,
		store:    newStore(),
	}
	e.server = &ipcServer{engine: e}
	return e
}

// AddSource registers an event source. Must be called before Start.
func (e *Engine) AddSource(s Source) {
	e.sources = append(e.sources, s)
}

// Refresh re-queries all watched ops with a name matching one of the prefixes.
func (e *Engine) Refresh(prefixes ...string) {
	e.store.refreshMatching(prefixes...)
}

// Start opens the socket and starts all sources. It returns once the daemon is ready to serve.
func (e *Engine) Start(ctx context.Context) error {
	path := SocketPath()
	if err := os.MkdirAll(RuntimeDir(), 0700); err != nil {
		return err
	}

	// A socket that still accepts connections belongs to a running daemon, anything else is stale.
	if conn, err := net.DialTimeout("unix", path, 200*time.Millisecond); err == nil {
		_ = conn.Close()
		return fmt.Errorf("daemon already running on %s", path)
	}
	_ = os.Remove(path)

	if err := e.server.start(ctx, path); err != nil {
		return err
	}

	for _, s := range e.sources {
		s(ctx, e.Refresh)
	}
	return nil
}
