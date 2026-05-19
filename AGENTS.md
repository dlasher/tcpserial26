# AGENTS.md — tcpserial

TCP-to-serial bridge replacing `socat TCP-LISTEN:2000,reuseaddr,fork FILE:/dev/ttyUSB0,raw,echo=0`.
Designed for Z-Wave JS UI. Z-Wave JS connects via `tcp://<host>:<port>`.

## Build & Test

- **Build**: `go build -o tcpserial ./cmd/tcpserial`
- **Test**: `go test ./...` (36 table-driven tests in `internal/config/` and `internal/bridge/`)
- **Release build**: `go build -ldflags "-X main.version=X -X main.commit=$(git rev-parse HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" ./cmd/tcpserial`

`cmd/tcpserial/main.go` has **no tests** — test coverage is in internal packages only.
`internal/serialport/` has **no tests** (requires real hardware).

## Architecture

Single Go binary, 3 internal packages:

- `internal/config/` — `Config` struct, defaults (port 2000, baud 115200, infinite retries), validation
- `internal/serialport/` — device detection (path, USB VID/PID, auto-first-port), typed error handling
- `internal/bridge/` — TCP listener, bidirectional `io.CopyBuffer`, IP whitelist (CIDR), systemd notify + socket activation

Entrypoint: `cmd/tcpserial/main.go` — cobra CLI, viper config (YAML + env vars `TCPSERIAL_*`), signal handling (SIGINT/SIGTERM), reconnect loop.

## Configuration Sources (in priority order)

1. CLI flags (cobra)
2. Env vars `TCPSERIAL_PORT`, `TCPSERIAL_DEVICE`, etc. (viper automatic)
3. YAML config from `/etc/tcpserial/tcpserial.yaml`, `$HOME/.tcpserial/tcpserial.yaml`, `./tcpserial.yaml`

Both VID and PID must be set together for USB auto-detection (`Validate()` enforces this).

## Service Management

systemd files in `systemd/`:
- `tcpserial.service` + `tcpserial.socket` — standard socket activation
- `hardened_tcpserial.service` — restricted sandbox, serial user group

## Key Details

- **MaxRetries=0** means infinite retries (not zero retries)
- **`version`/`commit`/`date`** vars in main.go set via ldflags at build time, default to "dev"/"unknown"
- **BufferSize** default 512000 bytes (max 10MB), **ReadBuffers** default 15 (max 32)
- **No CGO** needed — pure Go serial via `go.bug.st/serial`
- **macOS USB enumeration requires CGO** — cross-compile to darwin skips it
- `socat.command.sh` is the original script being replaced — don't delete
- Built binary `tcpserial` is tracked in repo
