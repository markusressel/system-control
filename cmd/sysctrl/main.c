// sysctrl: minimal client for the system-control daemon.
//
// Process startup is what makes a Go binary the wrong tool for keybindings and status bars:
// starting `system-control` costs a few milliseconds and ~10 MB of resident memory per
// invocation, and a `--watch` of every bar widget keeps one such process alive.
// This client does nothing but forward the command line to the daemon and relay its answer.
//
// The daemon decides whether it handles a command. If it does not, or is not running,
// the command is run by the regular `system-control` binary.
//
// Protocol (see internal/daemon/ipc.go):
//   -> "ARGV <n>\n" followed by n times "<len>\n<bytes>"
//   <- "O <line>\n" stdout, "E <line>\n" stderr, "X <code>\n" exit code,
//      "U\n" not handled, "W\n" the request is a watch (streams until we disconnect)
//
// Once the request is sent, the command is only run locally on "U". Anything else could
// apply a mutation twice.

#include <errno.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/time.h>
#include <sys/un.h>
#include <unistd.h>

#define FALLBACK_BINARY "system-control"
#define NO_DAEMON_ENV "SYSTEM_CONTROL_NO_DAEMON"
// a stuck daemon must not hang a keybinding forever, watches are exempt once announced
#define ANSWER_TIMEOUT_SEC 5

static void run_fallback(char *argv[]) {
    // the daemon was asked already, don't let the binary ask again
    setenv(NO_DAEMON_ENV, "1", 1);
    argv[0] = FALLBACK_BINARY;
    execvp(FALLBACK_BINARY, argv);
    perror("execvp " FALLBACK_BINARY);
    exit(127);
}

static int write_all(int fd, const char *buf, size_t len) {
    while (len > 0) {
        ssize_t n = write(fd, buf, len);
        if (n < 0) {
            if (errno == EINTR) continue;
            return -1;
        }
        buf += n;
        len -= (size_t) n;
    }
    return 0;
}

static int send_request(int sock, int argc, char *argv[]) {
    char header[64];
    int n = snprintf(header, sizeof(header), "ARGV %d\n", argc - 1);
    if (write_all(sock, header, (size_t) n) < 0) return -1;

    for (int i = 1; i < argc; i++) {
        size_t len = strlen(argv[i]);
        n = snprintf(header, sizeof(header), "%zu\n", len);
        if (write_all(sock, header, (size_t) n) < 0) return -1;
        if (write_all(sock, argv[i], len) < 0) return -1;
    }
    return 0;
}

int main(int argc, char *argv[]) {
    if (argc < 2) {
        return 0;
    }

    signal(SIGPIPE, SIG_IGN);

    struct sockaddr_un addr;
    memset(&addr, 0, sizeof(addr));
    addr.sun_family = AF_UNIX;

    // same lookup as RuntimeDir() in internal/daemon/runtime.go
    const char *override_dir = getenv("SYSTEM_CONTROL_RUNTIME_DIR");
    const char *runtime_dir = getenv("XDG_RUNTIME_DIR");
    int path_len;
    if (override_dir != NULL && override_dir[0] != '\0') {
        path_len = snprintf(addr.sun_path, sizeof(addr.sun_path), "%s/ipc.sock", override_dir);
    } else if (runtime_dir != NULL && runtime_dir[0] != '\0') {
        path_len = snprintf(addr.sun_path, sizeof(addr.sun_path), "%s/system-control/ipc.sock", runtime_dir);
    } else {
        path_len = snprintf(addr.sun_path, sizeof(addr.sun_path), "/run/user/%d/system-control/ipc.sock", (int) getuid());
    }
    if (path_len < 0 || (size_t) path_len >= sizeof(addr.sun_path)) {
        run_fallback(argv);
    }

    int sock = socket(AF_UNIX, SOCK_STREAM, 0);
    if (sock < 0) {
        run_fallback(argv);
    }

    if (connect(sock, (struct sockaddr *) &addr, sizeof(addr)) < 0) {
        run_fallback(argv); // no daemon
    }

    struct timeval timeout = {.tv_sec = ANSWER_TIMEOUT_SEC, .tv_usec = 0};
    setsockopt(sock, SOL_SOCKET, SO_RCVTIMEO, &timeout, sizeof(timeout));

    if (send_request(sock, argc, argv) < 0) {
        run_fallback(argv);
    }

    FILE *in = fdopen(sock, "r");
    if (in == NULL) {
        run_fallback(argv);
    }

    char *line = NULL;
    size_t cap = 0;

    while (getline(&line, &cap, in) > 0) {
        if (line[0] == 'U') {
            run_fallback(argv); // not handled by the daemon
        }

        if (line[0] == 'W') {
            struct timeval none = {.tv_sec = 0, .tv_usec = 0};
            setsockopt(sock, SOL_SOCKET, SO_RCVTIMEO, &none, sizeof(none));
            continue;
        }

        if (line[0] == 'O' || line[0] == 'E') {
            FILE *out = line[0] == 'O' ? stdout : stderr;
            // frames are "<kind> <text>\n"
            fputs(line + 2, out);
            fflush(out);
        } else if (line[0] == 'X') {
            return atoi(line + 2);
        }
    }

    // The request is out, the daemon may have executed it already.
    fprintf(stderr, "sysctrl: daemon did not answer\n");
    return 1;
}
