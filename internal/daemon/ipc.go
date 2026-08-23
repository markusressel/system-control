package daemon

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type IpcRequest struct {
	Action  string            `json:"action"` // "subscribe", "push", "ping"
	Command string            `json:"command"`
	Flags   map[string]string `json:"flags,omitempty"`
	Value   string            `json:"value,omitempty"`
}

type IpcResponse struct {
	Status string `json:"status"`
	File   string `json:"file,omitempty"`
	Error  string `json:"error,omitempty"`
}

type IpcServer struct {
	socketPath string
	engine     *Engine
	listener   net.Listener
}

func NewIpcServer(socketPath string, engine *Engine) *IpcServer {
	return &IpcServer{
		socketPath: socketPath,
		engine:     engine,
	}
}

func (s *IpcServer) Start(ctx context.Context) error {
	_ = os.Remove(s.socketPath)
	_ = os.MkdirAll(filepath.Dir(s.socketPath), 0755)

	l, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return err
	}
	s.listener = l

	go func() {
		<-ctx.Done()
		_ = l.Close()
	}()

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go s.handleConnection(ctx, conn)
		}
	}()

	return nil
}

func (s *IpcServer) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	peek, err := reader.Peek(1)
	if err != nil {
		return
	}

	if peek[0] == 'E' {
		// Raw Execute request from C client
		_, _ = reader.ReadByte() // consume 'E'
		bytes, err := io.ReadAll(reader)
		if err != nil {
			s.sendError(conn, "failed to read args")
			return
		}
		argsStr := string(bytes)
		args := strings.Split(argsStr, "\x00")
		// Remove trailing empty string from split
		if len(args) > 0 && args[len(args)-1] == "" {
			args = args[:len(args)-1]
		}

		ctx = context.WithValue(ctx, "queuedAt", time.Now())
		err = s.engine.Execute(ctx, args)
		if err != nil {
			s.sendError(conn, err.Error())
		} else {
			_ = json.NewEncoder(conn).Encode(IpcResponse{Status: "ok"})
		}
		return
	}

	var req IpcRequest
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&req); err != nil {
		s.sendError(conn, "invalid request")
		return
	}

	if req.Action == "ping" || req.Command == "ping" {
		s.engine.Wakeup()
		_ = json.NewEncoder(conn).Encode(IpcResponse{Status: "ok"})
		return
	}

	if req.Action == "push" {
		err := s.engine.Push(ctx, req.Command, req.Flags, req.Value)
		if err != nil {
			s.sendError(conn, err.Error())
		} else {
			_ = json.NewEncoder(conn).Encode(IpcResponse{Status: "ok"})
		}
		return
	}

	// Default to subscribe action
	filePath, err := s.engine.Subscribe(ctx, req.Command, req.Flags)
	if err != nil {
		s.sendError(conn, err.Error())
		return
	}

	resp := IpcResponse{
		Status: "ok",
		File:   filePath,
	}
	_ = json.NewEncoder(conn).Encode(resp)
}

func (s *IpcServer) sendError(conn net.Conn, errMsg string) {
	resp := IpcResponse{
		Status: "error",
		Error:  errMsg,
	}
	_ = json.NewEncoder(conn).Encode(resp)
}

// IpcClient handles connecting to the daemon.
type IpcClient struct {
	socketPath string
}

func NewIpcClient(socketPath string) *IpcClient {
	return &IpcClient{socketPath: socketPath}
}

func (c *IpcClient) IsActive() bool {
	_, err := os.Stat(c.socketPath)
	return err == nil
}

func (c *IpcClient) Subscribe(req IpcRequest) (string, error) {
	req.Action = "subscribe"
	conn, err := net.DialTimeout("unix", c.socketPath, 2*time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return "", err
	}

	var resp IpcResponse
	decoder := json.NewDecoder(bufio.NewReader(conn))
	if err := decoder.Decode(&resp); err != nil {
		return "", err
	}

	if resp.Status != "ok" {
		return "", fmt.Errorf("daemon error: %s", resp.Error)
	}

	return resp.File, nil
}

func (c *IpcClient) Push(req IpcRequest) error {
	req.Action = "push"
	conn, err := net.DialTimeout("unix", c.socketPath, 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return err
	}

	var resp IpcResponse
	decoder := json.NewDecoder(bufio.NewReader(conn))
	if err := decoder.Decode(&resp); err != nil {
		return err
	}

	if resp.Status != "ok" {
		return fmt.Errorf("daemon error: %s", resp.Error)
	}
	return nil
}

func (c *IpcClient) Ping() error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	req := IpcRequest{Action: "ping"}
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return err
	}

	var res IpcResponse
	return json.NewDecoder(conn).Decode(&res)
}
