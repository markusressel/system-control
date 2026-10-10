package daemon

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
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
//	STATUS\n            describe the daemon, answered with O frames
//
// The server answers with frames, one per line:
//
//	O <text>\n   a line for stdout
//	E <text>\n   a line for stderr
//	X <code>\n   the command finished with exit code <code> (not sent for watch requests)
//	U\n          request not handled by the daemon, run the command locally
//	W\n          the request is a watch: O frames follow until the client closes the connection
//
// Once a request was sent, the command is only ever run locally if the daemon answers U. A
// missing answer is an error, as the daemon might have executed the command already.
// The format is kept trivial on purpose, the C client (cmd/sysctrl) implements it too.

// ErrUnhandled can be returned by Op.Run, before it changed anything, when the daemon can't
// serve this invocation right now (e.g. its environment lacks what the op needs). The client
// then runs the command itself.
var ErrUnhandled = errors.New("not handled by the daemon")

const (
	maxArgs   = 64
	maxArgLen = 16 * 1024
)

// runningEngine is the engine of this process, if it is a daemon.
var runningEngine atomic.Pointer[Engine]

// Running reports whether this process is the daemon. Op code uses it to tell the two ways it
// can be executed apart, e.g. to refresh configuration that only the daemon holds for long.
func Running() bool {
	return runningEngine.Load() != nil
}

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
	case "NOTIFY":
		s.engine.Refresh(fields[1:]...)
		writeFrames(conn, 'X', "0")

	case "STATUS":
		for _, line := range s.engine.status() {
			writeFrames(conn, 'O', line)
		}
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
	if errors.Is(err, ErrUnhandled) {
		writeFrames(conn, 'U', "")
		return
	}
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

	updates, release, pending := s.engine.store.subscribe(ctx, op, args)
	if updates == nil {
		// only ErrUnhandled ends a subscription before it started
		writeFrames(conn, 'U', "")
		return
	}
	defer release()

	writeFrames(conn, 'W', "")
	if pending != nil {
		// not available yet, values follow once it is
		writeFrames(conn, 'E', pending.Error())
	}

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
	if kind == 'U' || kind == 'W' {
		_, err := w.Write([]byte{kind, '\n'})
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

// answerTimeout is how long the daemon may take to answer a one-shot request.
const answerTimeout = 5 * time.Second

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
	_ = conn.SetDeadline(time.Now().Add(answerTimeout))

	var sb strings.Builder
	fmt.Fprintf(&sb, "ARGV %d\n", len(argv))
	for _, arg := range argv {
		fmt.Fprintf(&sb, "%d\n%s", len(arg), arg)
	}
	if _, err := io.WriteString(conn, sb.String()); err != nil {
		return false, 0
	}

	r := bufio.NewReader(conn)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			// The request is out, the daemon may have executed it. Running it again could apply a
			// mutation twice, so this is an error and not a reason to fall back.
			fmt.Fprintf(stderr, "daemon did not answer: %v\n", err)
			return true, 1
		}

		line = strings.TrimSuffix(line, "\n")
		switch {
		case line == "U":
			return false, 0
		case line == "W":
			_ = conn.SetReadDeadline(time.Time{})
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

// Status asks the running daemon to describe itself.
func Status() ([]string, error) {
	conn, err := net.DialTimeout("unix", SocketPath(), 200*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("daemon is not running (%s): %w", SocketPath(), err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(answerTimeout))

	if _, err := io.WriteString(conn, "STATUS\n"); err != nil {
		return nil, err
	}

	var lines []string
	r := bufio.NewReader(conn)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("daemon did not answer: %w", err)
		}
		line = strings.TrimSuffix(line, "\n")
		switch {
		case strings.HasPrefix(line, "O "):
			lines = append(lines, line[2:])
		case strings.HasPrefix(line, "X "):
			return lines, nil
		case line == "U":
			return nil, errors.New("daemon does not know the status request, it is older than this binary")
		}
	}
}

// Notify tells a running daemon that state behind the given op name prefixes changed.
// It is best effort and silent if no daemon is running.
func Notify(prefixes ...string) {
	if len(prefixes) == 0 {
		return
	}
	// inside the daemon there is nobody to talk to
	if e := runningEngine.Load(); e != nil {
		e.Refresh(prefixes...)
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
