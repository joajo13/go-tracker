/**
 * Domain types — mirror the Go domain (internal/domain) informally until
 * codegen lands. The shape follows the mock data so the views render
 * unchanged once the real API arrives in Phase 1-3 (the API layer will
 * produce these same shapes).
 */

export type TickerType = 'cedear' | 'us_stock' | 'crypto' | 'bond' | 'fx';

/** The currency the user wants to read totals in. */
export type DisplayCurrency = 'ARS' | 'USD_CCL' | 'USD_OFI' | 'USD_TARJ';

/** The currency a quote or operation was denominated in. */
export type NativeCurrency = 'USD' | 'ARS';

export type OperationType = 'BUY' | 'SELL';
export type OperationSource = 'manual' | 'broker_sync' | 'csv_import';

export type AlertStatus = 'new' | 'seen' | 'archived';
export type AlertSeverity = 'pos' | 'neg' | 'warn' | 'info';

export type SourceStatus = 'ok' | 'degraded' | 'manual' | 'down';

export type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';

export type AlertKind =
  | 'price_abs'
  | 'pct_change'
  | 'rsi'
  | 'ma_cross'
  | 'cedear_spread'
  | 'portfolio_pct'
  | 'volume_anom';

export type Route =
  | 'overview'
  | 'portfolio'
  | 'operations'
  | 'tickers'
  | 'charts'
  | 'alerts'
  | 'config';

export interface FxSnapshot {
  ccl: number;
  mep: number;
  oficial: number;
  tarjeta: number;
}

export interface Ticker {
  id: number;
  symbol: string;
  name: string;
  type: TickerType;
  underlying?: string;
  ratio?: number;
  sources: string[];
  poll: number;        // poll interval in seconds
  active: boolean;
  byma?: number;       // CEDEARs only — local ARS price
}

export interface Price {
  px: number;
  ccy: NativeCurrency;
  d1: number;          // absolute 1-day change in `ccy`
  dpct: number;        // percent 1-day change
}

export interface Operation {
  id: number;
  ticker: string;
  type: OperationType;
  ts: string;          // ISO-like local time (no Z) — operation timestamp
  qty: number;
  px: number;          // unit price in `ccy`
  ccy: NativeCurrency;
  comm: number;
  fees: number;
  broker: string;
  source: OperationSource;
}

export interface Position {
  ticker: string;
  type: TickerType;
  name: string;
  ratio?: number;
  qty: number;
  avgCostARS: number;
  avgCostUSD: number;
  mvARS: number;
  mvUSD: number;
  plARS: number;
  plUSD: number;
  plPctARS: number;
  plPctUSD: number;
  d1Pct: number;
  d1Abs: number;
  lastPx: number;
  lastCcy: NativeCurrency;
  byma?: number;
}

export interface AlertRule {
  id: number;
  name: string;
  ticker: string | null;   // null = portfolio-global rule
  kind: AlertKind;
  op: string;              // '>', '<', 'up', 'down', etc.
  value?: number;
  currency?: NativeCurrency;
  window?: string;         // '1d', '14d', '50d', ...
  cooldown: number;        // seconds
  active: boolean;
  hits: number;
}

export interface AlertEvent {
  id: number;
  ruleId: number;
  name: string;
  ticker: string;
  ts: string;
  obs: string;
  status: AlertStatus;
  severity: AlertSeverity;
}

export interface SourceHealth {
  name: string;
  host: string;
  status: SourceStatus;
  latency: number | null;
  calls: number;
  errors: number;
  rate: string;
}

export interface LogEntry {
  ts: string;          // 'HH:MM:SS'
  lvl: LogLevel;
  src: string;
  msg: string;
}

export interface Mover {
  symbol: string;
  dpct: number;
  px: number;
  type: TickerType;
}

export interface PortfolioPoint {
  t: number;           // unix ms
  v: number;           // portfolio value (ARS)
}

export interface Candle {
  t: number;
  o: number;
  h: number;
  l: number;
  c: number;
  v: number;
}
