# Portfolio Agent — Phase 4 (Dashboard) Implementation Plan

> **For agentic workers:** Use `superpowers:subagent-driven-development` or `superpowers:executing-plans`. Checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port the Claude Design handoff bundle (`design-bundle-6474/go-tracker/project/`) into a real Vite + React + TypeScript app under `web/`, embed it in the Go binary via `embed.FS`, and ship all 7 views from the prototype.

**Architecture:**
- `web/` is a Vite + React 18 + TS app. Build artifact goes to `web/dist/`.
- `internal/web/embed.go` exports `dist embed.FS` via `//go:embed all:dist`.
- `cmd/agent/main.go` mounts the embed at `/` with SPA fallback to `index.html`.
- Design system is the bundle's `styles.css` imported verbatim. Tailwind is OUT.
- State-based routing inside `App.tsx` (no react-router for MVP — mirror the prototype's structure; revisit later if URL-shareable routes are needed).
- Mock data lives in `web/src/mock/` as TS modules. Real API calls replace mocks per-view in Phase 4c+ as backend Phase 1-3 lands.
- TDD applies to logic-bearing helpers (formatters, FIFO display helpers, expression preview). Pure presentational components do not need tests.

**Tech Stack:**
- Vite 5 + React 18 + TypeScript 5
- @fontsource/ibm-plex-mono (self-hosted, no CDN dependency at runtime)
- vitest + @testing-library/react for tests
- No CSS framework (using the bundle's CSS variables + class system)

**Companion docs:**
- Requirements: `docs/SRS_PortfolioAgent.md` (§3.7 Dashboard, §5.2.8 Frontend)
- Architecture: `docs/superpowers/specs/2026-05-12-portfolio-agent-architecture-design.md` (§2.11 Frontend embed)
- Design bundle: extracted at `%TEMP%/design-bundle-6474/` (this session); reference paths below.

---

## Scope phasing

This plan covers **all 7 views**. It is split into three executable phases. Each phase ends with a green build + commit; merge to main happens after Phase 4c (or earlier checkpoints if the user wants to stop).

- **Phase 4a — Foundation** (this session, ~2 hrs): Vite scaffolding, design system import, mock data port, App shell, routing, embed.FS wiring, stub views. Exit gate: `go build && ./bin/agent` serves the dashboard skeleton at `/`.
- **Phase 4b — MVP views** (subsequent session, ~3-4 hrs): Overview, Cartera, Operaciones, Alertas — all with mock data. CRUD modals where applicable.
- **Phase 4c — Post-MVP views** (subsequent session, ~3-4 hrs): Tickers CRUD, Gráficos (charts + indicators), Config.

---

## File Structure (target — Phase 4a creates the highlighted paths)

```
web/
├── package.json
├── tsconfig.json
├── tsconfig.node.json
├── vite.config.ts
├── index.html
├── public/
│   └── favicon.svg
├── src/
│   ├── main.tsx              ── React root + style imports
│   ├── App.tsx               ── shell, state routing, live-tick effect
│   ├── styles.css            ── verbatim from bundle (or src/index.css)
│   ├── types/
│   │   └── domain.ts         ── Ticker, Operation, Position, Alert, FxSnapshot, ...
│   ├── mock/
│   │   ├── data.ts           ── port of mock-data.js
│   │   └── positions.ts      ── port of MOCK.positions() FIFO calc
│   ├── lib/
│   │   ├── format.ts         ── currency, percent, decimal formatters (TDD)
│   │   ├── format.test.ts
│   │   └── routing.ts        ── Route enum + breadcrumb titles
│   ├── components/
│   │   ├── Icon.tsx          ── SVG icon set (port from prototype)
│   │   ├── Sidebar.tsx
│   │   ├── TopBar.tsx
│   │   ├── CurrencyPill.tsx  ── (used in many views — small reusable)
│   │   ├── KPI.tsx
│   │   ├── Card.tsx
│   │   ├── Chip.tsx
│   │   └── ...               ── more in 4b
│   └── views/                ── 7 view stubs in 4a; filled in 4b/4c
│       ├── Overview.tsx
│       ├── Portfolio.tsx
│       ├── Operations.tsx
│       ├── Tickers.tsx
│       ├── Charts.tsx
│       ├── Alerts.tsx
│       └── Config.tsx
└── dist/                     ── vite build output (gitignored)

internal/web/
└── embed.go                  ── //go:embed all:dist + Handler() http.Handler
```

The existing `.gitkeep` markers in `internal/web/` and `web/` get removed as real files land.

---

## Phase 4a — Foundation

### Task 4a-1: Scaffold Vite + React + TS in web/

**Files:**
- Create: `web/package.json`, `web/tsconfig.json`, `web/tsconfig.node.json`, `web/vite.config.ts`, `web/index.html`, `web/.gitignore` (additive to repo root .gitignore).

- [ ] **Step 1: Remove the `.gitkeep` placeholder in `web/`**

```bash
rm web/.gitkeep
```

- [ ] **Step 2: Bootstrap with create-vite (React + TS template)**

Run from the project root:

```bash
npm create vite@5.5.5 -- --template react-ts web -y
```

This creates the standard Vite React+TS skeleton at `web/`.

- [ ] **Step 3: Add the testing + font dependencies**

```bash
cd web
npm install
npm install -D vitest @testing-library/react @testing-library/jest-dom @testing-library/user-event jsdom @types/node
npm install @fontsource/ibm-plex-mono
cd ..
```

- [ ] **Step 4: Update `vite.config.ts` to enable vitest**

Replace `web/vite.config.ts` contents with:

```ts
/// <reference types="vitest" />
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  base: '/',
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    sourcemap: false,
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test-setup.ts'],
  },
});
```

- [ ] **Step 5: Create `web/src/test-setup.ts`**

```ts
import '@testing-library/jest-dom/vitest';
```

- [ ] **Step 6: Add a Makefile target for the frontend**

Replace the `web-build` target in the project root `Makefile` with:

```makefile
web-build: ## Build the embedded frontend
	cd web && npm install && npm run build

web-dev: ## Run frontend dev server (Vite)
	cd web && npm run dev

web-test: ## Run frontend tests once
	cd web && npm test -- --run

web-test-watch: ## Run frontend tests in watch mode
	cd web && npm test
```

(`web-test` and `web-test-watch` are new; keep the other targets unchanged.)

- [ ] **Step 7: Verify dev tools work**

```bash
cd web
npm run build 2>&1 | tail -5
cd ..
ls -la web/dist/index.html
```

Expected: `index.html` exists in `web/dist/`.

- [ ] **Step 8: Commit**

```bash
git add web/ Makefile
git rm web/.gitkeep 2>/dev/null || true
git status --short
git commit -m "$(cat <<'EOF'
chore(web): scaffold Vite + React + TS in web/

Adds the frontend project skeleton with vitest, jsdom, and
@testing-library wired up. Makefile gains web-build / web-dev /
web-test / web-test-watch.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 4a-2: Port design system CSS and fonts

**Files:**
- Create: `web/src/styles.css` (verbatim port of bundle's `styles.css`)
- Modify: `web/src/main.tsx` (import fonts + styles)
- Replace: `web/src/App.tsx` (will be rewritten in 4a-5)

- [ ] **Step 1: Copy the design system CSS into the project**

Source: `%TEMP%/design-bundle-6474/go-tracker/project/styles.css` (686 lines, untouched).

Write the file to `web/src/styles.css`. (Do this via the harness's Write — full content provided in the bundle.)

- [ ] **Step 2: Remove the create-vite scaffold CSS that conflicts**

```bash
rm web/src/App.css web/src/index.css
```

- [ ] **Step 3: Update `web/src/main.tsx`**

```tsx
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import '@fontsource/ibm-plex-mono/400.css';
import '@fontsource/ibm-plex-mono/500.css';
import '@fontsource/ibm-plex-mono/600.css';
import './styles.css';
import { App } from './App';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
```

- [ ] **Step 4: Replace `web/index.html`** (root, not src)

```html
<!doctype html>
<html lang="es">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/favicon.svg" />
    <meta name="viewport" content="width=1440" />
    <title>go-tracker — Dashboard</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

- [ ] **Step 5: Minimal `App.tsx` stub so build passes**

```tsx
export function App() {
  return <div className="app">{/* shell coming in 4a-5 */}</div>;
}
```

- [ ] **Step 6: Verify the build**

```bash
cd web && npm run build && cd ..
ls -la web/dist/
```

Expected: `web/dist/index.html` + asset bundle.

- [ ] **Step 7: Commit**

```bash
git add web/src/styles.css web/src/main.tsx web/src/App.tsx web/index.html
git rm web/src/App.css web/src/index.css 2>/dev/null || true
git commit -m "$(cat <<'EOF'
feat(web): import design system CSS and IBM Plex Mono

Verbatim port of the bundle's styles.css (~700 LOC of CSS variables
and component classes). Self-hosted font via @fontsource so the
production binary has no network dependency at runtime.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 4a-3: TypeScript domain types + mock data port

**Files:**
- Create: `web/src/types/domain.ts`
- Create: `web/src/mock/data.ts`
- Create: `web/src/mock/positions.ts`

- [ ] **Step 1: `web/src/types/domain.ts`**

```ts
export type TickerType = 'cedear' | 'us_stock' | 'crypto' | 'bond' | 'fx';

export type Currency = 'ARS' | 'USD_CCL' | 'USD_OFI' | 'USD_TARJ';

export interface Ticker {
  symbol: string;
  name: string;
  type: TickerType;
  ratio?: number;            // CEDEARs only
  underlying?: string;       // CEDEARs only
  pollIntervalSeconds: number;
  sources: string[];
  active: boolean;
}

export interface Price {
  symbol: string;
  px: number;
  change24h: number;         // percent
  spark: number[];           // last 30 closes for sparkline
  ts: number;                // unix ms
}

export interface FxSnapshot {
  ccl: number;
  ofi: number;
  tarj: number;
  mep: number;
  ts: number;
}

export interface Operation {
  id: string;
  symbol: string;
  type: 'BUY' | 'SELL';
  ts: number;
  quantity: number;
  unitPrice: number;
  currency: Currency;
  commission: number;
  marketFees: number;
  broker?: string;
  source: 'manual' | 'broker_sync' | 'csv_import';
  notes?: string;
}

export interface Position {
  symbol: string;
  ticker: Ticker;
  quantity: number;
  avgCost: number;           // FIFO, in operation currency
  costBasis: number;
  marketValue: number;
  pnlAbs: number;            // unrealized, in display currency
  pnlPct: number;
  weight: number;            // % of portfolio
}

export interface AlertRule {
  id: string;
  name: string;
  symbol?: string;           // null = portfolio-global
  expressionJSON: unknown;   // AST — see internal/alerts spec
  cooldownSeconds: number;
  active: boolean;
}

export interface AlertEvent {
  id: string;
  ruleId: string;
  ruleName: string;
  ts: number;
  severity: 'info' | 'warn' | 'crit';
  observed: Record<string, string | number>;
  status: 'new' | 'seen' | 'archived';
}

export interface SourceHealth {
  name: string;
  status: 'healthy' | 'degraded' | 'down';
  lastSuccessMs: number;
  errors24h: number;
  rateLimitPct: number;
}

export interface LogEntry {
  ts: number;
  level: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';
  msg: string;
  meta: Record<string, unknown>;
}

export type Route = 'overview' | 'portfolio' | 'operations' | 'tickers' | 'charts' | 'alerts' | 'config';
```

- [ ] **Step 2: `web/src/mock/data.ts`**

Port the bundle's `mock-data.js` to TypeScript. The source file lists 20 tickers, 20 operations, FX rates, 7 alerts, source healths, and log entries. Structure:

```ts
import type {
  Ticker, Price, Operation, AlertEvent, AlertRule, FxSnapshot, SourceHealth, LogEntry,
} from '../types/domain';

export const TICKERS: Ticker[] = [/* 20 entries — see bundle mock-data.js */];
export const LATEST: Record<string, Price> = {/* 20 entries */};
export const OPERATIONS: Operation[] = [/* 20 entries */];
export const ALERTS: AlertEvent[] = [/* 7 entries */];
export const ALERT_RULES: AlertRule[] = [/* 5 entries */];
export const FX: FxSnapshot = {/* current rates */};
export const SOURCES: SourceHealth[] = [/* 6 entries */];
export const LOGS: LogEntry[] = [/* recent entries */];
```

The full data values are large — copy them straight from `%TEMP%/design-bundle-6474/go-tracker/project/mock-data.js` and adapt the syntax (object keys quoted as needed for TS, types annotated). Do NOT paraphrase or "improve" the data values — pixel-perfect downstream depends on the numbers staying intact.

- [ ] **Step 3: `web/src/mock/positions.ts`**

Port the `MOCK.positions()` FIFO + valuation function from the bundle:

```ts
import { TICKERS, OPERATIONS, LATEST, FX } from './data';
import type { Position, Currency } from '../types/domain';

export function computePositions(displayCcy: Currency = 'USD_CCL'): Position[] {
  // FIFO lots per symbol, then valuation at LATEST[symbol].px,
  // converted to displayCcy using FX. Mirror the prototype logic.
}
```

Reference: the function lives in `mock-data.js` between the LATEST/OPERATIONS definitions and the bottom of the file. Port the algorithm 1-for-1.

- [ ] **Step 4: Verify TS compiles**

```bash
cd web && npx tsc --noEmit && cd ..
```

Expected: no output, exit 0.

- [ ] **Step 5: Commit**

```bash
git add web/src/types web/src/mock
git commit -m "$(cat <<'EOF'
feat(web): port TS domain types and mock data from design bundle

Establishes the TS types that mirror the Go domain (until codegen).
Mock data ported verbatim from mock-data.js so views can render
real-looking content before the backend API lands in Phase 1-3.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 4a-4: TDD formatters

**Files:**
- Create: `web/src/lib/format.ts`
- Create: `web/src/lib/format.test.ts`

This is the first TDD-discipline commit on the frontend. Demonstrates the pattern that view-level logic (e.g., FIFO breakdown rendering, expression preview) will follow.

- [ ] **Step 1 (RED): Write the failing test**

`web/src/lib/format.test.ts`:

```ts
import { describe, it, expect } from 'vitest';
import { formatMoney, formatPercent, formatTicks } from './format';

describe('formatMoney', () => {
  it('formats USD with 2 decimals and group separators', () => {
    expect(formatMoney(1234567.89, 'USD_CCL')).toBe('$ 1,234,567.89');
  });
  it('formats ARS with 0 decimals (no cents for large amounts)', () => {
    expect(formatMoney(1234567.89, 'ARS')).toBe('AR$ 1,234,568');
  });
  it('handles negative values with a leading minus', () => {
    expect(formatMoney(-42.5, 'USD_CCL')).toBe('-$ 42.50');
  });
  it('returns "—" for non-finite input', () => {
    expect(formatMoney(NaN, 'USD_CCL')).toBe('—');
    expect(formatMoney(Infinity, 'USD_CCL')).toBe('—');
  });
});

describe('formatPercent', () => {
  it('renders 2 decimals with a sign for positive', () => {
    expect(formatPercent(3.4567)).toBe('+3.46%');
  });
  it('renders negative without a redundant double sign', () => {
    expect(formatPercent(-1.2)).toBe('-1.20%');
  });
  it('returns "0.00%" for zero (no sign)', () => {
    expect(formatPercent(0)).toBe('0.00%');
  });
});

describe('formatTicks', () => {
  it('humanizes a unix-ms timestamp as a short HH:MM:SS', () => {
    const ts = Date.UTC(2026, 4, 13, 12, 34, 56);
    expect(formatTicks(ts)).toMatch(/\d{2}:\d{2}:\d{2}/);
  });
});
```

- [ ] **Step 2: Run the test and confirm compile-error RED**

```bash
cd web && npm test -- --run 2>&1 | tail -20 && cd ..
```

Expected: vitest fails — cannot resolve `./format`.

- [ ] **Step 3 (GREEN-MINIMAL): Stub that compiles but fails**

`web/src/lib/format.ts`:

```ts
import type { Currency } from '../types/domain';

export function formatMoney(_amount: number, _ccy: Currency): string { return ''; }
export function formatPercent(_pct: number): string { return ''; }
export function formatTicks(_tsMs: number): string { return ''; }
```

Run tests: assertion-level RED.

- [ ] **Step 4 (GREEN): Real implementation**

```ts
import type { Currency } from '../types/domain';

const SYMBOLS: Record<Currency, string> = {
  ARS: 'AR$',
  USD_CCL: '$',
  USD_OFI: '$',
  USD_TARJ: '$',
};

function isFinite(x: number): boolean {
  return Number.isFinite(x);
}

export function formatMoney(amount: number, ccy: Currency): string {
  if (!isFinite(amount)) return '—';
  const sym = SYMBOLS[ccy];
  const fractionDigits = ccy === 'ARS' ? 0 : 2;
  const absStr = Math.abs(amount).toLocaleString('en-US', {
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  });
  return `${amount < 0 ? '-' : ''}${sym} ${absStr}`;
}

export function formatPercent(pct: number): string {
  if (!isFinite(pct)) return '—';
  if (pct === 0) return '0.00%';
  const sign = pct > 0 ? '+' : '-';
  return `${sign}${Math.abs(pct).toFixed(2)}%`;
}

export function formatTicks(tsMs: number): string {
  const d = new Date(tsMs);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}
```

- [ ] **Step 5: Tests pass**

```bash
cd web && npm test -- --run && cd ..
```

Expected: all green.

- [ ] **Step 6: Commit**

```bash
git add web/src/lib
git commit -m "$(cat <<'EOF'
feat(web): add money/percent/tick formatters with TDD

First TDD-discipline commit on the frontend. Establishes the pattern
that view-level logic (FIFO breakdown, expression preview, etc.) will
follow as views land in Phase 4b/4c.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 4a-5: App shell + state routing

**Files:**
- Modify: `web/src/App.tsx`
- Create: `web/src/lib/routing.ts`

- [ ] **Step 1: `web/src/lib/routing.ts`**

```ts
import type { Route } from '../types/domain';

export const ROUTE_TITLES: Record<Route, [string, string, string?]> = {
  overview:   ['Cartera', 'Overview'],
  portfolio:  ['Cartera', 'Posiciones'],
  operations: ['Cartera', 'Operaciones'],
  tickers:    ['Mercado', 'Tickers'],
  charts:     ['Mercado', 'Gráficos'],
  alerts:     ['Sistema', 'Alertas'],
  config:     ['Sistema', 'Configuración'],
};
```

- [ ] **Step 2: `web/src/App.tsx`**

Replace the stub from 4a-2 with the real shell. Mirrors the prototype's `app.jsx` but typed:

```tsx
import { useState, useEffect, useMemo, useCallback } from 'react';
import type { Route, Currency, Position } from './types/domain';
import { FX, LATEST } from './mock/data';
import { computePositions } from './mock/positions';
import { Sidebar } from './components/Sidebar';
import { TopBar } from './components/TopBar';
import { Overview } from './views/Overview';
import { Portfolio } from './views/Portfolio';
import { Operations } from './views/Operations';
import { Tickers } from './views/Tickers';
import { Charts } from './views/Charts';
import { Alerts } from './views/Alerts';
import { Config } from './views/Config';

export interface AppContext {
  ccy: Currency;
  setCcy: (c: Currency) => void;
  range: string;
  setRange: (r: string) => void;
  positions: Position[];
  nav: (to: Route, params?: Record<string, unknown>) => void;
  params: Record<string, unknown>;
  route: Route;
  toast: (msg: string) => void;
  lastTick: number;
}

export function App() {
  const [route, setRoute] = useState<Route>('overview');
  const [params, setParams] = useState<Record<string, unknown>>({});
  const [ccy, setCcy] = useState<Currency>('USD_CCL');
  const [range, setRange] = useState('1M');
  const [toastMsg, setToastMsg] = useState<string | null>(null);
  const [lastTick, setLastTick] = useState(Date.now());

  // Live tick simulation — Phase 4c will replace with SSE
  useEffect(() => {
    const id = setInterval(() => {
      const syms = Object.keys(LATEST);
      for (let i = 0; i < 3; i++) {
        const s = syms[Math.floor(Math.random() * syms.length)];
        const l = LATEST[s];
        if (l) l.px += (Math.random() - 0.5) * 0.0025 * l.px;
      }
      setLastTick(Date.now());
    }, 3500);
    return () => clearInterval(id);
  }, []);

  const positions = useMemo(() => computePositions(ccy), [ccy, lastTick]);

  const nav = useCallback((to: Route, p: Record<string, unknown> = {}) => {
    setRoute(to);
    setParams(p);
    document.querySelector('.main')?.scrollTo(0, 0);
  }, []);

  const toast = useCallback((msg: string) => {
    setToastMsg(msg);
    setTimeout(() => setToastMsg(null), 2200);
  }, []);

  const ctx: AppContext = { ccy, setCcy, range, setRange, positions, nav, params, route, toast, lastTick };

  return (
    <div className="app">
      <Sidebar active={route} onNavigate={nav} />
      <TopBar route={route} fx={FX} lastTick={lastTick} />
      <main className="main">
        {route === 'overview'   && <Overview ctx={ctx} />}
        {route === 'portfolio'  && <Portfolio ctx={ctx} />}
        {route === 'operations' && <Operations ctx={ctx} />}
        {route === 'tickers'    && <Tickers ctx={ctx} />}
        {route === 'charts'     && <Charts ctx={ctx} />}
        {route === 'alerts'     && <Alerts ctx={ctx} />}
        {route === 'config'     && <Config ctx={ctx} />}
      </main>
      {toastMsg && <div className="toast"><span>{toastMsg}</span></div>}
    </div>
  );
}
```

- [ ] **Step 3: Skip build verification here**

The components and views referenced don't exist yet — they land in 4a-6 and 4a-7. The build will fail until those tasks complete. Don't commit yet.

---

### Task 4a-6: Layout components (Sidebar, TopBar, Icon)

**Files:**
- Create: `web/src/components/Icon.tsx`
- Create: `web/src/components/Sidebar.tsx`
- Create: `web/src/components/TopBar.tsx`

- [ ] **Step 1: `web/src/components/Icon.tsx`**

Port the prototype's `Icon({ name, size = 14 })` component from `components.jsx`. The bundle has icons for: `chart`, `coin`, `clock`, `bell`, `cog`, `sliders`, `box`, `plug`, `download`, `upload`, `plus`, `check`, `x`, `search`, `arrow-up`, `arrow-down`, etc.

Replicate the SVG paths verbatim. Type:

```tsx
export interface IconProps {
  name: string;
  size?: number;
}

export function Icon({ name, size = 14 }: IconProps) {
  // switch (name) { case 'chart': return <svg ...>...</svg>; ... }
}
```

Source: `%TEMP%/design-bundle-6474/go-tracker/project/components.jsx`, lines defining the `Icon` component.

- [ ] **Step 2: `web/src/components/Sidebar.tsx`**

Port the `<Sidebar>` from the prototype. Structure:

```tsx
import type { Route } from '../types/domain';
import { Icon } from './Icon';
import { ALERTS } from '../mock/data';

export interface SidebarProps {
  active: Route;
  onNavigate: (to: Route) => void;
}

export function Sidebar({ active, onNavigate }: SidebarProps) {
  // Brand block + grouped nav items per the bundle's structure.
  // Groups: "Cartera" (overview, portfolio, operations), "Mercado" (tickers, charts), "Sistema" (alerts, config).
  // The active item gets className="sb-item active".
  // Alert group shows a dot if there are unseen alerts in ALERTS.
}
```

Reference: `components.jsx` `<Sidebar>` block. Replicate exactly.

- [ ] **Step 3: `web/src/components/TopBar.tsx`**

```tsx
import type { Route, FxSnapshot } from '../types/domain';
import { ROUTE_TITLES } from '../lib/routing';
import { LATEST } from '../mock/data';
import { formatTicks } from '../lib/format';

export interface TopBarProps {
  route: Route;
  fx: FxSnapshot;
  lastTick: number;
}

export function TopBar({ route, fx, lastTick }: TopBarProps) {
  // tb-title with crumbs from ROUTE_TITLES
  // tb-spacer
  // tb-ticker: scrolling list of top symbols with current px
  // tb-search (kbd: ⌘K, no behavior yet — Phase 4c)
  // tb-status: LIVE indicator with formatted lastTick
}
```

Reference: `components.jsx` `<TopBar>` block.

- [ ] **Step 4: Don't commit yet — views still missing**

---

### Task 4a-7: Stub views (7 files)

**Files:**
- Create: `web/src/views/{Overview,Portfolio,Operations,Tickers,Charts,Alerts,Config}.tsx`

- [ ] **Step 1: Create each stub view**

Each view file follows this template:

```tsx
import type { AppContext } from '../App';

export interface ViewProps {
  ctx: AppContext;
}

export function Overview({ ctx: _ctx }: ViewProps) {
  return (
    <div className="page">
      <div className="page-head">
        <div>
          <h1 className="page-title">Overview</h1>
          <p className="page-sub">Phase 4b will replace this stub with real content.</p>
        </div>
      </div>
      <div className="card">
        <div className="card-head"><div className="card-title">Placeholder</div></div>
        <div className="card-body empty">
          <div className="title">Stubbed view</div>
          <div className="sub">Implemented in Phase 4b.</div>
        </div>
      </div>
    </div>
  );
}
```

Repeat for `Portfolio`, `Operations`, `Tickers`, `Charts`, `Alerts`, `Config` — each with its own page-title and page-sub matching the view's purpose.

- [ ] **Step 2: Build the frontend and confirm everything compiles**

```bash
cd web && npm run build 2>&1 | tail -10 && cd ..
ls -la web/dist/index.html
```

Expected: build succeeds, `dist/index.html` exists.

- [ ] **Step 3: Run tests one more time**

```bash
cd web && npm test -- --run && cd ..
```

Expected: formatters still green.

- [ ] **Step 4: Commit the shell + stub views**

```bash
git add web/src/App.tsx web/src/lib/routing.ts web/src/components web/src/views
git commit -m "$(cat <<'EOF'
feat(web): app shell, sidebar, topbar, and 7 stub views

State-based routing (no react-router — mirror the prototype).
Live-tick effect ports the prototype's setInterval (Phase 4c
replaces with SSE). All 7 view modules exist as stubs that the
shell can mount.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 4a-8: Go embed.FS wiring

**Files:**
- Create: `internal/web/embed.go`
- Modify: `cmd/agent/main.go`
- Delete: `internal/web/.gitkeep`

- [ ] **Step 1: `internal/web/embed.go`**

```go
// Package web embeds the built frontend assets and serves them via
// an http.Handler with SPA-style fallback to index.html.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// Handler returns an http.Handler that serves the embedded frontend.
// Unknown paths (no extension or no matching file) fall back to
// index.html so client-side routing continues to work.
func Handler() (http.Handler, error) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, err
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// API routes are mounted before this handler in main.go, so
		// anything reaching here is a frontend path.
		if hasFileExtension(r.URL.Path) {
			fileServer.ServeHTTP(w, r)
			return
		}
		// SPA fallback — serve index.html for client-side routes.
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/"
		fileServer.ServeHTTP(w, r2)
	}), nil
}

func hasFileExtension(p string) bool {
	idx := strings.LastIndex(p, ".")
	if idx == -1 {
		return false
	}
	return !strings.Contains(p[idx:], "/")
}
```

- [ ] **Step 2: Remove the placeholder**

```bash
rm internal/web/.gitkeep
```

- [ ] **Step 3: Update `cmd/agent/main.go` `newRouter` to mount the frontend**

Add the embed handler at `/` (after API routes), and add the import.

Edit `cmd/agent/main.go`:

In the imports block, add: `"github.com/joajo13/go-tracker/internal/web"`.

Replace `newRouter`:

```go
func newRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Handle("/healthz", api.HealthHandler())

	if h, err := web.Handler(); err == nil {
		r.Mount("/", h)
	} else {
		slog.Default().Error("failed to mount embedded frontend", "err", err)
	}

	return r
}
```

- [ ] **Step 4: Build the frontend, then build Go**

```bash
cd web && npm run build && cd ..
go build -o bin/agent ./cmd/agent
ls -la bin/agent
```

Expected: both succeed. The Go binary now embeds the frontend.

- [ ] **Step 5: Run tests for Go**

```bash
go test ./... 2>&1 | tail -10
```

Expected: all green. The existing health and main tests should still pass.

- [ ] **Step 6: Lint**

```bash
golangci-lint run ./... 2>&1; echo "EXIT=$?"
```

Expected: clean.

- [ ] **Step 7: Commit**

```bash
git add internal/web/ cmd/agent/main.go
git rm internal/web/.gitkeep 2>/dev/null || true
git commit -m "$(cat <<'EOF'
feat(web): embed frontend in Go binary with SPA fallback

internal/web/embed.go uses go:embed all:dist to bundle the Vite
build output. The handler does naive extension-based dispatch:
paths with an extension hit the file server, everything else falls
back to index.html so client-side routing keeps working.

The cmd/agent/main.go router now mounts the embed at /, after the
/healthz route.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 4a-9: Exit gate

- [ ] **Step 1: Full local CI gate**

```bash
echo "=== frontend tests ===" && (cd web && npm test -- --run)
echo "=== frontend build ===" && (cd web && npm run build) && ls -la web/dist/index.html
echo "=== backend lint ===" && golangci-lint run ./...
echo "=== backend tests ===" && go test ./...
echo "=== build agent ===" && go build -o bin/agent ./cmd/agent && ls -la bin/agent
```

All five sections must succeed.

- [ ] **Step 2: Smoke run end-to-end**

```bash
bin/agent &
AGENT_PID=$!
sleep 1
echo "--- /healthz ---"
curl -s http://localhost:8080/healthz
echo
echo "--- /  (HTML root) ---"
curl -s -o /tmp/index.html -w "status=%{http_code} content_type=%{content_type}\n" http://localhost:8080/
head -3 /tmp/index.html
echo "--- /assets/... (a built asset) ---"
ASSET=$(ls web/dist/assets | head -1)
curl -s -o /dev/null -w "asset=%{http_code}\n" "http://localhost:8080/assets/$ASSET"
echo "--- /portfolio (SPA fallback) ---"
curl -s -o /tmp/spa.html -w "status=%{http_code} content_type=%{content_type}\n" http://localhost:8080/portfolio
diff <(head -3 /tmp/index.html) <(head -3 /tmp/spa.html) && echo "(SPA fallback returns the same HTML)"
kill $AGENT_PID 2>/dev/null
wait $AGENT_PID 2>/dev/null
```

Expected: `/healthz` returns `{"status":"ok"}`, `/` returns 200 + HTML starting with `<!doctype html>`, an asset URL returns 200, and `/portfolio` returns the same `index.html` (SPA fallback).

- [ ] **Step 3: Update README with frontend workflow**

Add a "Frontend" section to the project root `README.md` after the existing quickstart:

```markdown
## Frontend (Phase 4)

The dashboard lives in `web/` (Vite + React + TypeScript) and is
embedded in the Go binary via `embed.FS`.

```bash
make web-dev          # Vite dev server with HMR (during view work)
make web-test         # frontend tests one-shot
make web-build        # production build (writes web/dist/)
make build            # embeds the latest web/dist/ into bin/agent
```

For production: run `make web-build` before `make build` (or before
`go build`). The repo intentionally does not check in `web/dist/`.
```

- [ ] **Step 4: Commit the README update**

```bash
git add README.md
git commit -m "$(cat <<'EOF'
docs: add frontend dev workflow to README

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

- [ ] **Step 5: STOP — Phase 4a checkpoint**

Do not start Phase 4b automatically. Report exit-gate results to the user and ask whether to continue inline, dispatch a subagent per view, or pause for another session.

---

## Phase 4b — MVP views (outline)

Each task below is one view, implemented red-green-refactor for any logic, and one final commit. Detailed steps are written when 4b starts (mirror the level of detail used in Phase 4a, with code blocks for each non-trivial component). Source for each view: `%TEMP%/design-bundle-6474/go-tracker/project/views-portfolio.jsx` (Overview, Portfolio, Operations) and `views-system.jsx` (Alerts).

- **4b-1: Overview** — KPI cards (multi-currency switcher), 30d portfolio chart (line + area), asset-type allocation, top winners/losers, source health summary, recent log entries. Logic to TDD: `topMovers(positions, n)` selector, `allocByType(positions)` selector.
- **4b-2: Portfolio** — dense table (sort, type filter, sparklines inline, weight bars, currency pill). Logic: `sortPositions(positions, key, dir)`.
- **4b-3: Operations** — historical table with tabs by origin, "Nueva operación" modal with form validation. Logic: `validateOperation(form)`.
- **4b-4: Alerts** — triggered alerts list + rule CRUD modal with expression preview. Logic: `previewExpression(rule)` — render the AST back to a human string.

Exit gate: `make ci && make web-test && make web-build && make build` all green; smoke-test each view in the browser.

---

## Phase 4c — Post-MVP views (outline)

- **4c-1: Tickers** — CRUD list (toggle active, ratio for CEDEARs, sources as chips). Source: `views-system.jsx`.
- **4c-2: Charts** — candlestick + volume + RSI/MA/Bollinger; spread CEDEAR vs underlying overlay; "Mi posición" overlay. Port `charts.jsx` (SVG primitives). Logic to TDD: `bucketize(ticks, resolution)`, `rsi(closes, window)`, `bollinger(closes, window, k)`. Coverage ≥90% on the indicator helpers (matches the SRS gate for `indicators/`).
- **4c-3: Config** — sidebar with 7 sub-sections, runtime stats card, danger-zone actions. Source: `views-system.jsx`.

Phase 4c does NOT yet replace mocks with real API. That migration happens once backend Phase 1-3 land and expose the real endpoints.

---

## Self-review checklist (run after Phase 4a tasks)

- [ ] **Spec coverage:**
  - SRS §3.7 (Dashboard) — shell + nav + all 7 view names → 4a-5, 4a-6, 4a-7
  - SRS §5.2.8 (Frontend) — Vite + TS + React → 4a-1; embed.FS → 4a-8
  - SRS NFR-S-01 (HTTPS) — out of scope here (Caddy/Nginx at deploy)
  - SRS NFR-S-04 (bcrypt) — auth deferred to a later phase
  - SRS Phase 4 MVP scope (Overview, Cartera, Operaciones, Alertas) → Phase 4b
- [ ] **Stack alignment:** Tailwind dropped per user decision in brainstorming; record in spec follow-up if needed.
- [ ] **No placeholders:** every step has runnable commands or full code. The "Reference: the bundle" notes point to specific source paths so the engineer can copy verbatim without guessing.
- [ ] **Type consistency:** `Route` / `Currency` / `Ticker` / `Position` / `AlertEvent` spelled identically across types/, mock/, components/, and views/.
- [ ] **TDD discipline:** 4a-4 is true red-green-refactor on real logic. 4b/4c each call out specific helpers to TDD (selectors, validators, indicator math).
- [ ] **Frontend tests do not block backend CI:** the GitHub Actions workflow in `.github/workflows/ci.yml` does not yet run the frontend test suite. Update it as part of 4a-9 if there's appetite; otherwise add it to the 4b plan.

---

## Execution handoff

Plan saved to `docs/superpowers/plans/2026-05-13-portfolio-agent-phase-4-dashboard.md`.

Default: **Inline execution** of Phase 4a in this session. After 4a-9 exit gate passes, present results to the user and ask whether to continue inline with 4b, dispatch parallel subagents per view, or pause for another session.
