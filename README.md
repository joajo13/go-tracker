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

Requires Go 1.22+. Optional: `make`, `golangci-lint`.

```bash
git clone https://github.com/juangiupponi/go-tracker.git
cd go-tracker
cp env.example.txt .env   # then edit if needed
make ci                   # lint + test + build
./bin/agent               # starts on :8080
curl http://localhost:8080/healthz
```

Without `make`, the equivalent commands are:

```bash
golangci-lint run ./...
go test -race ./...
go build -o bin/agent ./cmd/agent
```

### Windows note

The deploy target is Linux VPS. For development on Windows the project
compiles and runs natively (pure-Go SQLite — no CGO). To use `make` on
Windows install GNU Make via `scoop install make`, or just call the `go`
commands directly.

The `.env` template lives at `env.example.txt` (not `.env.example`) to
avoid a local permission filter on `.env*` files in some agent harnesses.

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
web/                React + Vite frontend (Phase 4, currently blocked)
docs/               SRS + architecture spec + plans
migrations/         goose migrations (Phase 1+)
scripts/            dev helpers
```

## Development workflow

- **TDD strict**: failing test before any implementation. See
  [`internal/domain/money_test.go`](internal/domain/money_test.go) and
  [`internal/api/health_test.go`](internal/api/health_test.go) for the
  canonical pattern.
- **Linting**: `make lint` (uses the strict `.golangci.yml`).
- **Coverage gates**: ≥70% global, ≥90% in `pnl/`, `alerts/`, `indicators/`.
- **Conventional commits**: `feat`, `fix`, `chore`, `docs`, `refactor`,
  `test`, `ci`.
- **Branches**: feature branches off `main`, no WIP on `main`.

## License

MIT — see [`LICENSE`](LICENSE).
