#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/socket.h>
#include <sys/un.h>

int main(int argc, char *argv[]) {
    if (argc < 2) {
        return 0;
    }

    uid_t uid = getuid();
    char socket_path[256];
    snprintf(socket_path, sizeof(socket_path), "/run/user/%d/system-control/ipc.sock", uid);

    int sock = socket(AF_UNIX, SOCK_STREAM, 0);
    if (sock < 0) {
        perror("socket");
        return 1;
    }

    struct timeval tv;
    tv.tv_sec = 2;
    tv.tv_usec = 0;
    setsockopt(sock, SOL_SOCKET, SO_RCVTIMEO, (const char*)&tv, sizeof(tv));
    setsockopt(sock, SOL_SOCKET, SO_SNDTIMEO, (const char*)&tv, sizeof(tv));

    struct sockaddr_un addr;
    memset(&addr, 0, sizeof(addr));
    addr.sun_family = AF_UNIX;
    strncpy(addr.sun_path, socket_path, sizeof(addr.sun_path) - 1);

    if (connect(sock, (struct sockaddr*)&addr, sizeof(addr)) < 0) {
        // Fallback to executing system-control directly if daemon is not running
        execvp("system-control", argv);
        perror("execvp");
        return 1;
    }

    // Send arguments separated by null bytes. Start with a non-'{' character (like 'E' for execute)
    // so the Go daemon knows it's a raw arg stream, not JSON.
    char prefix = 'E';
    write(sock, &prefix, 1);

    for (int i = 1; i < argc; i++) {
        write(sock, argv[i], strlen(argv[i]));
        char null_byte = '\0';
        write(sock, &null_byte, 1);
    }

    // Shut down the write half to signal EOF
    shutdown(sock, SHUT_WR);

    // Read response (if any) and print to stdout
    char buf[1024];
    ssize_t n;
    int is_error = 0;
    while ((n = read(sock, buf, sizeof(buf) - 1)) > 0) {
        buf[n] = '\0';
        if (strstr(buf, "\"status\":\"error\"") != NULL) {
            is_error = 1;
        }
    }
    close(sock);

    if (is_error) {
        execvp("system-control", argv);
        perror("execvp");
        return 1;
    }

    return 0;
}
