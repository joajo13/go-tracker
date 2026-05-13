# Portfolio Agent — Phase 0 Setup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stand up the project skeleton so that `make test`, `make lint`, and `make build` pass locally and in CI, with the first TDD-disciplined commit landing real production code (health handler + Money type) following the project's strict TDD rule.

**Architecture:** Single Go binary at `cmd/agent`. Pure-Go SQLite. HTTP via chi. Domain logic in pure packages. Frontend embedded later via `embed.FS` (Phase 4, currently blocked). Setup phase establishes scaffolding, tooling (linter, Makefile, CI), and demonstrates the TDD pattern with two small but real production features: a `/healthz` endpoint and a `Money` value type.

**Tech Stack:** Go 1.22, `go-chi/chi/v5`, `shopspring/decimal`, `stretchr/testify`, `log/slog` (stdlib), `golangci-lint`, GitHub Actions.

**Companion docs:**
- Requirements: `docs/SRS_PortfolioAgent.md`
- Architecture: `docs/superpowers/specs/2026-05-12-portfolio-agent-architecture-design.md`
- Project rules: `CLAUDE.md`

---

## File Structure

Files created in this phase:

| Path | Responsibility |
|------|----------------|
| `.gitignore` | Ignore binaries, env, IDE, OS, DB, node_modules. |
| `.gitattributes` | Force LF line endings repo-wide. |
| `.editorconfig` | Editor-agnostic indent/encoding rules. |
| `LICENSE` | MIT, Juan Giupponi 2026. |
| `README.md` | Project intro + dev quickstart. |
| `.env.example` | Tracked template of required env vars. |
| `go.mod`, `go.sum` | Module declaration + dep lockfile. |
| `.golangci.yml` | Strict linter config. |
| `Makefile` | Dev/CI entry points: `test`, `lint`, `build`, `run`, `cover`, `ci`. |
| `.github/workflows/ci.yml` | Lint + test + build on push/PR. |
| `cmd/agent/main.go` | Entrypoint: load env, init slog, start chi server, graceful shutdown. |
| `cmd/agent/main_test.go` | Integration test for the router wiring. |
| `internal/api/health.go` | `/healthz` handler (NFR-O-03). |
| `internal/api/health_test.go` | Unit test for the health handler. |
| `internal/domain/money.go` | `Money` type wrapping `shopspring/decimal`. |
| `internal/domain/money_test.go` | TDD demo: parsing, precision, error paths. |
| `internal/{config,sources,persistence,scheduler,workers,pnl,alerts,indicators,downsampler,broadcaster,web,calendar}/.gitkeep` | Empty dirs reserved for later phases. |
| `web/.gitkeep` | Frontend dir reserved (Phase 4). |
| `migrations/.gitkeep` | Goose migrations dir reserved (Phase 1). |
| `scripts/.gitkeep` | Dev scripts dir reserved. |

**Module path:** `github.com/joajo13/go-tracker` — change in Task 2 step 1 if your GitHub namespace differs.

---

## Task 1: Repo metadata files

**Files:**
- Create: `.gitignore`
- Create: `.gitattributes`
- Create: `.editorconfig`
- Create: `LICENSE`
- Create: `README.md` (stub — polished in Task 9)
- Create: `.env.example`

- [ ] **Step 1.1: Create `.gitignore`**

```
# Go binaries
/bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test + coverage
*.test
*.out
coverage.out
coverage.html

# Go workspace
go.work
go.work.sum

# Environment / secrets
.env
.env.local
*.pem

# Editors / OS
.idea/
.vscode/
*.swp
.DS_Store
Thumbs.db

# Frontend (Phase 4)
web/node_modules/
web/dist/
web/.vite/

# SQLite local DB
*.db
*.sqlite
*.sqlite3
data/

# Logs
*.log
logs/
```

- [ ] **Step 1.2: Create `.gitattributes`**

```
* text=auto eol=lf

*.go        text eol=lf
*.md        text eol=lf
*.yml       text eol=lf
*.yaml      text eol=lf
*.json      text eol=lf
*.toml      text eol=lf
*.sh        text eol=lf
Makefile    text eol=lf

*.png       binary
*.jpg       binary
*.ico       binary
*.db        binary
```

- [ ] **Step 1.3: Create `.editorconfig`**

```
root = true

[*]
end_of_line = lf
insert_final_newline = true
charset = utf-8
trim_trailing_whitespace = true

[*.go]
indent_style = tab
indent_size = 4

[{*.yml,*.yaml,*.json,*.md,*.toml}]
indent_style = space
indent_size = 2

[Makefile]
indent_style = tab
```

- [ ] **Step 1.4: Create `LICENSE`**

```
MIT License

Copyright (c) 2026 Juan Giupponi

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

- [ ] **Step 1.5: Create `README.md` (stub — fleshed out in Task 9)**

```markdown
# Portfolio Agent (go-tracker)

Personal investment portfolio monitoring agent in Go. Read-only over the
financial world — no trading.

Full spec: [`docs/SRS_PortfolioAgent.md`](docs/SRS_PortfolioAgent.md).
Architecture: [`docs/superpowers/specs/2026-05-12-portfolio-agent-architecture-design.md`](docs/superpowers/specs/2026-05-12-portfolio-agent-architecture-design.md).

## Status

Phase 0 setup in progress. See `docs/superpowers/plans/` for the active plan.

## License

MIT — see [`LICENSE`](LICENSE).
```

- [ ] **Step 1.6: Create `.env.example`**

```
# HTTP server
HTTP_ADDR=:8080

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Database (Phase 1+)
DB_PATH=./data/portfolio.db

# Dashboard auth (Phase 4+)
# Generate a bcrypt hash from your password before setting this.
DASHBOARD_PASSWORD_HASH=

# Worker pool (Phase 1+)
WORKER_POOL_SIZE=10

# External API keys (Phase 1+)
FINNHUB_API_KEY=

# Broker creds (post-MVP, encrypted)
BULL_USER=
BULL_PASS_ENCRYPTED=
```

- [ ] **Step 1.7: Stage and commit**

```bash
git add .gitignore .gitattributes .editorconfig LICENSE README.md .env.example
git status
git commit -m "$(cat <<'EOF'
chore: add repo metadata, license, env template

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

Expected: clean commit, no test/lint runs yet (no Go code).

---

## Task 2: Initialize Go module and commit existing project foundation

**Files:**
- Create: `go.mod`
- Modify (track): `CLAUDE.md`, `docs/SRS_PortfolioAgent.md`, `docs/superpowers/specs/2026-05-12-portfolio-agent-architecture-design.md` (the spec was committed in the brainstorming phase, but `CLAUDE.md` and the SRS are still untracked).

- [ ] **Step 2.1: Initialize Go module**

```bash
go mod init github.com/joajo13/go-tracker
```

Expected output: `go: creating new go.mod: module github.com/joajo13/go-tracker`

`go.mod` contents after this step:

```
module github.com/joajo13/go-tracker

go 1.22
```

- [ ] **Step 2.2: Track the existing foundation docs that pre-dated this phase**

```bash
git add go.mod CLAUDE.md docs/SRS_PortfolioAgent.md
git commit -m "$(cat <<'EOF'
chore: initialize Go module and track project foundation docs

Track CLAUDE.md and the SRS that already existed in the working tree.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: golangci-lint config

**Files:**
- Create: `.golangci.yml`

- [ ] **Step 3.1: Create `.golangci.yml`**

```yaml
run:
  timeout: 5m
  go: "1.22"

linters:
  disable-all: true
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    - revive
    - gocritic
    - gosec
    - misspell
    - unconvert
    - prealloc
    - nilerr
    - errorlint

linters-settings:
  govet:
    enable-all: true
    disable:
      - fieldalignment

  revive:
    rules:
      - name: var-naming
      - name: package-comments
      - name: exported
      - name: error-return
      - name: error-strings
      - name: error-naming
      - name: blank-imports
      - name: context-as-argument
      - name: context-keys-type
      - name: dot-imports
      - name: indent-error-flow
      - name: range
      - name: receiver-naming
      - name: time-naming
      - name: unexported-return
      - name: unreachable-code

  gocritic:
    enabled-tags:
      - diagnostic
      - performance
      - style

  errcheck:
    check-type-assertions: true
    check-blank: false

issues:
  exclude-dirs:
    - web
    - bin
    - migrations
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
        - gosec
```

- [ ] **Step 3.2: Verify the config parses**

```bash
golangci-lint config verify
```

Expected: no output, exit code 0. If `golangci-lint` is not installed, install with:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
```

(Or use `scoop install golangci-lint` on Windows.)

- [ ] **Step 3.3: Commit**

```bash
git add .golangci.yml
git commit -m "$(cat <<'EOF'
chore: add strict golangci-lint config

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: Makefile

**Files:**
- Create: `Makefile`

- [ ] **Step 4.1: Create `Makefile`**

```makefile
.PHONY: help test test-race lint build run cover mocks migrate ci web-build clean

GO ?= go
BIN := bin/agent
CMD := ./cmd/agent
COVERAGE := coverage.out

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "%-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## Run all tests
	$(GO) test ./...

test-race: ## Run all tests with the race detector
	$(GO) test -race ./...

cover: ## Run tests with coverage report
	$(GO) test -coverprofile=$(COVERAGE) ./...
	$(GO) tool cover -func=$(COVERAGE)

lint: ## Run golangci-lint
	golangci-lint run ./...

build: ## Build the agent binary
	$(GO) build -o $(BIN) $(CMD)

run: build ## Build and run the agent
	./$(BIN)

mocks: ## Regenerate gomock mocks (Phase 1+)
	$(GO) generate ./...

migrate: ## Run database migrations (wired in Phase 1, no-op for now)
	@echo "migrate target reserved for Phase 1 (goose). Skipping."

web-build: ## Build the embedded frontend (Phase 4+)
	cd web && npm install && npm run build

ci: lint test build ## Run the full local CI check

clean: ## Remove build artifacts
	rm -rf bin/ $(COVERAGE)
```

- [ ] **Step 4.2: Verify Makefile targets list cleanly**

```bash
make help
```

Expected: a list of all targets with descriptions. If `make` is not installed on Windows, install via `choco install make` or `scoop install make` (or use the Bash tool that the agent already has).

- [ ] **Step 4.3: Commit**

```bash
git add Makefile
git commit -m "$(cat <<'EOF'
chore: add Makefile with test/lint/build/ci targets

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: Folder scaffolding

**Files:**
- Create empty directories with `.gitkeep` files per SRS §5.4.

- [ ] **Step 5.1: Create directories**

```bash
mkdir -p cmd/agent
mkdir -p internal/{config,domain,sources,persistence,persistence/sqlite,persistence/mocks,scheduler,workers,broadcaster,pnl,alerts,indicators,downsampler,api,web,calendar}
mkdir -p web
mkdir -p migrations
mkdir -p scripts
mkdir -p .github/workflows
```

- [ ] **Step 5.2: Add `.gitkeep` to each empty directory**

Run (PowerShell-friendly version using the Bash tool):

```bash
for d in \
  internal/config \
  internal/sources \
  internal/persistence \
  internal/persistence/sqlite \
  internal/persistence/mocks \
  internal/scheduler \
  internal/workers \
  internal/broadcaster \
  internal/pnl \
  internal/alerts \
  internal/indicators \
  internal/downsampler \
  internal/web \
  internal/calendar \
  web \
  migrations \
  scripts \
; do
  touch "$d/.gitkeep"
done
```

(`internal/domain`, `internal/api`, `cmd/agent`, and `.github/workflows` get real files in later tasks — no `.gitkeep` needed there.)

- [ ] **Step 5.3: Verify the structure**

```bash
find . -type d -not -path './.git*' -not -path './bin*' | sort
```

Expected: directories per the SRS §5.4 layout.

- [ ] **Step 5.4: Commit**

```bash
git add internal/ web/ migrations/ scripts/
git commit -m "$(cat <<'EOF'
chore: scaffold project folder structure per SRS section 5.4

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: TDD — Health handler

This is the first TDD-discipline commit on real production code. Red → green → refactor → commit.

**Files:**
- Create: `internal/api/health_test.go`
- Create: `internal/api/health.go`

- [ ] **Step 6.1 (RED): Write the failing test**

Create `internal/api/health_test.go`:

```go
package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joajo13/go-tracker/internal/api"
)

func TestHealthHandler_ReturnsOK(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	api.HealthHandler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, "ok", body["status"])
}
```

- [ ] **Step 6.2: Add the testify dep**

```bash
go get github.com/stretchr/testify@v1.9.0
go mod tidy
```

Expected: `go.mod` gains a `require github.com/stretchr/testify v1.9.0` line and `go.sum` is populated.

- [ ] **Step 6.3: Run the test and confirm it fails (no implementation yet)**

```bash
go test ./internal/api/... -run TestHealthHandler_ReturnsOK -v
```

Expected: compile error — `undefined: api.HealthHandler`.

- [ ] **Step 6.4 (GREEN-MINIMAL): Add a stub that compiles but does not yet pass the assertions**

Create `internal/api/health.go`:

```go
// Package api hosts the HTTP handlers exposed by the portfolio agent.
package api

import "net/http"

// HealthHandler returns a handler that reports the service is alive.
func HealthHandler() http.Handler {
	return http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})
}
```

- [ ] **Step 6.5: Run the test and confirm it now fails on the assertion (true RED)**

```bash
go test ./internal/api/... -run TestHealthHandler_ReturnsOK -v
```

Expected: FAIL — body decode fails or `status` is empty.

- [ ] **Step 6.6 (GREEN): Replace the stub with the real implementation**

Replace the body of `internal/api/health.go`:

```go
// Package api hosts the HTTP handlers exposed by the portfolio agent.
package api

import (
	"encoding/json"
	"net/http"
)

// HealthHandler returns a handler that reports the service is alive.
// Used by /healthz and any external monitoring (NFR-O-03).
func HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
}
```

- [ ] **Step 6.7: Run the test and confirm GREEN**

```bash
go test ./internal/api/... -v
```

Expected: PASS.

- [ ] **Step 6.8: Run the linter on the new code**

```bash
golangci-lint run ./internal/api/...
```

Expected: no issues.

- [ ] **Step 6.9: Commit**

```bash
git add go.mod go.sum internal/api/health.go internal/api/health_test.go
git commit -m "$(cat <<'EOF'
feat(api): add /healthz handler with TDD

First TDD-discipline commit. Implements NFR-O-03 health probe used by
external monitoring (UptimeRobot).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 7: Main entrypoint with chi router

**Files:**
- Create: `cmd/agent/main.go`
- Create: `cmd/agent/main_test.go`

- [ ] **Step 7.1 (RED): Write the integration test for the router wiring**

Create `cmd/agent/main_test.go`:

```go
package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouter_HealthzReturnsOK(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(newRouter())
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/healthz")
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(body), `"status":"ok"`))
}
```

- [ ] **Step 7.2: Add the chi dep**

```bash
go get github.com/go-chi/chi/v5@v5.0.12
go mod tidy
```

- [ ] **Step 7.3: Run the test and confirm it fails (no main yet)**

```bash
go test ./cmd/agent/... -v
```

Expected: compile error — `undefined: newRouter`.

- [ ] **Step 7.4 (GREEN): Implement `cmd/agent/main.go`**

Create `cmd/agent/main.go`:

```go
// Package main is the portfolio-agent entrypoint.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/joajo13/go-tracker/internal/api"
)

const shutdownTimeout = 5 * time.Second

func main() {
	logger := newLogger()
	slog.SetDefault(logger)

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           newRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-stop:
		logger.Info("shutdown signal received")
	case err := <-errCh:
		logger.Error("server error", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}
	logger.Info("server stopped cleanly")
}

func newLogger() *slog.Logger {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}
	if os.Getenv("LOG_FORMAT") == "text" {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}

func newRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Handle("/healthz", api.HealthHandler())

	return r
}
```

- [ ] **Step 7.5: Run the test and confirm GREEN**

```bash
go test ./cmd/agent/... -v
```

Expected: PASS.

- [ ] **Step 7.6: Build the binary**

```bash
go build -o bin/agent ./cmd/agent
```

Expected: produces `bin/agent` (or `bin/agent.exe` on Windows). No errors.

- [ ] **Step 7.7: Smoke-run the binary and hit /healthz manually**

```bash
./bin/agent &
AGENT_PID=$!
sleep 1
curl -s http://localhost:8080/healthz
echo
kill $AGENT_PID
wait $AGENT_PID 2>/dev/null
```

Expected: `{"status":"ok"}` and a clean shutdown log line.

- [ ] **Step 7.8: Run the linter on the new code**

```bash
golangci-lint run ./cmd/agent/...
```

Expected: no issues.

- [ ] **Step 7.9: Commit**

```bash
git add go.mod go.sum cmd/agent/main.go cmd/agent/main_test.go
git commit -m "$(cat <<'EOF'
feat(agent): chi-based HTTP entrypoint with graceful shutdown

Wires /healthz, configures slog from env, supports SIGINT/SIGTERM.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 8: TDD — Money type (domain smoke test)

This task locks in the project's TDD discipline on a piece of foundational
domain code that downstream phases will use.

**Files:**
- Create: `internal/domain/money_test.go`
- Create: `internal/domain/money.go`

- [ ] **Step 8.1 (RED): Write the failing test**

Create `internal/domain/money_test.go`:

```go
package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joajo13/go-tracker/internal/domain"
)

func TestParseAmount_ValidDecimal(t *testing.T) {
	t.Parallel()

	m, err := domain.ParseAmount("1234.5678")

	require.NoError(t, err)
	assert.Equal(t, "1234.5678", m.String())
}

func TestParseAmount_PreservesPrecision(t *testing.T) {
	t.Parallel()

	m, err := domain.ParseAmount("0.00000001")

	require.NoError(t, err)
	assert.Equal(t, "0.00000001", m.String())
}

func TestParseAmount_NegativeValue(t *testing.T) {
	t.Parallel()

	m, err := domain.ParseAmount("-42.5")

	require.NoError(t, err)
	assert.Equal(t, "-42.5", m.String())
}

func TestParseAmount_InvalidString(t *testing.T) {
	t.Parallel()

	_, err := domain.ParseAmount("not-a-number")

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestParseAmount_EmptyString(t *testing.T) {
	t.Parallel()

	_, err := domain.ParseAmount("")

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestMoney_ZeroValueStringIsZero(t *testing.T) {
	t.Parallel()

	var m domain.Money
	assert.Equal(t, "0", m.String())
}
```

- [ ] **Step 8.2: Add the decimal dep**

```bash
go get github.com/shopspring/decimal@v1.4.0
go mod tidy
```

- [ ] **Step 8.3: Run the test and confirm it fails (no implementation)**

```bash
go test ./internal/domain/... -v
```

Expected: compile error — `undefined: domain.ParseAmount`, `undefined: domain.Money`, `undefined: domain.ErrInvalidAmount`.

- [ ] **Step 8.4 (GREEN-MINIMAL): Write a stub that compiles but fails the assertions**

Create `internal/domain/money.go`:

```go
// Package domain holds the pure domain types of the portfolio agent.
// No type in this package performs IO.
package domain

import "errors"

// ErrInvalidAmount is returned when ParseAmount cannot decode a string.
var ErrInvalidAmount = errors.New("invalid amount")

// Money is a precise decimal monetary value.
type Money struct{}

// ParseAmount parses a decimal string into a Money value.
func ParseAmount(_ string) (Money, error) {
	return Money{}, nil
}

// String returns the canonical decimal representation.
func (Money) String() string { return "" }
```

- [ ] **Step 8.5: Run the test and confirm RED on assertions**

```bash
go test ./internal/domain/... -v
```

Expected: FAIL — assertions on parsed values fail.

- [ ] **Step 8.6 (GREEN): Replace with the real implementation**

Replace `internal/domain/money.go` with:

```go
// Package domain holds the pure domain types of the portfolio agent.
// No type in this package performs IO.
package domain

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

// ErrInvalidAmount is returned when ParseAmount cannot decode a string.
var ErrInvalidAmount = errors.New("invalid amount")

// Money is a precise decimal monetary value. It wraps shopspring/decimal so
// callers cannot accidentally use float arithmetic on financial data.
type Money struct {
	amount decimal.Decimal
}

// ParseAmount parses a decimal string into a Money value.
// Empty strings and non-numeric inputs return ErrInvalidAmount.
func ParseAmount(s string) (Money, error) {
	if s == "" {
		return Money{}, fmt.Errorf("%w: empty string", ErrInvalidAmount)
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return Money{}, fmt.Errorf("%w: %q: %w", ErrInvalidAmount, s, err)
	}
	return Money{amount: d}, nil
}

// String returns the canonical decimal representation. The zero Money is "0".
func (m Money) String() string {
	return m.amount.String()
}
```

- [ ] **Step 8.7: Run the test and confirm GREEN**

```bash
go test ./internal/domain/... -v
```

Expected: all six tests PASS.

- [ ] **Step 8.8: Coverage check on the domain package**

```bash
go test -coverprofile=coverage.out ./internal/domain/...
go tool cover -func=coverage.out | tail -n 1
```

Expected: `total: (statements) ... %` with coverage ≥ 90% for `internal/domain`. (Both functions are exercised.)

- [ ] **Step 8.9: Run the linter on the new code**

```bash
golangci-lint run ./internal/domain/...
```

Expected: no issues.

- [ ] **Step 8.10: Commit**

```bash
git add go.mod go.sum internal/domain/money.go internal/domain/money_test.go
git commit -m "$(cat <<'EOF'
feat(domain): add Money type with TDD-driven parser

Demonstrates the project's TDD pattern on a foundational decimal value
that downstream P&L and persistence code will use. No floats on the
money path — uses shopspring/decimal underneath.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 9: GitHub Actions CI

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 9.1: Create the workflow**

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: v1.59.1
          args: --timeout=5m

  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: Run tests with race detector and coverage
        run: go test -race -coverprofile=coverage.out ./...

      - name: Coverage summary
        run: go tool cover -func=coverage.out

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: Build
        run: go build -o bin/agent ./cmd/agent
```

- [ ] **Step 9.2: Lint the workflow YAML locally (optional but helpful)**

```bash
# If `yamllint` is installed:
yamllint .github/workflows/ci.yml || true
```

Expected: clean or skipped. (Not a hard gate — CI itself is the source of truth.)

- [ ] **Step 9.3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "$(cat <<'EOF'
ci: add GitHub Actions workflow for lint + test + build

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 10: Polish README and final exit-gate verification

**Files:**
- Modify: `README.md` (replace the stub with a developer-ready quickstart).

- [ ] **Step 10.1: Replace `README.md` with the polished version**

Overwrite `README.md` with:

```markdown
# Portfolio Agent (go-tracker)

Personal investment portfolio monitoring agent in Go. Monitors CEDEARs, US
stocks, crypto, Argentine bonds, and FX rates (MEP / CCL / oficial / tarjeta).
Runs 24/7 on a VPS, keeps an SQLite store updated with prices and positions,
and exposes a React dashboard embedded in the binary.

> Read-only over the financial world — no trading.

## Status

Pre-MVP. Phase 0 (setup) complete. See `docs/superpowers/plans/` for the
active plan.

## Companion documents

- Requirements: [`docs/SRS_PortfolioAgent.md`](docs/SRS_PortfolioAgent.md)
- Architecture: [`docs/superpowers/specs/2026-05-12-portfolio-agent-architecture-design.md`](docs/superpowers/specs/2026-05-12-portfolio-agent-architecture-design.md)
- Project rules for Claude / agent tooling: [`CLAUDE.md`](CLAUDE.md)

## Quickstart (developers)

Requires Go 1.22+ and `make`. Optional: `golangci-lint` for linting.

```bash
git clone https://github.com/joajo13/go-tracker.git
cd go-tracker
cp .env.example .env  # edit if needed
make ci               # lint + test + build
./bin/agent           # starts on :8080
curl http://localhost:8080/healthz
```

### Windows note

The deploy target is Linux. For development on Windows, the project compiles
and runs natively (pure-Go SQLite — no CGO). Use Git Bash or WSL to invoke
`make` cleanly, or install GNU Make via `scoop install make`.

## Layout

```
cmd/agent/          entrypoint
internal/
  api/              HTTP handlers
  domain/           pure domain types (Money, Ticker, Operation, ...)
  sources/          external API adapters (Phase 1+)
  persistence/      repos + migrations (Phase 1+)
  scheduler/        poll scheduling (Phase 1+)
  workers/          worker pool (Phase 1+)
  broadcaster/      price event fan-out (Phase 1+)
  pnl/              FIFO P&L calculator (Phase 2+)
  alerts/           rule engine (Phase 3+)
  indicators/       RSI, MA, etc (post-MVP)
  downsampler/      historical aggregation (post-MVP)
  web/              embedded frontend (Phase 4+)
  calendar/         market holidays (Phase 1+)
web/                React + Vite frontend (Phase 4)
docs/               SRS + architecture spec + plans
migrations/         goose migrations (Phase 1+)
scripts/            dev helpers
```

## Development workflow

- TDD strict: failing test before any implementation. See
  `internal/domain/money_test.go` for the canonical pattern.
- Linting: `make lint` (uses the strict `.golangci.yml`).
- Coverage gates: ≥70% global, ≥90% in `pnl/`, `alerts/`, `indicators/`.
- Conventional commits: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`.

## License

MIT — see [`LICENSE`](LICENSE).
```

- [ ] **Step 10.2: Run the full local CI gate**

```bash
make ci
```

Expected: `golangci-lint` reports no issues, `go test ./...` is all green, and `go build` succeeds. If anything fails, fix inline before continuing.

- [ ] **Step 10.3: Run coverage to confirm the smoke targets**

```bash
make cover
```

Expected: `internal/domain` ≥ 90% coverage, `internal/api` ≥ 90% coverage (the only two real-code packages so far). Global ≥ 70%.

- [ ] **Step 10.4: Commit the README polish**

```bash
git add README.md
git commit -m "$(cat <<'EOF'
docs: flesh out README with quickstart and layout

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

- [ ] **Step 10.5: Confirm clean working tree**

```bash
git status
git log --oneline
```

Expected: working tree clean. Log shows the Phase 0 commits on `main` since the architecture design commit.

---

## Self-review checklist

Walk this list once after executing all tasks; fix anything that's red, then move on.

- [ ] **Spec coverage:** every Phase 0 deliverable in
  `docs/superpowers/specs/.../section-7 Phase 0` has a corresponding task.
  - `.gitignore`, `.editorconfig`, `LICENSE`, `README.md` → Task 1 + Task 10
  - `go.mod` with locked deps → Task 2 (init) + Task 6/7/8 (deps added per usage)
  - Folder scaffolding → Task 5
  - `.golangci.yml` → Task 3
  - `Makefile` (test/lint/build/run/cover/ci/web-build/mocks/migrate) →
    Task 4 (`migrate` target is intentionally deferred to Phase 1 when
    goose enters; the Makefile leaves a slot for it.)
  - `.github/workflows/ci.yml` → Task 9
  - `cmd/agent/main.go` with `/healthz` → Tasks 6 + 7
  - TDD smoke commit on real domain code → Task 8
  - `.env.example` → Task 1
- [ ] **Exit gate:** `make test lint build` runs green locally → Task 10
  step 10.2. CI green is verified after the first push to the remote.
- [ ] **No placeholders:** every step in this plan has runnable commands or
  full code. No "TBD".
- [ ] **Type consistency:** `HealthHandler`, `ParseAmount`, `Money`,
  `ErrInvalidAmount`, `newRouter` are spelled identically across test and
  impl files.
- [ ] **TDD discipline:** every code task starts with a failing test
  (Tasks 6, 7, 8). Tasks 1–5 and 9 are config-only and do not need TDD.
- [ ] **Engram saves:** after the plan finishes, save a short
  `mem_save` describing the locked module path, dep versions, and any
  surprises encountered during execution.

---

## Execution handoff

Plan complete and saved to
`docs/superpowers/plans/2026-05-12-portfolio-agent-phase-0-setup.md`.

Two execution options:

1. **Subagent-Driven** — dispatch a fresh subagent per task with two-stage
   review. Best when context pressure is high and tasks are independent.
2. **Inline Execution (recommended for this phase)** — single-thread the
   tasks in this session, commit per task, checkpoint after Tasks 6, 8,
   and 10. The tasks are tightly coupled (deps added per task, each commit
   green) so isolation per-task adds friction without value.

Default: **Inline Execution** unless redirected.
