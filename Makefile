.PHONY: probe probe-only install setup test clean web web-deps web-build dev-web dev-api clean-web clean-all

# ─── Web Dashboard ───────────────────────────────────────────────────────────

web-deps:
	cd web && npm install

web-build: web-deps
	cd web && npm run build

web: web-build

# ─── Go Binary ───────────────────────────────────────────────────────────────

probe: web
	go build -o probe ./cmd/probe

probe-only:
	go build -o probe ./cmd/probe

install: probe
	go install ./cmd/probe
	./probe setup

setup:
	./probe setup

# ─── Test ────────────────────────────────────────────────────────────────────

test:
	cd ~/test-repo && probe --full

# ─── Dev ─────────────────────────────────────────────────────────────────────

dev-web:
	cd web && npm run dev

dev-api: probe-only
	./probe

# ─── Clean ───────────────────────────────────────────────────────────────────

clean:
	rm -f probe
	@if command -v ./probe >/dev/null 2>&1; then \
		./probe clean; \
	else \
		rm -rf ~/.probe/probes/*.md 2>/dev/null || true; \
	fi

clean-web:
	rm -rf web/.next web/out

clean-all: clean clean-web
