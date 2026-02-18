# Audit Tool

A CLI tool for auditing codebases with a web interface.

## Installation

```bash
go mod tidy
```

## Usage

```bash
# Run a full audit on current directory
audit --full

# Run a quick audit
audit --quick

# Audit a specific path
audit --full --specific /path/to/repo

# Audit changes only
audit --changes
```

## API

```bash
# List all audits
curl http://localhost:3030/api/audits

# Get specific audit
curl http://localhost:3030/api/audits/{id}
```

## Structure

```
audit/
├── main.go
├── cmd/
│   └── root.go
├── internal/
│   ├── auditor/
│   │   └── auditor.go
│   ├── db/
│   │   └── db.go
│   └── server/
│       └── server.go
└── audits/
    └── audits.db
```
