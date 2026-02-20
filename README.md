# probeTool

> AI-powered security scanning tool with web dashboard and system tray integration

[![Release](https://img.shields.io/github/v/release/ndzuma/probeTool)](https://github.com/ndzuma/probeTool/releases)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## ⚠️ Windows Notice

**The Windows build is not working correctly.** Users should use WSL (Windows Subsystem for Linux) instead. Anyone else using Windows does so at their own risk.

[See DOCUMENTATION.md](#troubleshooting) for WSL setup instructions.

---

## Features

- 🔍 **AI-Powered Scanning** - Automated security analysis using Claude AI
- 📊 **Web Dashboard** - View scan results in a modern Next.js interface
- 🖥️ **System Tray** - Background operation with menu bar integration (includes server)
- 💾 **Local Storage** - SQLite database for scan history
- 🚀 **Fast & Lightweight** - Single binary with embedded frontend
- 🔧 **Configurable** - Support for multiple AI providers

---

## Installation

### Quick Install (macOS/Linux)

```bash
curl -sSL https://raw.githubusercontent.com/ndzuma/probeTool/main/install.sh | bash
```

### Homebrew (macOS)

```bash
brew tap ndzuma/probetool
brew install probetool
```

### Manual Download

Download the latest release for your platform:

#### macOS (Apple Silicon):

```bash
curl -L -o probe.tar.gz https://github.com/ndzuma/probeTool/releases/latest/download/probeTool_*_darwin_arm64.tar.gz
tar -xzf probe.tar.gz
sudo mv probe /usr/local/bin/
```

#### macOS (Intel):

```bash
curl -L -o probe.tar.gz https://github.com/ndzuma/probeTool/releases/latest/download/probeTool_*_darwin_amd64.tar.gz
tar -xzf probe.tar.gz
sudo mv probe /usr/local/bin/
```

#### Linux:

```bash
curl -L -o probe.tar.gz https://github.com/ndzuma/probeTool/releases/latest/download/probeTool_*_linux_amd64.tar.gz
tar -xzf probe.tar.gz
sudo mv probe /usr/local/bin/
```

#### Windows:

Download `probeTool_*_windows_amd64.zip` from [releases](https://github.com/ndzuma/probeTool/releases) and extract to your PATH.

---

## Quick Start

### 1. Configure API Provider

```bash
probe config add-provider openrouter
```

You'll be prompted for:

- Base URL (e.g., `https://openrouter.ai/api/v1`)
- API Key
- Models (e.g., `anthropic/claude-3.5-haiku`)

### 2. Run System Tray (Recommended)

```bash
probe tray
```

What it does:

- Launches system tray icon in your menu bar
- Automatically starts dashboard server
- Provides quick access menu:
  - Open Dashboard
  - Check for Updates
  - Restart Server
  - View Version
  - Quit

### 3. Run a Security Scan

```bash
# Full scan of current directory
probe --full

# Quick scan
probe --quick

# With custom model
probe --model anthropic/claude-3.5-sonnet

# Verbose output
probe --verbose
```

### 4. View Results

The dashboard opens automatically at `http://localhost:37330`

Or manually start the server:

```bash
probe serve
```

---

## Usage

### Commands

```text
probe                     Run a security scan (default: full)
probe tray                Launch system tray (includes dashboard server)
probe serve               Start dashboard server only
probe serve --quiet       Start server as background daemon
probe stop                Stop running server daemon
probe status              Show server/tray status with PIDs
probe config              Manage configuration
probe setup               Install agent files (runs automatically on first use)
probe clean               Clean scan reports
probe migrate             Migrate config to new location
probe version             Show version information
probe --help              Show all commands and flags
```

### Flags

```text
--full                    Run a full scan (default)
--quick                   Run a quick scan
--model <model>           Override default AI model
--verbose, -v             Enable verbose output
--version                 Show short version
```

### Examples

```bash
# Run tray in background
probe tray &

# Full scan with specific model
probe --full --model anthropic/claude-3.5-sonnet

# Check version
probe -v
probe version           # Detailed version info
probe version --json    # JSON format

# Configure new provider
probe config add-provider anthropic
probe config set-key openrouter sk-...
probe config list
```

---

## Configuration

### File Locations

- **macOS:** `~/Library/Application Support/probeTool/`
- **Linux:** `~/.config/probeTool/`
- **Windows:** `%APPDATA%\probeTool\`

### Structure

```text
probeTool/
├── config.json          # Provider configuration
├── probes/
│   ├── probes.db        # SQLite database
│   └── *.md             # Scan reports
├── agent/               # AI agent files
└── cache/               # Temporary files
```

### Config File Format

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

## Architecture

```text
┌───────────────────────────────────────────┐
│                   User                    │
└────────┬──────────────────────────────────┘
         │
         ▼
┌───────────────────────────────────────────┐
│   CLI (Go)                                │
│  ┌────────┬────────┬────────┬──────────┐  │
│  │ probe  │ serve  │  tray  │  config  │  │
│  └────────┴────────┴────────┴──────────┘  │
└────┬─────────────────┬────────────────────┘
     │                 │
     ▼                 ▼
┌───────────┐    ┌───────────────┐
│   Agent   │    │  Dashboard    │
│ (Node.js) │    │  (Next.js)    │
│           │    │  Port 37330   │
└────┬──────┘    └───────────────┘
     │
     ▼
┌───────────┐
│ SQLite DB │
└───────────┘
```

See [DOCUMENTATION.md](DOCUMENTATION.md) for detailed architecture.

---

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md) for:

- Development setup
- Project structure
- Making changes
- Testing
- Release process

---

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:

- Setting up development environment
- Commit message format
- Pull request process
- Code style

---

## Documentation

- [DOCUMENTATION.md](DOCUMENTATION.md) - Architecture and API reference
- [DEVELOPMENT.md](DEVELOPMENT.md) - Development guide with diagrams
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
- [.github/AI_CONTEXT.md](.github/AI_CONTEXT.md) - Context for AI assistants

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

## Support

- 🐛 [Report Issues](https://github.com/ndzuma/probeTool/issues)
- 📧 Questions? [Open an issue!](https://github.com/ndzuma/probeTool/issues/new)

---

## Acknowledgments

- Built with [Claude Agent SDK](https://platform.claude.com/docs/en/agent-sdk/overview) for security analysis
- Powered by [Go](https://go.dev), [Next.js](https://nextjs.org), and [Node.js](https://nodejs.org)
