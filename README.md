# Local Webhook Time Machine (`wtm`)

[![CI](https://github.com/155-Sukreeth/webhook-time-machine/actions/workflows/ci.yml/badge.svg)](https://github.com/155-Sukreeth/webhook-time-machine/actions/workflows/ci.yml)

A lightweight, local-first developer tool built in Go to capture, inspect, edit, and replay HTTP/HTTPS webhook requests with an old-school Unix/Linux desktop visual aesthetic.

---

## Directory Architecture

```
webhook-time-machine/
├── cmd/
│   └── wtm/
│       └── main.go                # CLI entry point
│
├── internal/
│   ├── app/                       # Application bootstrap & lifecycle
│   ├── api/                       # REST API handlers
│   ├── cli/                       # Cobra CLI commands (root, init, start, version)
│   ├── config/                    # Config loader (Flags > ENV > YAML > Defaults)
│   ├── dashboard/                 # HTTP server & static web UI handler
│   ├── logger/                    # Custom structured logger
│   ├── models/                    # Domain data models & API responses
│   ├── proxy/                     # Reverse proxy & request forwarder
│   ├── replay/                    # Replay engine & signature header filter
│   ├── storage/                   # SQLite persistence & migrations
│   └── utils/                     # JSON & helper utilities
│
├── web/                           # Embedded web frontend assets (HTML/CSS/JS)
├── migrations/                    # SQL Schema migration files
├── configs/                       # Example configuration templates
├── scripts/                       # Build & Release scripts
├── docs/                          # Architecture & API documentation
├── .github/workflows/             # CI/CD (golangci-lint, gosec, multi-OS tests, goreleaser)
└── go.mod                         # Go module definitions
```

---

## Quick Start

### 1. Initialize Configuration
```bash
wtm init
```
Generates a `.wtm.yaml` configuration file:
```yaml
port: 8080
ui_port: 8081
forward_url: http://localhost:3000
db_path: ./wtm.db
log_level: info
strip_signatures: true
```

### 2. Start Proxy & UI Dashboard
```bash
wtm start --forward-to http://localhost:3000 --port 8080 --ui-port 8081
```

- Configure external webhooks to target `http://localhost:8080`
- Open your browser at `http://localhost:8081` to view and interact with the **Webhook Time Machine** desktop dashboard!

### Development & Local Tooling Setup

Install local CI quality tooling binaries into your `$GOPATH/bin`:

```bash
# Install Task runner
go install github.com/go-task/task/v3/cmd/task@latest

# Install linters & security analyzers
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
```

### Task Commands (using `task` / `go-task`)
- `task build`: Compiles `bin/wtm` binary.
- `task test`: Runs unit test suite with coverage profile.
- `task lint`: Runs `golangci-lint`.
- `task sec`: Runs `gosec` AST security scanner.
- `task vulndb`: Runs `govulncheck` vulnerability checker.
- `task ci`: Executes test, lint, sec, and vulndb targets sequentially.
- `task clean`: Cleans build artifacts and test outputs.

---

## License
MIT License
