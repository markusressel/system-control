# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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
    - `internal/util/` — Shared helpers: `ExecCommand`/`ExecCommandAndFork` wrappers, file I/O for sysfs paths,
      string/math utilities

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
