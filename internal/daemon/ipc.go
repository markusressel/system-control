package daemon

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Wire protocol
//
// A connection carries exactly one request.
//
// Requests are a header line, optionally followed by a body:
//
//	ARGV <n>\n          followed by n arguments, each encoded as "<len>\n<len bytes>"
//	NOTIFY <prefix>...\n  state behind the given op name prefixes changed, re-query watchers
//	PING\n
//
// The server answers with frames, one per line:
//
//	O <text>\n   a line for stdout
//	E <text>\n   a line for stderr
//	X <code>\n   the command finished with exit code <code> (not sent for watch requests)
//	U\n          request not handled by the daemon, run the command locally
//
// A watch request keeps the connection open and streams O frames until the client closes it.
// The format is kept trivial on purpose, the C client (cmd/sysctrl) implements it too.

const (
	maxArgs   = 64
	maxArgLen = 16 * 1024
)

type ipcServer struct {
	engine   *Engine
	listener net.Listener
}

func (s *ipcServer) start(ctx context.Context, path string) error {
	l, err := net.Listen("unix", path)
	if err != nil {
		return err
	}
	if err := os.Chmod(path, 0600); err != nil {
		_ = l.Close()
		return err
	}
	s.listener = l

	go func() {
		<-ctx.Done()
		_ = l.Close() // also unlinks the socket
	}()

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go s.handle(ctx, conn)
		}
	}()

	return nil
}

func (s *ipcServer) handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic while handling request: %v", r)
			writeFrames(conn, 'E', "internal error")
			writeFrames(conn, 'X', "1")
		}
	}()

	receivedAt := time.Now()
	r := bufio.NewReader(conn)

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	header, err := r.ReadString('\n')
	if err != nil {
		return
	}
	fields := strings.Fields(header)
	if len(fields) == 0 {
		return
	}

	switch fields[0] {
	case "PING":
		writeFrames(conn, 'X', "0")

	case "NOTIFY":
		s.engine.Refresh(fields[1:]...)
		writeFrames(conn, 'X', "0")

	case "ARGV":
		n, err := strconv.Atoi(strings.Join(fields[1:], ""))
		if err != nil || n < 0 || n > maxArgs {
			writeFrames(conn, 'U', "")
			return
		}
		argv, err := readArgs(r, n)
		if err != nil {
			return
		}
		_ = conn.SetReadDeadline(time.Time{})
		s.handleArgv(ctx, conn, r, argv, receivedAt)

	default:
		writeFrames(conn, 'U', "")
	}
}

func (s *ipcServer) handleArgv(ctx context.Context, conn net.Conn, r *bufio.Reader, argv []string, receivedAt time.Time) {
	op, args, ok := s.engine.registry.MatchArgv(argv)
	if !ok {
		writeFrames(conn, 'U', "")
		return
	}

	if args.Watch {
		s.serveWatch(ctx, conn, r, op, args)
		return
	}

	lines, err := op.Execute(ctx, args, receivedAt)
	if err != nil {
		writeFrames(conn, 'E', err.Error())
		writeFrames(conn, 'X', "1")
		return
	}

	if len(op.Invalidates) > 0 {
		s.engine.Refresh(op.Invalidates...)
	}

	for _, line := range lines {
		writeFrames(conn, 'O', line)
	}
	writeFrames(conn, 'X', "0")
}

func (s *ipcServer) serveWatch(ctx context.Context, conn net.Conn, r *bufio.Reader, op *Op, args Args) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	updates, release, err := s.engine.store.subscribe(ctx, op, args)
	if err != nil {
		writeFrames(conn, 'E', err.Error())
		writeFrames(conn, 'X', "1")
		return
	}
	defer release()

	// The client sends nothing after its request, so a read only returns once it went away.
	go func() {
		_, _ = io.Copy(io.Discard, r)
		cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case lines := <-updates:
			for _, line := range lines {
				if err := writeFrames(conn, 'O', line); err != nil {
					return
				}
			}
		}
	}
}

func readArgs(r *bufio.Reader, n int) ([]string, error) {
	argv := make([]string, 0, n)
	for i := 0; i < n; i++ {
		lenLine, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		size, err := strconv.Atoi(strings.TrimSpace(lenLine))
		if err != nil || size < 0 || size > maxArgLen {
			return nil, fmt.Errorf("invalid argument length %q", lenLine)
		}
		buf := make([]byte, size)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		argv = append(argv, string(buf))
	}
	return argv, nil
}

// writeFrames writes text as one frame per line.
func writeFrames(w io.Writer, kind byte, text string) error {
	if kind == 'U' {
		_, err := w.Write([]byte("U\n"))
		return err
	}

	var sb strings.Builder
	for _, line := range strings.Split(text, "\n") {
		sb.WriteByte(kind)
		sb.WriteByte(' ')
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
	_, err := io.WriteString(w, sb.String())
	return err
}

// EnvNoDaemon disables forwarding to the daemon. It is set by the C client before it
// hands over to this binary, so the request is not sent twice.
const EnvNoDaemon = "SYSTEM_CONTROL_NO_DAEMON"

// TryRemote lets the daemon run a command line. handled is false if the daemon is not
// reachable or does not handle the command, in which case nothing was written to stdout/stderr
// and the caller must run the command itself.
func TryRemote(argv []string, stdout, stderr io.Writer) (handled bool, exitCode int) {
	if os.Getenv(EnvNoDaemon) != "" {
		return false, 0
	}

	conn, err := net.DialTimeout("unix", SocketPath(), 200*time.Millisecond)
	if err != nil {
		return false, 0
	}
	defer conn.Close()

	var sb strings.Builder
	fmt.Fprintf(&sb, "ARGV %d\n", len(argv))
	for _, arg := range argv {
		fmt.Fprintf(&sb, "%d\n%s", len(arg), arg)
	}
	if _, err := io.WriteString(conn, sb.String()); err != nil {
		return false, 0
	}

	r := bufio.NewReader(conn)
	gotFrame := false
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if !gotFrame {
				return false, 0
			}
			return true, 1 // the daemon went away mid-command (or a watch was cut off)
		}
		gotFrame = true

		line = strings.TrimSuffix(line, "\n")
		switch {
		case line == "U":
			return false, 0
		case strings.HasPrefix(line, "O "):
			fmt.Fprintln(stdout, line[2:])
		case strings.HasPrefix(line, "E "):
			fmt.Fprintln(stderr, line[2:])
		case strings.HasPrefix(line, "X "):
			code, _ := strconv.Atoi(line[2:])
			return true, code
		}
	}
}

// Notify tells a running daemon that state behind the given op name prefixes changed.
// It is best effort and silent if no daemon is running.
func Notify(prefixes ...string) {
	if len(prefixes) == 0 {
		return
	}
	conn, err := net.DialTimeout("unix", SocketPath(), 100*time.Millisecond)
	if err != nil {
		return
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(time.Second))

	if _, err := fmt.Fprintf(conn, "NOTIFY %s\n", strings.Join(prefixes, " ")); err != nil {
		return
	}
	_, _ = bufio.NewReader(conn).ReadString('\n')
}
