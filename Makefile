.PHONY: probe probe-only install setup test clean web web-deps web-build dev-web dev-api clean-web clean-runtime clean-all bundle-node bundle-web rebuild-quick

# ─── Version Info ────────────────────────────────────────────────────────────

VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION ?= $(shell go version | awk '{print $$3}')

LDFLAGS := -ldflags "\
	-X github.com/ndzuma/probeTool/internal/version.Version=$(VERSION) \
	-X github.com/ndzuma/probeTool/internal/version.Commit=$(COMMIT) \
	-X github.com/ndzuma/probeTool/internal/version.BuildDate=$(BUILD_DATE) \
	-X github.com/ndzuma/probeTool/internal/version.GoVersion=$(GO_VERSION)"

# ─── Platform Detection ──────────────────────────────────────────────────────

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# ─── Web Dashboard ───────────────────────────────────────────────────────────

web-deps:
	cd web && npm install

web-build: web-deps
	cd web && npm run build

web: web-build

# ─── Node.js Runtime Bundling ────────────────────────────────────────────────

bundle-node:
	@echo "📦 Bundling Node.js for $(GOOS)/$(GOARCH)..."
	@chmod +x scripts/bundle-node.sh
	@./scripts/bundle-node.sh $(GOOS) $(GOARCH)

bundle-node-darwin-arm64:
	@GOOS=darwin GOARCH=arm64 $(MAKE) bundle-node

bundle-node-darwin-amd64:
	@GOOS=darwin GOARCH=amd64 $(MAKE) bundle-node

bundle-node-linux-amd64:
	@GOOS=linux GOARCH=amd64 $(MAKE) bundle-node

bundle-node-windows-amd64:
	@GOOS=windows GOARCH=amd64 $(MAKE) bundle-node

bundle-web: web
	@echo "📦 Bundling web directory..."
	@rm -rf internal/runtime/web
	@mkdir -p internal/runtime/web
	@cp -r web/* internal/runtime/web/
	@rm -rf internal/runtime/web/node_modules internal/runtime/web/.next internal/runtime/web/out
	@echo "✅ Web directory bundled"

# Check if runtime is bundled
check-runtime:
	@if [ ! -d "internal/runtime/node-$(GOOS)-$(GOARCH)" ]; then \
		echo "❌ Node.js runtime not bundled for $(GOOS)/$(GOARCH)"; \
		echo "   Run 'make bundle-node' first"; \
		exit 1; \
	fi
	@if [ ! -d "internal/runtime/web" ]; then \
		echo "❌ Web directory not bundled"; \
		echo "   Run 'make bundle-web' first"; \
		exit 1; \
	fi

# ─── Go Binary ───────────────────────────────────────────────────────────────

probe: bundle-node bundle-web
	@echo "🔨 Building probe with bundled runtime..."
	go build $(LDFLAGS) -o probe ./cmd/probe
	@echo "✅ Build complete: ./probe"

probe-only:
	go build $(LDFLAGS) -o probe ./cmd/probe

install: probe
	@echo "📦 Installing probe..."
	cp probe /usr/local/bin/
	probe setup
	@echo "✅ Installed to /usr/local/bin/probe"

setup:
	./probe setup

# ─── Test ────────────────────────────────────────────────────────────────────

test:
	cd ~/test-repo && probe --full

test-go:
	go test ./... -v

test-web:
	cd web && npm test

# ─── Dev ─────────────────────────────────────────────────────────────────────

dev-web:
	cd web && npm run dev

dev-api: probe-only
	./probe serve

# ─── Clean ───────────────────────────────────────────────────────────────────

clean:
	rm -f probe
	@if command -v ./probe >/dev/null 2>&1; then \
		./probe clean; \
	else \
		rm -rf ~/.probe/probes/*.md 2>/dev/null || true; \
	fi

clean-web:
	rm -rf web/.next web/out web/node_modules

clean-runtime:
	@echo "🧹 Cleaning bundled runtime..."
	rm -rf internal/runtime/node-* internal/runtime/web
	@echo "✅ Runtime cleaned"

clean-all: clean clean-web clean-runtime
	@echo "✨ All clean!"

# ─── Release Helpers ─────────────────────────────────────────────────────────

# Bundle all platforms (for releases)
bundle-all:
	@echo "📦 Bundling all platforms..."
	$(MAKE) bundle-node-darwin-arm64
	$(MAKE) bundle-node-darwin-amd64
	$(MAKE) bundle-node-linux-amd64
	$(MAKE) bundle-node-windows-amd64
	$(MAKE) bundle-web
	@echo "✅ All platforms bundled"

# Quick rebuild (assumes runtime already bundled)
rebuild: check-runtime
	@echo "🔨 Quick rebuild (runtime already bundled)..."
	go build $(LDFLAGS) -o probe ./cmd/probe
	@echo "✅ Build complete"
