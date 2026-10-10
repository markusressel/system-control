# AGENTS.md

system-control: Go CLI (cobra + viper) that wraps Linux system tools (nmcli, wpctl/pw-dump, redshift, xrandr, BlueZ
over D-Bus, sysfs) behind one command tree. An optional systemd user daemon makes frequent commands fast and serves
`--watch` for status bars (polybar).

## Commands

```bash
make build            # bin/system-control and the C client bin/sysctrl (needs gcc)
make test             # all tests, verbose, cache cleared
go test ./internal/audio/pipewire/... -run TestParsePwDump -v   # single test
gofmt -l .            # must print nothing
go vet ./...
go test -race ./...   # what CI runs, with the two lines above
make deploy           # sudo install to /usr/bin and enable the user unit (user action, do not run unasked)
```

Run the three CI checks before reporting a task done.

## Rules

- Never run real external tools or touch the user's real daemon in tests or experiments. Tests fake commands with
  `util.SetRunner` and point sysfs variables (`battery.PowerSupplyBasePath`, `util.DisplayBacklightPath`, ...) to a
  temp dir. Parsers are pure functions with table tests.
- Use `util.ExecCommand` for one-shot tool calls (it enforces a timeout and can be faked). No `log.Fatal`/`os.Exit`
  in code the daemon can reach: return errors.
- No personal values in code (user names, interfaces, sysfs paths): use flags, environment or `configuration`. New
  config keys also go into `system-control.yaml`.
- Persistence files (`internal/persistence`) are replaced atomically, keys must be plain file names.
- Untracked `testdir/` is junk: never commit it. Stage explicitly, e.g.
  `git add -A -- . ':!testdir' ':!internal/persistence/key.sav'`.
- Commit and push only when asked. Commit messages: prose, imperative first line, body explains why. End with
  the `Co-Authored-By` line from the harness.

## Adding or changing a command

1. Put the logic in `internal/<subsystem>/`. `cmd/` only declares flags, arguments, help and calls into it.
2. If the daemon should serve it (frequent, status bar, key binding): register a `daemon.Op` from an `init()` in
   `internal/<subsystem>/ops.go` (name = CLI path, e.g. `audio.volume.inc`) and make the cobra command
   `RunE: commandutil.OpRunE("<op name>")`. One implementation only: never write daemon behavior separately.
3. `Registry.Bind` (in `cmd/daemon.go`) derives flags, positional-arg validation and subcommands from the cobra
   command, so no flag lists in the op. `cmd/ops_test.go` fails if an op has no matching command or a watchable op
   lacks a bool `--watch`.
4. Add table tests for parsing; use `internal/daemon/daemon_test.go` helpers for op behavior.

## Daemon semantics (easy to get wrong)

- **Fallback:** the daemon answers "unhandled" for anything it does not fully understand (unknown flag, `--help`,
  root flags such as `--configuration`). The regular cobra path then runs and is the reference behavior. An op that
  cannot run right now returns `daemon.ErrUnhandled` before changing anything; ops that need environment
  (X11 tools: `DISPLAY`) declare `Op.Requires`, the engine imports it from the user manager or falls back.
  `daemon.Running()` tells daemon and CLI context apart.
- **Never re-run after silence:** once a request was sent, run locally only on an explicit "unhandled" answer. A
  silent daemon may already have executed a mutation.
- **State:** the daemon owns published state, clients never push values. `--watch` subscribes to the engine store;
  refresh comes from event sources (`Engine.AddSource`, e.g. PipeWire monitor) and `Op.Invalidates`. `Op.Poll` is
  only a fallback (media: `playerctl --follow` pushes status, polling drops to a minute while it runs). A CLI process that changes state itself calls `daemon.Notify(prefix)`.
- **Output flags:** a flag that only changes printing (e.g. `--format`) is listed in `Op.OutputFlags` and applied by
  `Op.Render`, so watchers differing only in it share one query (`display.redshift.brightness`, `.gamma`).
- **Mutations** hold `Op.Lock` (cross-process, shared with the CLI). Waiters beyond `daemon.MaxQueueWait` are dropped
  on purpose (scroll spam): a command waits for its lock at most `MaxQueueWait` in total. An op that queries and
  mutates by argument sets `Op.LockFor` instead of `Op.Lock`. Slow relative steps (inc/dec of redshift) are
  summed with `daemon.Accumulator` and applied in one go instead of dropped, see `cmd/display/redshift/coalesce.go`.
- **Observability:** `system-control daemon status` (`-w` live view, `--json`) shows watchers with their current
  value, per-op calls/latency/drops/failures and source health. Add a source's health with `Engine.AddStatus`.
  The structured form is `StatusInfo` (`STATUS json`), `RenderStatus` draws it.
- **Configuration is a snapshot:** read it with `configuration.Get()` for every use, never keep it. `Get()` reloads
  the file when its mtime changed (checked once per second), a broken file keeps the old values. SIGHUP forces it.
  Poll intervals are asked for again after every tick.
- **Two clients:** the Go client (`cmd.Execute` calls `daemon.TryRemote` first) and the C client
  `cmd/sysctrl/main.c`. Protocol lives in `internal/daemon/ipc.go`; change both clients together,
  `internal/daemon/sysctrl_test.go` covers the C one.
- **Test daemon:** set `SYSTEM_CONTROL_RUNTIME_DIR` (not `XDG_RUNTIME_DIR`, PipeWire needs it) and use own binaries.
  Never `pkill -f`; find the test daemon by its environment so the user's real one is not disturbed.

## Layout pointers

Only what is not obvious from the tree:

- `internal/daemon/` — op registry, engine/state store, unix socket server and client, locks.
- `internal/commandutil/` — `OpRunE`, glue between cobra and ops.
- `internal/audio/pipewire/` — parses `pw-dump` JSON; fixtures in `test/pipewire/`.
- `internal/bluetooth/` — BlueZ D-Bus only (one `GetManagedObjects` per query, signals for `--watch`).
- `internal/wifi/` — nmcli in terse mode, typed structs.
- `internal/battery/`, `internal/disk/` — sysfs batteries (incl. Logitech HID++), smartctl/nvme.
- `internal/util/` — `ExecCommand`, sysfs file I/O, X11 tool wrappers (xrandr, xinput, wmctrl). X11 is assumed.
- `cmd/match_test.go` — compares daemon command line parsing with cobra's; keep it green when touching flags.

## Configuration

YAML, searched in `~/.system-control.yaml`, `~/.config/system-control/`, `/etc/system-control/` (example:
`system-control.yaml`). Sections, all optional: `redshift` (brightness/color temperature/gamma ranges, transition
duration), `daemon.pollIntervals`, `hotspot`, `fan.policyFileGlob`.

## Known gaps (do not "fix" unasked)

- X11 only, no Wayland backends.
- `mouse battery` prints an empty line with exit 0 when no mouse exists (status bars depend on it).
- `display redshift update/reset`, `redshift` (root), `bluetooth` pairing/scan, `network`, `shutdown`/`restart`,
  `video`, `config` are plain CLI commands on purpose (slow, rare or interactive).
