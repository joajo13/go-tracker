# Portfolio Agent — Architecture Design

> Companion spec to `docs/SRS_PortfolioAgent.md` (v1.0).
> Resolves implementation-level decisions left open by the SRS.
> Author: Juan Giupponi
> Date: 2026-05-12

## 1. Context & scope

This document closes the implementation-level decisions the SRS leaves open:
how the scheduler/workers/sources/persistence wire together, how
decimals/timezones/repos/errors/sessions are modeled, and the day-1 testing
pattern.

It does **not** restate requirements — those live in the SRS.

**In scope**: backend modules end-to-end.
**Out of scope**: frontend component design (blocked on Claude Design handoff
bundle delivery — see §9).

## 2. Component design

### 2.1 Scheduler (`internal/scheduler`)
- **Purpose**: trigger polls per configured interval, respecting groupings.
- **Interface**: `Scheduler.Run(ctx) error` blocks; `Scheduler.Add(ticker)` /
  `Remove(id)` mutate the active set at runtime.
- **Internals**: `map[time.Duration]*group`, each group owns one `time.Ticker`
  and the list of ticker IDs in its interval. On tick, enqueues
  `PollJob{TickerID, Source}` per ticker into the worker channel.
- **Depends on**: `chan<- PollJob` (workers), `TickerRepo` for the active set,
  `Clock` interface for testability (never `time.Now()` directly).

### 2.2 Worker pool (`internal/workers`)
- **Purpose**: bounded concurrency for HTTP polls; rate-limit per source.
- **Interface**:
  `NewPool(n int, broadcaster *PriceBroadcaster, sources map[string]PriceSource) *Pool`;
  `Pool.Run(ctx context.Context, jobs <-chan PollJob) error`.
- **Internals**: N goroutines reading the job channel; each lookup source →
  `Fetch(ctx, symbol)` (rate-limited via `golang.org/x/time/rate` inside the
  adapter) → push `PriceEvent` to broadcaster.
- **Resilience**: each worker's outer loop wraps job execution in
  `defer recover()`; on panic the worker logs panic + stack and continues to
  the next job (the goroutine is not killed). Per-job context has a 10 s
  timeout default. Failed fetch logged; backoff handled inside the adapter.

### 2.3 Source adapters (`internal/sources`)
- **Purpose**: encapsulate one external API.
- **Interface**:
  ```go
  type PriceSource interface {
      Name() string
      Fetch(ctx context.Context, symbol string) (*domain.Price, error)
  }
  ```
- **One file per source**: `yahoo.go`, `dolarapi.go`, `finnhub.go`, `byma.go`,
  `bull.go`. Each owns its own `*rate.Limiter`.
- **Tests**: HTTP mocks via `httptest.Server`; never real network in CI. Golden
  JSON fixtures of real responses captured once.

### 2.4 PriceBroadcaster (`internal/broadcaster`)
- **Purpose**: fan-out price events to multiple consumers (persistence,
  alerts, SSE) without coupling them.
- **Interface**:
  ```go
  type Broadcaster interface {
      Subscribe(name string, buf int) <-chan PriceEvent
      Unsubscribe(name string)
      Publish(PriceEvent)
  }
  ```
- **Behavior**: drop-on-full per subscriber, log
  `metric=broadcaster_drop subscriber=X`. Never block the producer.

### 2.5 Persistence (`internal/persistence`)
- **Layout**:
  - `internal/persistence/sqlite/`: concrete repos (`PriceRepoSQLite`,
    `OperationRepoSQLite`, etc).
  - `internal/persistence/mocks/`: gomock-generated mocks.
  - `internal/domain/ports.go`: repo interfaces (`PriceRepo`, `OperationRepo`,
    `AlertRepo`, `TickerRepo`).
- **Decimal handling**: stored as `TEXT` in SQLite. Conversion at repo
  boundary via `decimal.NewFromString` / `Decimal.String()`. No floats anywhere
  on the money path.
- **Migrations**: `goose` from `migrations/` folder. Auto-run on app start.
- **Driver**: `modernc.org/sqlite` (pure Go, no CGO).

### 2.6 P&L (`internal/pnl`)
- **Purpose**: pure FIFO calculator, multi-currency.
- **Interface**:
  ```go
  func Calculate(
      ops []domain.Operation,
      lastPrices map[domain.TickerID]domain.Price,
      fx domain.FxSnapshot,
  ) domain.PnLReport
  ```
- **Behavior**: FIFO lots per ticker, commissions added to cost / subtracted
  from proceeds. USD conversions use the FX snapshot stored on each operation
  (RF-PL-03).
- **Side-effects**: zero. Coverage target ≥ 95%.

### 2.7 Alerts (`internal/alerts`)
- **Rule expression**: JSON AST stored in `alert_rules.expression`. Grammar:
  - Leaves: `price_above`, `price_below`, `pct_change`, `crosses_ma`, `rsi`,
    `volume_anom`, `spread`, `portfolio_change`.
  - Operators: `and`, `or`.
- **Evaluator**: pure function
  `Evaluate(rule Rule, ctx EvalContext) (triggered bool, observed map[string]any, err error)`.
  `EvalContext` carries price history snapshots passed in by the caller —
  evaluator never fetches.
- **Cooldown**: enforced by the caller service, not the pure evaluator.

### 2.8 Indicators (`internal/indicators`)
Pure functions: `MA(prices, window)`, `RSI(prices, window)`,
`VolumeAvg(volumes, window)`. Independently tested. Coverage ≥ 95%.

### 2.9 HTTP API (`internal/api`)
- **Router**: `chi` v5.
- **Mount points**:
  - `/api/login`, `/api/logout` (no auth)
  - `/api/v1/{tickers,operations,alerts,pnl}` (auth required)
  - `/api/v1/events` (SSE, auth required)
  - `/healthz` (no auth)
  - `/` → static frontend (`embed.FS`, SPA fallback to `index.html`)
- **Auth middleware**: cookie `pa_session` signed with HMAC. Session store:
  in-memory map in process memory; lost on restart, acceptable for single user.
- **Validation**: `go-playground/validator` on DTOs.
- **Error mapping**: sentinel errors per domain → `errors.Is` switch → HTTP
  status + JSON body `{"error":"...","code":"..."}`.

### 2.10 Config (`internal/config`)
- `caarlos0/env` parses env into a `Config` struct on boot.
- **Required**: `DB_PATH`, `DASHBOARD_PASSWORD_HASH`.
- **Optional with defaults**: `HTTP_ADDR=:8080`, `LOG_LEVEL=info`,
  `LOG_FORMAT=json`, `WORKER_POOL_SIZE=10`.
- **Secrets-only env**: API keys, broker credentials.
- **Mutable runtime config** (per-ticker intervals, alert rules) lives in
  SQLite, never in env.

### 2.11 Frontend embed (`internal/web`)
```go
//go:embed all:dist
var dist embed.FS
```
- Served via `http.FileServer` wrapped to fall back to `index.html` for SPA
  routes.
- `Makefile` target `web-build` runs `npm --prefix web run build` before
  `go build`.

## 3. Data flow — price tick lifecycle

```
time.Ticker → Scheduler.tickGroup → chan PollJob
  → Worker pool → adapter.Fetch → PriceEvent
  → Broadcaster.Publish
     ├→ PriceRepo.Insert            (writes to prices_1m)
     ├→ AlertEvaluator              (evaluates rules touching this ticker)
     └→ SSEHub.Broadcast            (pushes to connected dashboard clients)
```

All channels buffered. If a subscriber's buffer is full, the event is dropped
for that subscriber and a metric is logged; the producer never blocks.

## 4. Error model

- **Domain errors**: sentinel vars (`ErrTickerNotFound`,
  `ErrInvalidOperation`, `ErrInsufficientHistory`, …).
- **HTTP boundary**: middleware catches errors → maps to status via
  `errors.Is` chain → returns JSON.
- **Workers/goroutines**: every long-lived goroutine wraps its inner loop
  iteration in `defer recover()` so a panic logs panic + stack and the
  goroutine resumes with the next iteration. Process is not torn down.

## 5. Time & timezones

- `time.Time` everywhere; **always UTC in storage**.
- Markets: `MarketNYSE` (`America/New_York`), `MarketBYMA`
  (`America/Argentina/Buenos_Aires`), `MarketCrypto` (always open).
- `IsMarketOpen(market Market, t time.Time) bool` reads embedded holiday
  JSON (`internal/calendar/holidays.json`).
- Tests inject a fake `Clock` — business code never calls `time.Now()`
  directly.

## 6. Testing strategy

- **TDD strict** from commit #1: red → green → refactor. A failing test
  exists before any implementation file.
- **Layering**:
  - **Unit**: domain / pnl / alerts / indicators / parsers — pure, fast,
    full suite < 1 s.
  - **Integration**: repos against in-memory SQLite (`:memory:`),
    broadcaster, scheduler with fake clock.
  - **HTTP**: `httptest.Server` smoke tests per route.
- **Mocking**: `uber-go/mock` for repo interfaces; adapters mocked via
  `httptest`.
- **Coverage gates** (CI-enforced via `go test -coverprofile` + a small
  script comparing per-package coverage):
  - Global ≥ 70%
  - `pnl/`, `alerts/`, `indicators/`: ≥ 90%
- **Fixtures**: golden JSON files for adapter parsers — real API responses
  captured once, replayed in tests.

## 7. Phased plan

### Phase 0 — Setup (immediate)
**Deliverables:**
1. `.gitignore`, `.editorconfig`, `LICENSE` (MIT), `README.md` with setup +
   dev workflow.
2. `go.mod` with all locked deps.
3. Folder scaffolding per SRS §5.4 (empty `.gitkeep` where needed to track
   empty dirs).
4. `.golangci.yml` strict preset (errcheck, gosec, gocritic, revive, govet,
   staticcheck, gosimple, ineffassign, unused, …).
5. `Makefile` targets: `test`, `lint`, `build`, `run`, `mocks`, `cover`,
   `web-build`, `migrate`.
6. `.github/workflows/ci.yml` running lint + test + coverage check on
   push / PR.
7. `cmd/agent/main.go`: loads config, inits slog, starts chi on `:8080`
   with `/healthz` only.
8. **TDD smoke commit**: a real pure function in `internal/domain` written
   red → green → refactor (target: a `Money`/`ParseAmount` helper around
   `shopspring/decimal`). Demonstrates the testing pattern that the rest of
   the codebase will follow.
9. `.env.example` committed, `.env` gitignored.

**Exit gate**: `make test lint build` green locally + CI green.

### Phase 1 — Ingest + persistence
- Goose migrations (`tickers`, `prices_1m`, `prices_1h`, `prices_1d`).
- Domain types + repos behind interfaces, with full test coverage.
- Yahoo + dolarapi adapters with `httptest` mocks.
- Scheduler + worker pool + broadcaster wired end-to-end.
- Integration test: `:memory:` DB + fake clock + mocked HTTP → verify a
  price lands in `prices_1m`.

### Phase 2 — P&L + operations
- `Operation` domain + repo.
- P&L calculator (FIFO multi-currency).
- CRUD endpoints for operations.
- CSV import.

### Phase 3 — Alerts (basic)
- `AlertRule` + evaluator (`price_above`, `pct_change` leaves; `and`/`or`).
- Trigger persistence with cooldown.
- CRUD endpoints for rules.

### Phase 4 — Dashboard MVP
**Blocked on design bundle delivery.** Once unblocked:
- Vite project under `web/` with React + Tailwind.
- Embed via `internal/web/embed.go`.
- TDD on logic-bearing components: formatters (decimal → display string),
  P&L breakdown, FIFO lot visualization.
- Visual review against the design bundle on the running app.

## 8. Decision log

| # | Decision | Alternatives | Why |
|---|----------|--------------|-----|
| D1 | One `time.Ticker` per interval group, shared worker pool | One goroutine per ticker | Respects per-source rate limits; scales without N goroutines. |
| D2 | PriceBroadcaster with drop-on-full | Direct calls / sync fan-out | Decouples consumers; slow SSE clients don't stall persistence. |
| D3 | Decimal as `TEXT` in SQLite | Integer cents | Variable precision (crypto, FX) without scaling pain. |
| D4 | UTC in storage, market tz at presentation | Store local | Single source of truth; no DST bugs in queries. |
| D5 | Repos behind interfaces in `internal/domain`, mocks via uber-go/mock | Hand-rolled fakes | Lower maintenance, idiomatic. |
| D6 | P&L as pure function | Stateful service | Trivially testable; aligns with SRS §5.2.6. |
| D7 | Alert rules as JSON AST | DSL parser | Simpler MVP; AST can grow later. |
| D8 | Sentinel errors per domain | Typed error structs | Lower ceremony; `errors.Is` is enough. |
| D9 | In-memory session store | Persisted sessions | Single user, single instance — restart = re-login is acceptable. |
| D10 | Frontend via `embed.FS` | Separate static server | Deploy = one binary (per SRS ADR-004). |

## 9. Open issues

- **Design bundle delivery**. The handoff URL provided
  (`api.anthropic.com/v1/design/h/CSms5EH9KDTiKp4n5_8NnA?…`) returned
  47.1 KB of binary/gzip content that `WebFetch` could not parse, and the
  URL pattern does not match a documented public Anthropic endpoint.
  Frontend work (Phase 4) is blocked until a readable bundle is available.
  Mitigation: backend (Phase 0–3) is fully independent and proceeds in
  parallel.

## 10. References

- `docs/SRS_PortfolioAgent.md` v1.0 (requirements source of truth).
- `CLAUDE.md` (project rules — TDD strict, no floats, no CGO, …).
- ADRs: SRS §11.A.
