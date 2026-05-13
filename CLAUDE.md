# Portfolio Agent — Project Instructions for Claude Code

## Project Context

Personal investment portfolio monitoring agent in Go. Monitors CEDEARs, US stocks, crypto, Argentine bonds, and FX rates (MEP/CCL/oficial/tarjeta). Runs 24/7 on a VPS, keeps SQLite updated with prices and positions, exposes a React dashboard embedded in the binary.

Full spec lives in `docs/SRS.md`. Read it before making architectural decisions.

## Owner

Juan Giupponi. Single-user system, no multi-tenancy needed.

## Tech stack (locked)

- Go 1.22+
- chi router for HTTP
- modernc.org/sqlite (pure Go, no CGO)
- shopspring/decimal for ALL monetary values (never float)
- log/slog for structured logging (stdlib)
- goose for migrations
- testify for assertions
- uber-go/mock for mocks
- React + Vite + TypeScript + Tailwind for frontend
- embed.FS to bundle frontend in the Go binary

## Non-negotiable rules

1. **TDD strict.** Red-green-refactor. Tests must fail before implementation exists.
2. **No floats for money.** Ever. Use shopspring/decimal everywhere or integer cents.
3. **No CGO dependencies.** Cross-compilation must stay trivial.
4. **No secrets in repo.** Env vars only. `.env.example` is committed, `.env` is gitignored.
5. **Linting must pass.** golangci-lint with the project config on every commit.
6. **Coverage targets.** >70% global, >90% on pnl/, alerts/, indicators/.
7. **Pure functions for business logic.** P&L calculator, rule evaluator, indicators are side-effect free and trivially testable.
8. **One source adapter per file.** No mega-files. Each external source isolated.

## Architectural patterns

- Operations are events. Positions are derived (never stored as mutable balance).
- Repositories live behind interfaces for mockability.
- All time values use `time.Time` with explicit `time.Location` (never `time.Local`).
- Markets timezone, NY for US/crypto, AR for ByMA/bonds.
- Context cancellation propagated through all goroutines.

## Folder conventions

- `cmd/agent/` — entrypoint only, no logic.
- `internal/domain/` — pure domain types, no IO.
- `internal/sources/` — external API adapters.
- `internal/persistence/` — repos + migrations.
- `internal/pnl/`, `internal/alerts/`, `internal/indicators/` — pure logic, highest coverage.
- `internal/api/` — HTTP handlers, thin layer over services.
- `web/` — React app, built into `internal/web/embed.go` via embed.FS.

## Commit hygiene

- Conventional commits: feat, fix, refactor, test, docs, chore.
- Each commit passes lint + tests.
- No "WIP" commits on main. Use feature branches.

## When in doubt

- Ask before assuming. Especially about financial calculations or broker-specific behavior.
- Prefer the simpler solution that fits the SRS over clever abstractions.
- If a third-party lib seems necessary, justify it against stdlib alternative first.
