# tcpserial

TCP-to-serial bridge for Z-Wave JS UI. Replaces `socat TCP-LISTEN:2000,reuseaddr,fork FILE:/dev/ttyUSB0,raw,echo=0` with config, auto-detection, and service management.

## Quick Start

```bash
tcpserial                                                # auto-detect, port 2000
tcpserial --device /dev/ttyUSB0 --baud 115200             # explicit device
tcpserial --vid 10C4 --pid EA60                           # auto-detect by USB IDs
tcpserial --address 192.168.1.100 --port 6638 --whitelist 192.168.1.0/24
```

## Z-Wave JS UI

Set serial port to `tcp://<host>:2000` in Z-Wave JS UI settings.

## CLI Options

```
Flags:
  -a, --address string        Listen address (default "0.0.0.0")
  -p, --port int              Listen port (default 2000)
      --whitelist strings     IP whitelist (CIDR notation, repeatable)
  -d, --device string         Serial device path (empty for auto-detect)
      --vid string            USB Vendor ID for auto-detection (e.g., 10C4)
      --pid string            USB Product ID for auto-detection (e.g., EA60)
  -b, --baud int              Baud rate (default 115200)
      --data-bits int         Data bits (5-8) (default 8)
      --parity string         Parity: none, odd, even, mark, space (default "none")
      --stop-bits string      Stop bits: 1, 1.5, 2 (default "1")
      --read-buffers int      Number of read buffers (1-32) (default 15)
  -s, --buffer-size int       TCP buffer size in bytes (1-10MB) (default 512000)
      --read-timeout duration Socket read timeout (default 30s)
      --write-timeout duration Socket write timeout (default 30s)
      --retry-interval duration Retry interval for reconnection (default 5s)
      --max-retries int       Max retry attempts (0 = infinite) (default 0)
      --log-level string      Log level: debug, info, warn, error (default "info")
      --config string         Config file path
  -h, --help                  Help for tcpserial
  -v, --version               Version for tcpserial
```

## systemd Service

```bash
# Standard with socket activation
sudo cp systemd/tcpserial.service /etc/systemd/system/
sudo cp systemd/tcpserial.socket /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now tcpserial.socket

# Hardened (production)
sudo cp systemd/hardened_tcpserial.service /etc/systemd/system/tcpserial.service
sudo systemctl daemon-reload
sudo systemctl enable --now tcpserial
```

## Configuration

Loads YAML from `/etc/tcpserial/tcpserial.yaml`, `$HOME/.tcpserial/tcpserial.yaml`, or `./tcpserial.yaml`. Env vars with `TCPSERIAL_` prefix also work.

```yaml
address: "0.0.0.0"
port: 2000
device: "/dev/ttyUSB0"
baud: 115200
parity: "none"
stop-bits: "1"
data-bits: 8
buffer-size: 512000
read-buffers: 15
read-timeout: "30s"
write-timeout: "30s"
retry-interval: "5s"
max-retries: 0
log-level: "info"
whitelist:
  - "192.168.1.0/24"
  - "10.0.0.0/8"
```

## Building

```bash
go build -o tcpserial ./cmd/tcpserial

# With version info
go build -ldflags "-X main.version=1.0.0 -X main.commit=$(git rev-parse HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" ./cmd/tcpserial
```

## License

MIT
