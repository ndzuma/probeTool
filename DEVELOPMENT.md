# probeTool Development

Technical reference for probeTool.

---

## Project Structure

```text
probeTool/
├── cmd/                     # CLI commands (Cobra)
│   ├── root.go              # Main scan command
│   ├── serve.go             # Dashboard server
│   ├── tray.go              # System tray (includes server)
│   └── config.go            # Configuration
├── internal/
│   ├── config/              # Provider management
│   ├── db/                  # SQLite operations
│   ├── prober/              # Scan execution
│   ├── server/              # HTTP routes
│   ├── tray/                # System tray logic
│   ├── paths/               # OS-specific paths
│   └── version/             # Version info
├── agent/
│   ├── probe-runner.js      # AI scanner
│   └── prompts.js           # AI prompts
└── web/                     # Next.js dashboard
    ├── app/                 # Pages
    └── components/          # React components
```

---

## Component Integration

```text
User
│
├─→ CLI (Go)
│   ├─→ Agent (Node.js) ──→ AI API
│   │                        ↓
│   │                   Analysis Result
│   │                        ↓
│   └─→ SQLite DB ←───────────┘
│
└─→ HTTP Server (Go) ──→ Next.js Dashboard
                          ↓
                       SQLite DB
```

**`probe tray`:**
- Starts system tray icon
- Spawns `probe serve` as subprocess
- Provides menu: Open Dashboard, Restart Server, Quit

**`probe serve`:**
- Starts HTTP server on port 37330
- Serves Next.js static build
- Provides API endpoints

**`probe --full`/`--quick`:**
- Spawns agent with target path
- Agent calls AI for analysis
- Saves results to database

---

## Scan Flow

```text
User: probe --full
  ↓
CLI validates config + API key
  ↓
CLI creates probe record in DB (status: "running")
  ↓
CLI spawns: node agent/probe-runner.js --target=...
  ↓
Agent reads files → sends to AI → receives findings
  ↓
Agent writes markdown report
  ↓
CLI parses report → updates DB (status: "completed")
  ↓
User views at http://localhost:37330/probes/{id}
```

---

## Tray Integration

```text
User: probe tray
  ↓
Tray registers system tray icon
  ↓
Tray spawns: probe serve --quiet
  ↓
Tray polls /api/health until ready
  ↓
Tray tooltip: "probeTool - Running"
  ↓
User clicks "Open Dashboard" → opens browser
  ↓
User clicks "Quit" → sends SIGINT → server stops → tray exits
```

---

## HTTP Request Flow

```text
Browser: GET /probes/abc123
  ↓
Go HTTP Server
  ↓
Is /api/* ? ──YES──→ API Handler → Query DB → JSON response
  │
  NO
  ↓
Serve Next.js static file from web/out/
```

---

## API Endpoints

Base URL: `http://localhost:37330`

### `GET /api/probes`

List all scans.

**Response:**

```json
{
  "probes": [{
    "id": "2026-02-20-150405-full",
    "type": "full",
    "status": "completed",
    "created_at": "2026-02-20T15:04:05Z",
    "findings_count": 12
  }]
}
```

### `GET /api/probes/:id`

Get scan details with markdown report.

**Response:**

```json
{
  "probe": {
    "id": "2026-02-20-150405-full",
    "type": "full",
    "status": "completed"
  },
  "report": "# Security Assessment\n...",
  "findings_count": 12
}
```

### `GET /api/findings/:probe_id`

Get findings for a scan.

**Query params:** `?severity=critical`

**Response:**

```json
{
  "findings": [{
    "id": "...",
    "probe_id": "...",
    "text": "SQL injection in auth.go:42",
    "severity": "critical"
  }]
}
```

### `GET /api/config`

Get current configuration.

**Response:**

```json
{
  "providers": ["openrouter"],
  "default": "openrouter",
  "current_model": "anthropic/claude-3.5-haiku"
}
```

### `GET /api/version`

Get version information.

**Response:**

```json
{
  "version": "0.1.0",
  "commit": "abc123d",
  "buildDate": "2026-02-20T15:04:05Z",
  "platform": "darwin/arm64"
}
```

### `GET /api/file-tree/:probe_id`

Get file tree for scanned directory.

---

## Database Schema

File: `~/.../probeTool/probes/probes.db`

```sql
CREATE TABLE probes (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    target_path TEXT NOT NULL,
    output_path TEXT NOT NULL,
    status TEXT DEFAULT 'running',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE findings (
    id TEXT PRIMARY KEY,
    probe_id TEXT NOT NULL,
    text TEXT NOT NULL,
    severity TEXT NOT NULL,
    FOREIGN KEY(probe_id) REFERENCES probes(id)
);
```

---

## Configuration

### Locations

- **macOS:** `~/Library/Application Support/probeTool/`
- **Linux:** `~/.config/probeTool/`
- **Windows:** `%APPDATA%\probeTool\`

### Format

```json
{
  "providers": {
    "openrouter": {
      "name": "openrouter",
      "base_url": "https://openrouter.ai/api/v1",
      "api_key": "sk-...",
      "models": ["anthropic/claude-3.5-haiku"],
      "default_model": "anthropic/claude-3.5-haiku"
    }
  },
  "default": "openrouter"
}
```

---

## Versioning

Uses Semantic Versioning:

- **Major** (`v2.0.0`): Breaking changes
- **Minor** (`v0.2.0`): New features
- **Patch** (`v0.1.1`): Bug fixes

Version injected at build time via ldflags:

```go
-X github.com/ndzuma/probeTool/internal/version.Version=v0.1.0
```

---

## Testing Overview

```bash
# Unit tests
go test ./...

# Agent tests
cd agent && npm test

# Integration
make probe
./probe --quick
./probe serve

# Local release test
goreleaser build --snapshot --clean --single-target
```

See code in `*_test.go` files for specific test cases.

---

## Troubleshooting

**Port in use:**

```bash
lsof -ti:37330 | xargs kill -9
```

**Database locked:**

```bash
pkill probe
rm ~/Library/Application\ Support/probeTool/probes/probes.db
```

**Agent missing:**

```bash
probe setup
```

---

For development details, see [DEVELOPMENT.md](DEVELOPMENT.md).