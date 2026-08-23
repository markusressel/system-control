package daemon

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
)

type Engine struct {
	basePath    string
	socketPath  string
	modules     map[string]Module
	updates     chan StateUpdate
	active      map[string]bool // keys are tracked file paths
	activeMutex sync.Mutex
}

func NewEngine() (*Engine, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	basePath := filepath.Join("/run/user", u.Uid, "system-control")

	return &Engine{
		basePath:   basePath,
		socketPath: filepath.Join(basePath, "ipc.sock"),
		modules:    make(map[string]Module),
		updates:    make(chan StateUpdate, 100),
		active:     make(map[string]bool),
	}, nil
}

func (e *Engine) RegisterModule(m Module) {
	e.modules[m.Name()] = m
}

func (e *Engine) Start(ctx context.Context) error {
	if err := os.MkdirAll(e.basePath, 0755); err != nil {
		return err
	}

	ipcServer := NewIpcServer(e.socketPath, e)
	if err := ipcServer.Start(ctx); err != nil {
		return err
	}

	go e.processUpdates(ctx)

	return nil
}

func (e *Engine) Subscribe(ctx context.Context, command string, flags map[string]string) (string, error) {
	parts := strings.SplitN(command, ".", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid command format: %s", command)
	}
	moduleName := parts[0]

	m, ok := e.modules[moduleName]
	if !ok {
		return "", fmt.Errorf("module not found: %s", moduleName)
	}

	filePath := GenerateStatePath(command, flags)

	e.activeMutex.Lock()
	isActive := e.active[filePath]
	if !isActive {
		e.active[filePath] = true
	}
	e.activeMutex.Unlock()

	if !isActive {
		err := m.Track(ctx, parts[1], flags, e.updates)
		if err != nil {
			e.activeMutex.Lock()
			delete(e.active, filePath)
			e.activeMutex.Unlock()
			return "", err
		}
	}

	return filepath.Join(e.basePath, filePath), nil
}

func (e *Engine) Push(ctx context.Context, command string, flags map[string]string, value string) error {
	filePath := GenerateStatePath(command, flags)
	fmt.Printf("Engine.Push called! Path: %s, Value: %s\n", filePath, value)
	select {
	case e.updates <- StateUpdate{Path: filePath, Value: value}:
	default:
	}
	return nil
}

func (e *Engine) processUpdates(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-e.updates:
			fullPath := filepath.Join(e.basePath, update.Path)
			_ = os.MkdirAll(filepath.Dir(fullPath), 0755)

			// Write safely to a temp file then rename for atomic fsnotify events
			tmpPath := fullPath + ".tmp"
			if err := os.WriteFile(tmpPath, []byte(update.Value+"\n"), 0644); err == nil {
				_ = os.Rename(tmpPath, fullPath)
			}
		}
	}
}

func (e *Engine) Wakeup() {
	for _, m := range e.modules {
		m.Wakeup()
	}
}

func (e *Engine) GetSocketPath() string {
	return e.socketPath
}

func (e *Engine) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no arguments provided")
	}

	// We route Execution to the appropriate module.
	moduleName := args[0]
	m, ok := e.modules[moduleName]
	if !ok {
		return fmt.Errorf("module not found: %s", moduleName)
	}

	// If the module implements an Executor interface, call it!
	type Executor interface {
		Execute(ctx context.Context, args []string) error
	}

	if exec, ok := m.(Executor); ok {
		return exec.Execute(ctx, args[1:])
	}

	return fmt.Errorf("module %s does not support execution", moduleName)
}
