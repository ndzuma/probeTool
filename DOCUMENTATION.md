# probeTool Documentation

Complete technical documentation for probeTool.

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Components](#components)
4. [Configuration](#configuration)
5. [Database](#database)
6. [API Reference](#api-reference)
7. [System Tray](#system-tray)
8. [Troubleshooting](#troubleshooting)

---

## Overview

probeTool is a security scanning CLI tool that combines:
- **Go binary** - Fast, cross-platform CLI
- **Node.js agent** - AI-powered code analysis
- **Next.js dashboard** - Modern web interface
- **System tray** - Background operation with server integration

---

## Architecture

### High-Level Overview

```text
┌─────────────────────────────────────────────┐
│ User                                         │
└──────────┬──────────────────────────────────┘
           │
           ▼
┌──────────────────────────────────────────────┐
│ CLI Binary (Go)                               │
│ ┌────────┬────────┬────────┬──────────┐      │
│ │ probe  │ serve  │  tray  │  config  │      │
│ │ (scan) │ (http) │ (both) │          │      │
│ └────────┴────────┴────────┴──────────┘      │
└──────┬────────────────┬──────────────────────┘
       │                │
       ▼                ▼
┌─────────────┐  ┌──────────────────────────┐
│  Agent      │  │  Dashboard               │
│  (Node.js)  │  │  (Next.js)               │
│             │  │  Port 37330              │
│  - AI calls │  │                          │
│  - Analysis │  │  - Scan list             │
│  - Reports  │  │  - Findings view         │
└──────┬──────┘  │  - File explorer         │
       │         └──────────────────────────┘
       ▼
┌──────────────────┐
│  SQLite DB       │
│  ~/.../probes/   │
│                  │
│  - probes table  │
│  - findings      │
└──────────────────┘
```

### Data Flow: Running a Scan

```text
User executes:
$ probe --full
   │
   ▼
CLI validates
   ├── Config exists?
   ├── API key set?
   └── Target path valid?
   │
   ▼
CLI creates probe record in DB
   ├── ID: timestamp-type
   └── Status: "running"
   │
   ▼
CLI spawns Node.js agent:
   node agent/probe-runner.js
   --target=/path/to/code
   --out=/path/to/report.md
   --model=claude-3.5-haiku
   │
   ▼
Agent reads codebase
   ├── Scans files
   ├── Filters by .gitignore
   └── Groups by severity
   │
   ▼
Agent calls AI API
   ├── Sends code context
   ├── Requests security analysis
   └── Receives findings
   │
   ▼
Agent writes markdown report
   ├── Formatted findings
   ├── Severity levels
   └── Code snippets
   │
   ▼
CLI updates DB
   ├── Status: "completed"
   ├── Parses findings
   └── Inserts into DB
   │
   ▼
User views in dashboard
   http://localhost:37330/probes/{id}
```

### System Tray Flow

```text
User executes:
$ probe tray
   │
   ▼
Tray app starts
   ├── Registers system tray icon
   ├── Builds menu
   └── Starts event loop
   │
   ▼
Tray spawns server subprocess:
   probe serve --quiet
   │
   ▼
Server starts
   ├── Next.js on port 37330
   ├── SQLite database
   └── API endpoints
   │
   ▼
User clicks "Open Dashboard"
   │
   ▼
Tray opens browser:
   http://localhost:37330
   │
   ▼
User clicks "Quit"
   │
   ▼
Tray stops server gracefully
   ├── Sends SIGINT
   ├── Waits for shutdown
   └── Exits
```

---

## Components

### CLI (`cmd/`)

**Main Commands:**

| Command | File | Description |
|---------|------|-------------|
| `probe` | `root.go` | Run security scan (default command) |
| `probe serve` | `serve.go` | Start dashboard HTTP server only |
| `probe serve --quiet` | `serve.go` | Start server as background daemon (no browser) |
| `probe tray` | `tray.go` | Launch system tray (includes server) |
| `probe stop` | `stop.go` | Stop running server daemon |
| `probe status` | `status.go` | Show server/tray status with PIDs |
| `probe config` | `config.go` | Manage API provider configuration |
| `probe setup` | `setup.go` | Install agent files from bundled archive |
| `probe clean` | `clean.go` | Clean scan reports |
| `probe migrate` | `migrate.go` | Migrate config to new location |
| `probe version` | `version.go` | Show version information |

**Internal Packages:**

| Package | Path | Purpose |
|---------|------|---------|
| config | `internal/config/` | Provider configuration management |
| db | `internal/db/` | SQLite database operations |
| prober | `internal/prober/` | Scan execution logic |
| server | `internal/server/` | HTTP server and API handlers |
| tray | `internal/tray/` | System tray functionality |
| paths | `internal/paths/` | OS-specific path resolution |
| version | `internal/version/` | Version information |
| findings | `internal/findings/` | Parsing scan results |

### Agent (`agent/`)

**Bundled Files:**

Agent files are embedded in the probeTool binary for easy distribution:
- `probe-runner.js` - Main scanner entrypoint
- `prompts.js` - AI prompt templates
- `package.json` - Node dependencies
- `.claude/` - AI context files

On first run, `probe setup` automatically extracts these to:
- **macOS/Linux:** `~/.../probeTool/agent/`
- **Windows (WSL):** `%APPDATA%\probeTool\agent\`

**Process:**
1. Parse command-line arguments
2. Read target directory
3. Filter files (respects `.gitignore`)
4. Send to AI for analysis
5. Parse AI response
6. Generate markdown report
7. Save to specified output path

**Manual Setup:**
```bash
# Manually install agent files
probe setup
```

**Environment Variables:**
- `ANTHROPIC_AUTH_TOKEN` - API key
- `ANTHROPIC_BASE_URL` - API endpoint
- `ANTHROPIC_API_KEY` - Alternative key format

### Dashboard (`web/`)

**Technology:**
- Next.js 14 (App Router)
- React Server Components
- TailwindCSS
- TypeScript

**Key Routes:**
- `/` - Home/scan list
- `/probes` - All scans
- `/probes/[id]` - Scan details
- `/probes/[id]/findings` - Findings list

**API Routes:**
Served by Go HTTP server (not Next.js API routes).

---

## Configuration

### File Locations

Determined by `internal/paths/paths.go`:

| OS | Location |
|----|----------|
| macOS | `~/Library/Application Support/probeTool/` |
| Linux | `~/.config/probeTool/` or `$XDG_CONFIG_HOME/probeTool/` |
| Windows | `%APPDATA%\probeTool\` |

### Directory Structure

```text
probeTool/
├── config.json          # Provider configuration
├── probes/
│   ├── probes.db        # SQLite database
│   ├── 2026-02-20-*.md  # Scan reports
│   └── ...
├── agent/
│   ├── probe-runner.js  # Agent files
│   ├── prompts.js
│   ├── package.json
│   └── node_modules/
└── cache/               # Temporary files
```

### Config File Format

```json
{
  "providers": {
    "openrouter": {
      "name": "openrouter",
      "base_url": "https://openrouter.ai/api/v1",
      "api_key": "sk-or-v1-...",
      "models": [
        "anthropic/claude-3.5-haiku",
        "anthropic/claude-3.5-sonnet"
      ],
      "default_model": "anthropic/claude-3.5-haiku"
    },
    "anthropic": {
      "name": "anthropic",
      "base_url": "https://api.anthropic.com/v1",
      "api_key": "sk-ant-...",
      "models": ["claude-3-5-haiku-20241022"],
      "default_model": "claude-3-5-haiku-20241022"
    }
  },
  "default": "openrouter"
}
```

---

## Database

### Schema

File: `~/.../probeTool/probes/probes.db`

**Tables:**

```sql
-- Scan records
CREATE TABLE probes (
    id TEXT PRIMARY KEY,              -- Format: 2026-02-20-150405-full
    type TEXT NOT NULL,               -- "full" or "quick"
    target_path TEXT NOT NULL,        -- Scanned directory
    output_path TEXT NOT NULL,        -- Markdown report location
    status TEXT DEFAULT 'running',    -- "running", "completed", "failed"
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Security findings
CREATE TABLE findings (
    id TEXT PRIMARY KEY,              -- Format: {probe_id}-finding-{n}
    probe_id TEXT NOT NULL,           -- Foreign key to probes.id
    text TEXT NOT NULL,               -- Finding description
    severity TEXT NOT NULL,           -- "critical", "high", "medium", "low"
    FOREIGN KEY(probe_id) REFERENCES probes(id) ON DELETE CASCADE
);
```

**Indexes:**

```sql
CREATE INDEX idx_probes_created ON probes(created_at DESC);
CREATE INDEX idx_findings_probe ON findings(probe_id);
CREATE INDEX idx_findings_severity ON findings(severity);
```

### Operations

Common queries:

```go
// List recent scans
SELECT * FROM probes ORDER BY created_at DESC LIMIT 20

// Get scan with findings count
SELECT p.*, COUNT(f.id) as findings_count
FROM probes p
LEFT JOIN findings f ON p.id = f.probe_id
GROUP BY p.id

// Get findings by severity
SELECT * FROM findings
WHERE probe_id = ?
AND severity = 'critical'
ORDER BY id
```

---

## API Reference

All endpoints served by Go HTTP server at `http://localhost:37330`.

### `GET /api/probes`

List all scans.

**Response:**

```json
{
  "probes": [
    {
      "id": "2026-02-20-150405-full",
      "type": "full",
      "target_path": "/Users/user/project",
      "output_path": "/Users/user/.../probes/2026-02-20-150405-full.md",
      "status": "completed",
      "created_at": "2026-02-20T15:04:05Z",
      "findings_count": 12
    }
  ]
}
```

### `GET /api/probes/:id`

Get specific scan details.

**Response:**

```json
{
  "probe": {
    "id": "2026-02-20-150405-full",
    "type": "full",
    "target_path": "/Users/user/project",
    "output_path": "/Users/user/.../probes/2026-02-20-150405-full.md",
    "status": "completed",
    "created_at": "2026-02-20T15:04:05Z"
  },
  "report": "# Security Assessment\n\n## Critical Findings\n...",
  "findings_count": 12
}
```

### `GET /api/findings/:probe_id`

Get all findings for a scan.

**Query Parameters:**
- `severity` (optional) - Filter by severity

**Response:**

```json
{
  "findings": [
    {
      "id": "2026-02-20-150405-full-finding-1",
      "probe_id": "2026-02-20-150405-full",
      "text": "SQL injection vulnerability in auth.go:42",
      "severity": "critical"
    }
  ]
}
```

### `GET /api/config`

Get current configuration.

**Response:**

```json
{
  "providers": ["openrouter", "anthropic"],
  "default": "openrouter",
  "current_model": "anthropic/claude-3.5-haiku"
}
```

### `GET /api/file-tree/:probe_id`

Get file tree for scanned directory.

**Response:**

```json
{
  "tree": {
    "name": "project",
    "type": "directory",
    "children": [
      {
        "name": "src",
        "type": "directory",
        "children": [...]
      },
      {
        "name": "main.go",
        "type": "file",
        "size": 1024
      }
    ]
  }
}
```

### `GET /api/version`

Get version information.

**Response:**

```json
{
  "version": "0.1.0",
  "commit": "abc123d",
  "commitShort": "abc123d",
  "buildDate": "2026-02-20T15:04:05Z",
  "goVersion": "go1.22.1",
  "platform": "darwin/arm64",
  "dashboardURL": "http://localhost:37330",
  "configPath": "/Users/user/Library/Application Support/probeTool"
}
```

---

## System Tray

### Menu Structure

```text
probeTool
├── Open Dashboard         (opens browser)
├── ─────────────────
├── Check for Updates      (placeholder)
├── ─────────────────
├── Version dev            (disabled, info only)
├── Restart Server         (restarts serve subprocess)
└── Quit                   (stops server, exits)
```

### Implementation

**Files:**
- `cmd/tray.go` - Command registration
- `internal/tray/tray.go` - Main tray logic
- `internal/tray/checks.go` - Server health checks

**Key Functions:**

```go
// Start tray and server
func (m *Manager) Start()

// Build menu items
func (m *Manager) buildMenu()

// Handle menu clicks
func (m *Manager) handleMenuActions()

// Start dashboard server subprocess
func (m *Manager) startServer() error

// Stop server gracefully
func (m *Manager) stopServer() error

// Restart server
func (m *Manager) restartServer()

// Open browser
func (m *Manager) openBrowser(url string)

// Wait for server to be ready
func (m *Manager) waitForServer(timeout time.Duration) error
```

### Server Management

Tray automatically:
- Spawns `probe serve --quiet` as subprocess
- Waits for HTTP server to be ready (polls `/api/health`)
- Updates tooltip: "probeTool - Running"
- On quit: Sends SIGINT, waits for graceful shutdown

---

## WSL (Windows Subsystem for Linux)

### Auto-Detection

On Linux systems, probeTool automatically detects if you're running under WSL:
- Checks `/proc/version` for WSL markers
- On first run with WSL detected, prompts user for system tray preference
- Saves preference to avoid repeated prompts

### WSL on Windows

**Important:** Windows native builds have issues. Use WSL instead:

1. **Install WSL2:**
   ```bash
   wsl --install -d Ubuntu
   ```

2. **Inside WSL terminal:**
   ```bash
   curl -sSL https://raw.githubusercontent.com/ndzuma/probeTool/main/install.sh | bash
   probe config add-provider openrouter
   probe tray
   ```

3. **Access dashboard:**
   - From WSL: `http://localhost:37330`
   - From Windows: `http://127.0.0.1:37330` or WSL IP address
   - WSL hostname: `hostname -I` shows available IPs

### System Tray on WSL

- Tray doesn't render in WSL terminal directly
- Use with `--quiet` flag to run in background:
  ```bash
  probe serve --quiet &
  ```
- Access dashboard from Windows browser at WSL IP:37330
- Or use `probe stop` to gracefully shut down

---

## New Status Command

### `probe status`

Shows the current state of server and tray processes.

**Usage:**
```bash
probe status
```

**Output:**
```
probeTool Status
================
Server:  running (PID: 12345)
Tray:    not running
Reports: 5
```

**Exit codes:**
- `0` - Server running
- `1` - Server not running
- `2` - Error checking status

---

## Process Management

### PID Files

probeTool tracks daemon processes using PID files:

- **Server PID:** `~/.../probeTool/.probe/server.pid`
- **Tray PID:** `~/.../probeTool/.probe/tray.pid`

These allow graceful shutdown and status checking.

### Daemon Spawning

Platform-specific daemon behavior:

- **Unix/macOS/Linux:** Uses `Setsid()` to detach from parent process
- **Windows:** Uses `CREATE_NEW_PROCESS_GROUP` flag (not recommended, use WSL)

### Graceful Shutdown

**Stop commands:**
```bash
# Stop server daemon
probe stop

# Or manually
kill $(cat ~/.../probeTool/.probe/server.pid)

# Force kill all probe processes
pkill probe
```

---

---

## Troubleshooting

### Windows Issues

**Note:** Windows native build is not working correctly. Please use WSL instead.

**Solution:**
1. Install WSL2 and Ubuntu
2. Run probeTool inside WSL terminal
3. Access dashboard from Windows browser at WSL IP:37330

### Port Already in Use

**Error:** `address already in use :37330`

**Solution:**

```bash
# Find process
lsof -ti:37330

# Kill process
lsof -ti:37330 | xargs kill -9

# Or change port (not recommended)
# Edit internal/server/server.go
```

### Database Locked

**Error:** `database is locked`

**Solution:**

```bash
# Close all probe instances
pkill probe

# Remove database
rm ~/Library/Application\ Support/probeTool/probes/probes.db

# Restart
probe serve
```

### Agent Files Missing

**Error:** `agent not installed. Run: probe setup`

**Solution:**

```bash
probe setup
```

Agent files are bundled with probeTool and installed to `~/.../probeTool/agent/`.

### Migration Issues

**Error:** `old config found`

**Solution:**

```bash
# Manual migration
probe migrate

# Or delete old config
rm -rf ~/.probe
```

### Tray Icon Not Showing

- **macOS:**
  - Check System Preferences → Dock & Menu Bar
  - Ensure tray icons are allowed

- **Linux:**
  - Install `libappindicator` or `libayatana-appindicator`
  - Check if running under WSL (auto-detected)

- **Windows:**
  - Use WSL instead of native Windows

### API Key Not Working

Check configuration:

```bash
probe config list

# Set key
probe config set-key openrouter sk-...

# Test scan
probe --quick --verbose
```

### Server Won't Start

Check logs:
```bash
# Kill existing server
probe stop

# Check for port conflicts
lsof -ti:37330

# Try again with verbose output
probe serve --verbose
```

---

## Performance

**Scan times:**
- Quick scan: 30-90 seconds
- Full scan: 2-5 minutes

**Factors:**
- Codebase size
- AI model speed
- Network latency

**Database:**
- SQLite with write-ahead logging (WAL)
- Indexed queries for fast retrieval
- Typically <10MB for 100 scans

---

## Security

### API Keys

- Stored in `config.json` with `0600` permissions
- Never logged or displayed
- Transmitted over HTTPS only

### Database

- Local SQLite database
- No cloud storage
- No telemetry or analytics

### Reports

- Markdown files stored locally
- Contain code snippets (review before sharing)
- Not uploaded anywhere

---

For development details, see [DEVELOPMENT.md](DEVELOPMENT.md).