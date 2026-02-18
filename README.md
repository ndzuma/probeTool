# Probe Tool

A CLI tool for probing codebases with a web interface.

## Installation

```bash
go mod tidy
```

## Usage

```bash
# Run a full probe on current directory
probe --full

# Run a quick probe
probe --quick

# Probe a specific path
probe --full --specific /path/to/repo

# Probe changes only
probe --changes
```

## API

```bash
# List all probes
curl http://localhost:3030/api/probes

# Get specific probe
curl http://localhost:3030/api/probes/{id}
```

## Structure

```
probeTool/
├── cmd/probe/main.go          # Entry point
├── cmd/root.go                # Cobra commands
├── internal/
│   ├── prober/prober.go       # Probe logic
│   ├── db/db.go               # SQLite database
│   └── server/server.go       # HTTP server
├── probes/                    # SQLite DB & markdown files
├── probe                      # Binary
├── Makefile
├── go.mod
└── README.md
```
