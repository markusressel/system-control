# AGENTS.md

This file provides guidance to AI Agents when working with code in this repository.

## Commands

```bash
make build       # Compile binary to bin/system-control
make test        # Run all tests (verbose)
make deploy      # Install to /usr/bin/ (requires sudo)
make deploy-custom  # Install to ~/.custom/bin/
make clean       # Remove build artifacts
```

Run a single test:

```bash
go test ./internal/audio/pipewire/... -run TestParsePwDump -v
```

## Architecture

**system-control** is a Go CLI tool (cobra + viper) that wraps Linux system utilities (`bluetoothctl`, `nmcli`, `wpctl`/
`pw-dump`, `redshift`, etc.) into a unified interface.

### Key structural patterns

- **`cmd/`** — Cobra command definitions, organized hierarchically by subsystem (e.g., `cmd/audio/volume/`,
  `cmd/display/redshift/`). Each subdirectory corresponds to a command group.
- **`internal/`** — Business logic, isolated from CLI layer:
    - `internal/audio/pipewire/` — PipeWire integration: parses `pw-dump` JSON, manages sinks/devices/profiles
    - `internal/audio/pulseaudio/` — Legacy PulseAudio fallback
    - `internal/bluetooth/` — Wraps `bluetoothctl`, parses its output, fuzzy-matches device names
    - `internal/wifi/` — Wraps `nmcli`, parses output into typed structs
    - `internal/configuration/` — Viper-based config; searches `~/.system-control.yaml`, `~/.config/system-control/`,
      `/etc/system-control/`
    - `internal/persistence/` — Key-value store backed by `~/.config/system-control/persistence/*.sav` files (integers,
      floats, or JSON structs); uses file locking (`gofrs/flock`) to prevent race conditions
    - `internal/daemon/` — Optional background daemon: op registry, state store for `--watch`, unix socket server
      and client (see "Daemon" below)
    - `internal/commandutil/` — Glue between cobra commands and ops (`OpRunE`), locking helpers
    - `internal/util/` — Shared helpers: `ExecCommand`/`ExecCommandAndFork` wrappers, file I/O for sysfs paths,
      string/math utilities

### Daemon

`system-control daemon` (systemd user unit) makes frequent commands fast and serves `--watch`. It is optional:
every command also works on its own, and that in-process path is the reference behavior.

- **One implementation per operation.** A command that the daemon can serve is a `daemon.Op` (name = CLI path, e.g.
  `audio.volume.inc`), registered from an `init()` (`internal/audio/ops.go`). The cobra command declares
  flags, arguments and help and runs `commandutil.OpRunE("<op name>")`. Do not implement daemon behavior twice.
- **Cobra is the only declaration of the CLI surface.** `Registry.Bind` (called by `cmd/daemon.go`) derives flags,
  argument validation and sub commands of an op from its cobra command, so the daemon parses raw command lines like
  cobra. Adding an op needs only the `daemon.Register` call; `cmd/ops_test.go` checks that every op binds and that
  lookalike command lines are left to cobra.
- **Ops that can't run in the daemon right now** return `daemon.ErrUnhandled` before changing anything (e.g. no
  `DISPLAY`), the client then runs the command itself. `daemon.Running()` tells the two execution contexts apart.
- **Configuration in the daemon is a snapshot**: read it with `configuration.Get()`, and call
  `configuration.Reload()` before using values that users edit (see the redshift ops).
- **Request flow**: `cmd.Execute()` first calls `daemon.TryRemote(os.Args[1:])`. The daemon matches the command line
  to an op and answers with output and exit code; for anything it does not fully understand (unknown flag, `--help`,
  set-flags that mutate...) it answers "unhandled" and the regular cobra path runs. `cmd/sysctrl/main.c` is the same
  client without the Go runtime (~1 ms, ~1.5 MB instead of ~3 ms, ~10 MB), which matters for key bindings and
  for the many long-lived `--watch` processes of a status bar. Protocol: `internal/daemon/ipc.go`.
  After a request was sent a command is only run locally on an explicit "unhandled" answer, never because the
  daemon was silent (it may have executed a mutation already).
- **State**: `--watch` subscribes to the engine's store. The daemon owns all published state, clients never push
  values. Event sources (`Engine.AddSource`, e.g. the PipeWire monitor) and `Op.Invalidates` refresh watched ops;
  `Op.Poll` is only a fallback for sources without events. A process that changes state on its own (cobra path)
  calls `daemon.Notify(prefix)`.
- **Mutations** hold `Op.Lock`, a cross-process lock shared by daemon and CLI. Commands waiting longer than
  `daemon.MaxQueueWait` are dropped (scroll spam).
- A second daemon for testing: set `SYSTEM_CONTROL_RUNTIME_DIR` (not `XDG_RUNTIME_DIR`, PipeWire needs that).

### Request lifecycle

```
main() → cmd.Execute()
  → cobra.OnInitialize: configuration.InitConfig() sets viper defaults
  → RootCmd.PersistentPreRunE: DetectAndReadConfigFile() → LoadConfig() → Validate()
  → Subcommand handler
      → internal package (bluetoothctl/nmcli/wpctl/sysfs)
      → persistence.Save/Read for stateful operations
      → stdout output
```

### Configuration

YAML config (example in `system-control.yaml`). The `internal/configuration.CurrentConfig` struct is populated at
startup and used by all commands. Redshift is the primary configurable subsystem (brightness range, color temp range,
transition duration).

### External tool dependencies

Commands shell out to: `bluetoothctl`, `nmcli`, `pw-dump`, `pw-cli`, `wpctl`, `pw-metadata`, `redshift`. The sysfs paths
`/sys/class/backlight/` and `/sys/class/power_supply/` are accessed directly for backlight and battery.
