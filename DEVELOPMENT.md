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
│   ├── stop.go              # Stop daemon
│   ├── status.go            # Check daemon status
│   ├── setup.go             # Install agent files
│   ├── config.go            # Configuration
│   ├── daemon_unix.go       # Unix daemon spawning
│   ├── daemon_windows.go    # Windows daemon spawning
│   └── ...
├── internal/
│   ├── config/              # Provider management
│   ├── db/                  # SQLite operations
│   ├── prober/              # Scan execution
│   ├── server/              # HTTP routes
│   ├── tray/                # System tray logic
│   ├── paths/               # OS-specific paths
│   ├── process/             # PID file management
│   ├── updater/             # Self-update functionality
│   ├── wsl/                 # WSL detection
│   ├── agent/               # Bundled agent extraction
│   ├── runtime/             # Bundled runtime extraction
│   └── version/             # Version info
├── agent/
│   ├── probe-runner.js      # AI scanner
│   ├── prompts.js           # AI prompts
│   ├── package.json         # Dependencies
│   └── .claude/             # AI context
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
Check if running on WSL
  ├─ YES: Prompt user for tray preference (cache choice)
  └─ NO: Continue with tray
  ↓
Tray registers system tray icon
  ↓
Tray spawns: probe serve --quiet
  ↓
Tray polls /api/health until ready
  ↓
Tray tooltip: "probeTool - Running"
  ↓
Tray starts update polling (every 4 hours)
  ↓
User clicks "Open Dashboard" → opens browser
  ↓
User clicks "Update Available" → downloads and installs update
  ↓
User clicks "Quit" → sends SIGINT → server stops → tray exits
```

---

## Self-Update Flow

```text
User: probe update
  ↓
Check update cache
  ├─ Recent (< 24h): Use cached result
  └─ Stale/missing: Query GitHub API
  ↓
Compare versions
  ├─ Same: "Already on latest version"
  └─ Newer available: Show release notes
  ↓
User confirms (or use -y flag)
  ↓
Download tar.gz from GitHub releases
  ↓
Extract binary to temp directory
  ↓
Backup current binary
  ↓
Replace with new binary
  ↓
Clear update cache
  ↓
"Update installed successfully"
```

**Update Notifications:**

```text
CLI: Shows notification at start of any command
  ↓
  ⚠ Update available: v0.1.5-beta (dev -> v0.1.5-beta)
    Run probe update to install

Tray: Menu item changes
  ↓
  "Check for Updates" → "Update Available (v0.1.5-beta)"
```

**Update Cache:**

Stored in `~/.../probeTool/update_cache.json`:

```json
{
  "last_check_time": "2026-02-20T15:04:05Z",
  "has_update": true,
  "latest_version": "v0.1.5-beta",
  "download_url": "https://github.com/...",
  "release_page_url": "https://github.com/..."
}
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

### `GET /api/health`

Health check endpoint (used by tray to detect server readiness).

**Response:**

```json
{
  "status": "ok"
}
```

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

## Bundled Agent and Runtime

### Agent Embedding

Agent files are embedded in the binary using Go's `embed` package:

**Files in `internal/agent/files/`:**
- `probe-runner.js` - Main scanner
- `prompts.js` - AI prompts
- `package.json` - Dependencies
- `.claude/skills/security-audit/` - AI context

**Build tags:**
- Default: Files embedded with full context
- `nodedebug`: Stub implementation for development without full files

### Extraction Process

On first run, `probe setup`:
1. Extracts bundled tar.gz archive
2. Saves to `~/.../probeTool/agent/`
3. Runs `npm install` if needed

**Files:**
- `internal/agent/embed.go` - Embedding directives
- `internal/agent/embed_stub.go` - Debug stub
- `internal/agent/agent.go` - Extraction logic

### Runtime Embedding

Similarly, Next.js dashboard is pre-built and embedded:
- `internal/runtime/embed.go` - Dashboard files
- `internal/runtime/runtime.go` - Serving logic

---

## Process Management

### Daemon Spawning

**Unix/macOS/Linux:**
- `cmd/daemon_unix.go` - Uses `Setsid()` to detach
- Process persists after parent exits

**Windows:**
- `cmd/daemon_windows.go` - Uses `CREATE_NEW_PROCESS_GROUP`
- Not recommended - use WSL instead

### PID Tracking

**Packages:**
- `internal/process/` - PID file management
- `server.pid` - Server process ID
- `tray.pid` - Tray process ID

**Usage:**
```go
pidFile := process.GetPIDFile("server")
process.WritePID(pidFile, os.Getpid())
pid := process.ReadPID(pidFile)
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