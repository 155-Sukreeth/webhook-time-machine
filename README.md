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

---

## License
MIT License
