# Contributing to probeTool

Thank you for your interest! This is a personal project, so contributions are kept simple and straightforward.

## Getting Started

### Prerequisites

- **Go** 1.22 or later
- **Node.js** 20 or later
- **npm** or **yarn**
- **Git**

### Setup

1. **Fork and clone:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/probeTool.git
   cd probeTool
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   cd web && npm install && cd ..
   cd agent && npm install && cd ..
   ```

3. **Build:**
   ```bash
   make probe
   ```

4. **Test:**
   ```bash
   ./probe --help
   ./probe tray
   ```

---

## Making Changes

### Branch Naming

Create a descriptive branch from main:

```bash
git checkout -b feat/add-export-feature
git checkout -b fix/tray-icon-crash
git checkout -b docs/update-readme
```

### Commit Message Format

Use conventional commits:

```text
<type>: <short description>

[optional body]

Types:
  feat:     New feature
  fix:      Bug fix
  docs:     Documentation only
  style:    Code style (formatting, semicolons, etc.)
  refactor: Code restructuring without behavior change
  test:     Adding or updating tests
  chore:    Build, dependencies, tooling

Examples:
  feat: add PDF export for scan reports
  fix: resolve tray icon not displaying on Windows
  docs: update installation instructions
  refactor: simplify database query logic
```

### Code Style

- **Go:** Follow `gofmt` (automatically enforced)
  ```bash
  go fmt ./...
  ```

- **TypeScript/JavaScript:** Use Prettier
  ```bash
  cd web && npm run format
  ```

- **Keep it simple** - This is a personal project, not enterprise code

---

## Testing

### Before Committing

```bash
# Run all tests
make test

# Or individually:
go test ./...                    # Go tests
cd agent && npm test             # Agent tests
cd web && npm run build          # Verify build
```

### Manual Testing

```bash
# Build and test locally
make probe
./probe version
./probe tray
```

---

## Pull Request Process

### 1. Prepare Your PR

- One feature/fix per PR - Keep changes focused
- Update tests if applicable
- Update documentation if behavior changes
- Test locally before pushing

### 2. Create Pull Request

Title format:

```text
feat: add dark mode to dashboard
fix: resolve database lock issue
docs: clarify configuration steps
```

Description should include:

- What changed
- Why it changed
- How to test it
- Screenshots (if UI changes)

### 3. Example PR Description

```text
## What

Adds dark mode toggle to the dashboard.

## Why

Improves usability in low-light conditions.

## How to Test

1. Build: `make probe`
2. Run: `probe serve`
3. Open dashboard
4. Click theme toggle in top-right

## Screenshots

[Screenshot here]
```

### 4. Review Process

- Maintainer will review within a few days
- Address any requested changes
- Once approved, it will be merged

---

## Project Structure

```text
probeTool/
├── cmd/                     # CLI commands
│   ├── probe/              # Main entrypoint
│   ├── root.go             # Root command
│   ├── serve.go            # Dashboard server
│   ├── tray.go             # System tray
│   ├── config.go           # Configuration
│   └── ...
├── internal/               # Internal packages
│   ├── config/            # Config management
│   ├── db/                # Database layer
│   ├── prober/            # Scan logic
│   ├── server/            # HTTP server
│   ├── tray/              # Tray functionality
│   ├── paths/             # OS-specific paths
│   └── version/           # Version info
├── agent/                  # Node.js AI agent
│   ├── probe-runner.js    # Main scanner
│   └── prompts.js         # AI prompts
├── web/                    # Next.js dashboard
│   ├── app/               # Pages (App Router)
│   ├── components/        # React components
│   └── lib/               # Utilities
└── .github/                # CI/CD workflows
```

---

## What to Work On

### Good First Issues

- Documentation improvements
- UI polish
- Simple bug fixes

### Feature Ideas

- Custom scan rules

> Before starting major features, open an issue to discuss!

---

## Code Guidelines

### Do's ✅

- Use `internal/paths` for file paths (OS-agnostic)
- Return errors, don't call `os.Exit()` in libraries
- Write tests for new features
- Keep functions small and focused
- Use descriptive variable names

### Don'ts ❌

- Hardcode file paths (e.g., `~/.probe/config.json`)
- Ignore errors
- Commit large binary files
- Break existing tests
- Change default port (`37330`)

---

## Getting Help

- **Questions?** Open an issue with `question` label
- **Stuck?** Check [DEVELOPMENT.md](DEVELOPMENT.md)
- **Bug?** Open an issue with reproduction steps

---

## Code of Conduct

Be respectful and constructive. This is a welcoming, inclusive project.

---

Thank you for contributing! 🎉