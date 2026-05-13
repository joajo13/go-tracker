/**
 * Mock data — verbatim port of the design bundle's mock-data.js.
 * Realistic Argentine portfolio across CEDEARs / USA / crypto / bonds / FX.
 *
 * The shapes match `types/domain.ts`. Phase 4c+ replaces these constants
 * with real API calls; the views consume the same types either way.
 */

import type {
  Ticker,
  Price,
  Operation,
  Position,
  AlertRule,
  AlertEvent,
  FxSnapshot,
  SourceHealth,
  LogEntry,
  Mover,
  PortfolioPoint,
  Candle,
} from '../types/domain';

export const NOW: number = new Date('2026-05-12T16:45:00-03:00').getTime();

export const FX: FxSnapshot = {
  ccl: 1284.5,
  mep: 1271.2,
  oficial: 942.0,
  tarjeta: 1601.4,
};

export const TICKERS: Ticker[] = [
  { id: 1,  symbol: 'AAPL',    name: 'Apple Inc.',           type: 'cedear',   underlying: 'AAPL',  ratio: 20, sources: ['byma', 'yahoo', 'dolarapi'], poll: 60,  active: true, byma: 6840 },
  { id: 2,  symbol: 'GOOGL',   name: 'Alphabet Class A',     type: 'cedear',   underlying: 'GOOGL', ratio: 53, sources: ['byma', 'yahoo', 'dolarapi'], poll: 60,  active: true, byma: 4290 },
  { id: 3,  symbol: 'MSFT',    name: 'Microsoft Corp.',      type: 'cedear',   underlying: 'MSFT',  ratio: 7,  sources: ['byma', 'yahoo', 'dolarapi'], poll: 60,  active: true, byma: 78230 },
  { id: 4,  symbol: 'AMD',     name: 'Adv. Micro Devices',   type: 'cedear',   underlying: 'AMD',   ratio: 4,  sources: ['byma', 'yahoo', 'dolarapi'], poll: 60,  active: true, byma: 47600 },
  { id: 5,  symbol: 'TSM',     name: 'Taiwan Semiconductor', type: 'cedear',   underlying: 'TSM',   ratio: 5,  sources: ['byma', 'yahoo', 'dolarapi'], poll: 60,  active: true, byma: 56120 },
  { id: 6,  symbol: 'NVDA',    name: 'NVIDIA Corp.',         type: 'cedear',   underlying: 'NVDA',  ratio: 10, sources: ['byma', 'yahoo', 'dolarapi'], poll: 60,  active: true, byma: 158800 },
  { id: 7,  symbol: 'KO',      name: 'Coca-Cola Co.',        type: 'cedear',   underlying: 'KO',    ratio: 7,  sources: ['byma', 'yahoo', 'dolarapi'], poll: 60,  active: true, byma: 12880 },
  { id: 8,  symbol: 'SPY',     name: 'SPDR S&P 500 ETF',     type: 'cedear',   underlying: 'SPY',   ratio: 20, sources: ['byma', 'yahoo', 'dolarapi'], poll: 60,  active: true, byma: 39740 },
  { id: 9,  symbol: 'TSLA',    name: 'Tesla Inc.',           type: 'us_stock', sources: ['yahoo', 'finnhub'], poll: 60,  active: true },
  { id: 10, symbol: 'META',    name: 'Meta Platforms',       type: 'us_stock', sources: ['yahoo', 'finnhub'], poll: 60,  active: true },
  { id: 11, symbol: 'BTC',     name: 'Bitcoin',              type: 'crypto',   sources: ['coingecko'], poll: 60,  active: true },
  { id: 12, symbol: 'ETH',     name: 'Ethereum',             type: 'crypto',   sources: ['coingecko'], poll: 60,  active: true },
  { id: 13, symbol: 'SOL',     name: 'Solana',               type: 'crypto',   sources: ['coingecko'], poll: 60,  active: true },
  { id: 14, symbol: 'AL30',    name: 'Bonar 2030 USD',       type: 'bond',     sources: ['byma'], poll: 300, active: true },
  { id: 15, symbol: 'GD30',    name: 'Global 2030 USD',      type: 'bond',     sources: ['byma'], poll: 300, active: true },
  { id: 16, symbol: 'GD35',    name: 'Global 2035 USD',      type: 'bond',     sources: ['byma'], poll: 300, active: true },
  { id: 17, symbol: 'CCL',     name: 'Dólar CCL',            type: 'fx',       sources: ['dolarapi'], poll: 300, active: true },
  { id: 18, symbol: 'MEP',     name: 'Dólar MEP',            type: 'fx',       sources: ['dolarapi'], poll: 300, active: true },
  { id: 19, symbol: 'OFICIAL', name: 'Dólar Oficial',        type: 'fx',       sources: ['dolarapi'], poll: 900, active: true },
  { id: 20, symbol: 'TARJETA', name: 'Dólar Tarjeta',        type: 'fx',       sources: ['dolarapi'], poll: 900, active: true },
];

/**
 * Mutated by App.tsx's live-tick effect — that's why this is `let`-style
 * (a const-bound record whose entries are reassigned in place).
 */
export const LATEST: Record<string, Price> = {
  AAPL:    { px: 219.60,  ccy: 'USD', d1: +0.92,   dpct: +0.42 },
  GOOGL:   { px: 178.34,  ccy: 'USD', d1: -1.21,   dpct: -0.67 },
  MSFT:    { px: 432.10,  ccy: 'USD', d1: +2.85,   dpct: +0.66 },
  AMD:     { px: 168.45,  ccy: 'USD', d1: +6.20,   dpct: +3.83 },
  TSM:     { px: 198.20,  ccy: 'USD', d1: -2.10,   dpct: -1.05 },
  NVDA:    { px: 938.50,  ccy: 'USD', d1: +14.80,  dpct: +1.60 },
  KO:      { px: 72.84,   ccy: 'USD', d1: +0.14,   dpct: +0.19 },
  SPY:     { px: 562.30,  ccy: 'USD', d1: +0.82,   dpct: +0.15 },
  TSLA:    { px: 257.40,  ccy: 'USD', d1: -4.85,   dpct: -1.85 },
  META:    { px: 612.18,  ccy: 'USD', d1: +3.40,   dpct: +0.56 },
  BTC:     { px: 92840,   ccy: 'USD', d1: +1820,   dpct: +2.00 },
  ETH:     { px: 3220.5,  ccy: 'USD', d1: -42.6,   dpct: -1.30 },
  SOL:     { px: 178.90,  ccy: 'USD', d1: +5.40,   dpct: +3.11 },
  AL30:    { px: 64.20,   ccy: 'USD', d1: +0.35,   dpct: +0.55 },
  GD30:    { px: 66.80,   ccy: 'USD', d1: +0.42,   dpct: +0.63 },
  GD35:    { px: 58.15,   ccy: 'USD', d1: -0.18,   dpct: -0.31 },
  CCL:     { px: 1284.50, ccy: 'ARS', d1: +3.10,   dpct: +0.24 },
  MEP:     { px: 1271.20, ccy: 'ARS', d1: +2.45,   dpct: +0.19 },
  OFICIAL: { px: 942.00,  ccy: 'ARS', d1: +1.00,   dpct: +0.11 },
  TARJETA: { px: 1601.40, ccy: 'ARS', d1: +1.70,   dpct: +0.11 },
};

export const OPERATIONS: Operation[] = [
  { id: 1,  ticker: 'AAPL',  type: 'BUY',  ts: '2024-03-15T14:32:10', qty: 40,   px: 6240,   ccy: 'ARS', comm: 312,   fees: 78,    broker: 'Bull Market', source: 'broker_sync' },
  { id: 2,  ticker: 'AAPL',  type: 'BUY',  ts: '2024-09-02T15:10:00', qty: 60,   px: 6480,   ccy: 'ARS', comm: 486,   fees: 122,   broker: 'Bull Market', source: 'broker_sync' },
  { id: 3,  ticker: 'AAPL',  type: 'SELL', ts: '2025-01-22T13:45:00', qty: 30,   px: 7180,   ccy: 'ARS', comm: 269,   fees: 67,    broker: 'Bull Market', source: 'broker_sync' },
  { id: 4,  ticker: 'NVDA',  type: 'BUY',  ts: '2024-05-10T14:00:00', qty: 25,   px: 84200,  ccy: 'ARS', comm: 1052,  fees: 263,   broker: 'Bull Market', source: 'broker_sync' },
  { id: 5,  ticker: 'NVDA',  type: 'BUY',  ts: '2024-11-18T15:30:00', qty: 15,   px: 142000, ccy: 'ARS', comm: 1065,  fees: 266,   broker: 'Bull Market', source: 'manual' },
  { id: 6,  ticker: 'AMD',   type: 'BUY',  ts: '2024-08-12T14:22:00', qty: 50,   px: 32400,  ccy: 'ARS', comm: 810,   fees: 202,   broker: 'Bull Market', source: 'broker_sync' },
  { id: 7,  ticker: 'MSFT',  type: 'BUY',  ts: '2024-02-05T14:10:00', qty: 12,   px: 68200,  ccy: 'ARS', comm: 409,   fees: 102,   broker: 'Bull Market', source: 'broker_sync' },
  { id: 8,  ticker: 'GOOGL', type: 'BUY',  ts: '2024-06-20T14:45:00', qty: 80,   px: 3820,   ccy: 'ARS', comm: 305,   fees: 76,    broker: 'Bull Market', source: 'csv_import' },
  { id: 9,  ticker: 'SPY',   type: 'BUY',  ts: '2024-04-08T14:05:00', qty: 40,   px: 32400,  ccy: 'ARS', comm: 1296,  fees: 324,   broker: 'Bull Market', source: 'broker_sync' },
  { id: 10, ticker: 'KO',    type: 'BUY',  ts: '2024-10-30T14:00:00', qty: 100,  px: 11200,  ccy: 'ARS', comm: 1120,  fees: 280,   broker: 'Bull Market', source: 'broker_sync' },
  { id: 11, ticker: 'TSM',   type: 'BUY',  ts: '2024-07-22T14:30:00', qty: 25,   px: 48200,  ccy: 'ARS', comm: 602,   fees: 150,   broker: 'Bull Market', source: 'broker_sync' },
  { id: 12, ticker: 'TSLA',  type: 'BUY',  ts: '2024-12-04T17:35:00', qty: 8,    px: 280.40, ccy: 'USD', comm: 4.50,  fees: 1.10,  broker: 'IBKR',        source: 'manual' },
  { id: 13, ticker: 'META',  type: 'BUY',  ts: '2025-02-14T17:20:00', qty: 5,    px: 575.20, ccy: 'USD', comm: 4.50,  fees: 1.10,  broker: 'IBKR',        source: 'manual' },
  { id: 14, ticker: 'BTC',   type: 'BUY',  ts: '2024-01-08T22:14:00', qty: 0.12, px: 42180,  ccy: 'USD', comm: 8.42,  fees: 0,     broker: 'Binance',     source: 'manual' },
  { id: 15, ticker: 'BTC',   type: 'BUY',  ts: '2024-08-30T20:00:00', qty: 0.08, px: 60840,  ccy: 'USD', comm: 12.16, fees: 0,     broker: 'Binance',     source: 'manual' },
  { id: 16, ticker: 'ETH',   type: 'BUY',  ts: '2024-02-22T21:00:00', qty: 1.8,  px: 2940,   ccy: 'USD', comm: 10.58, fees: 0,     broker: 'Binance',     source: 'manual' },
  { id: 17, ticker: 'SOL',   type: 'BUY',  ts: '2024-09-12T20:30:00', qty: 18,   px: 138.20, ccy: 'USD', comm: 4.97,  fees: 0,     broker: 'Binance',     source: 'manual' },
  { id: 18, ticker: 'AL30',  type: 'BUY',  ts: '2024-05-22T14:00:00', qty: 500,  px: 56.20,  ccy: 'USD', comm: 28.10, fees: 7.02,  broker: 'Bull Market', source: 'broker_sync' },
  { id: 19, ticker: 'GD30',  type: 'BUY',  ts: '2024-09-18T14:20:00', qty: 400,  px: 60.40,  ccy: 'USD', comm: 24.16, fees: 6.04,  broker: 'Bull Market', source: 'broker_sync' },
  { id: 20, ticker: 'GD35',  type: 'BUY',  ts: '2025-01-08T14:15:00', qty: 300,  px: 54.80,  ccy: 'USD', comm: 16.44, fees: 4.11,  broker: 'Bull Market', source: 'broker_sync' },
];

export const RULES: AlertRule[] = [
  { id: 1, name: 'AMD breakout',           ticker: 'AMD',  kind: 'price_abs',     op: '>',    value: 170,    currency: 'USD', cooldown: 3600,  active: true,  hits: 2 },
  { id: 2, name: 'BTC ATH approach',       ticker: 'BTC',  kind: 'price_abs',     op: '>',    value: 95000,  currency: 'USD', cooldown: 1800,  active: true,  hits: 0 },
  { id: 3, name: 'NVDA daily dump',        ticker: 'NVDA', kind: 'pct_change',    op: '<',    value: -3,     window: '1d',                     cooldown: 7200,  active: true,  hits: 1 },
  { id: 4, name: 'TSLA RSI overbought',    ticker: 'TSLA', kind: 'rsi',           op: '>',    value: 70,     window: '14d',                    cooldown: 86400, active: true,  hits: 0 },
  { id: 5, name: 'AAPL crosses MA50',      ticker: 'AAPL', kind: 'ma_cross',      op: 'down',                window: '50d',                    cooldown: 86400, active: true,  hits: 0 },
  { id: 6, name: 'Spread AAPL CEDEAR',     ticker: 'AAPL', kind: 'cedear_spread', op: '>',    value: 2,                                          cooldown: 3600,  active: true,  hits: 4 },
  { id: 7, name: 'Portfolio daily drop',   ticker: null,   kind: 'portfolio_pct', op: '<',    value: -3,     window: '1d',                     cooldown: 86400, active: true,  hits: 0 },
  { id: 8, name: 'CCL > 1300',             ticker: 'CCL',  kind: 'price_abs',     op: '>',    value: 1300,   currency: 'ARS', cooldown: 1800,  active: false, hits: 12 },
];

export const ALERTS: AlertEvent[] = [
  { id: 1, ruleId: 1, name: 'AMD breakout',         ticker: 'AMD',  ts: '2026-05-12T15:42:00', obs: 'AMD = 170.20 USD',           status: 'new',      severity: 'pos'  },
  { id: 2, ruleId: 6, name: 'Spread AAPL CEDEAR',   ticker: 'AAPL', ts: '2026-05-12T14:18:00', obs: 'Spread = 2.31% (CEDEAR caro)', status: 'new',      severity: 'warn' },
  { id: 3, ruleId: 1, name: 'AMD breakout',         ticker: 'AMD',  ts: '2026-05-12T11:05:00', obs: 'AMD = 170.85 USD',           status: 'seen',     severity: 'pos'  },
  { id: 4, ruleId: 3, name: 'NVDA daily dump',      ticker: 'NVDA', ts: '2026-05-09T20:01:00', obs: 'NVDA -3.42% en el día',      status: 'seen',     severity: 'neg'  },
  { id: 5, ruleId: 6, name: 'Spread AAPL CEDEAR',   ticker: 'AAPL', ts: '2026-05-08T18:30:00', obs: 'Spread = 2.05%',             status: 'seen',     severity: 'warn' },
  { id: 6, ruleId: 6, name: 'Spread AAPL CEDEAR',   ticker: 'AAPL', ts: '2026-05-08T15:12:00', obs: 'Spread = 2.18%',             status: 'archived', severity: 'warn' },
  { id: 7, ruleId: 6, name: 'Spread AAPL CEDEAR',   ticker: 'AAPL', ts: '2026-05-07T16:20:00', obs: 'Spread = 2.42%',             status: 'archived', severity: 'warn' },
];

export const SOURCES: SourceHealth[] = [
  { name: 'Yahoo Finance',                 host: 'query1.finance.yahoo.com',    status: 'ok',       latency: 184,  calls: 12420, errors: 3,  rate: '120/min' },
  { name: 'Finnhub',                       host: 'finnhub.io',                  status: 'ok',       latency: 92,   calls: 4180,  errors: 1,  rate: '60/min' },
  { name: 'dolarapi.com',                  host: 'dolarapi.com',                status: 'ok',       latency: 56,   calls: 8830,  errors: 0,  rate: 'unlim.' },
  { name: 'ByMA (data.byma.com.ar)',       host: 'data.byma.com.ar',            status: 'degraded', latency: 612,  calls: 9120,  errors: 41, rate: '30/min' },
  { name: 'CoinGecko',                     host: 'api.coingecko.com',           status: 'ok',       latency: 138,  calls: 6740,  errors: 0,  rate: '50/min' },
  { name: 'Bull Market Brokers (scraper)', host: 'cuenta.bullmarketbrokers.com', status: 'manual',  latency: null, calls: 2,     errors: 0,  rate: 'manual' },
];

export const LOGS: LogEntry[] = [
  { ts: '16:45:02', lvl: 'INFO',  src: 'scheduler',   msg: 'poll AAPL via yahoo OK in 142ms' },
  { ts: '16:45:02', lvl: 'INFO',  src: 'scheduler',   msg: 'poll AMD via yahoo OK in 168ms' },
  { ts: '16:45:01', lvl: 'INFO',  src: 'sse',         msg: 'fanout price update aapl=219.60 to 1 client' },
  { ts: '16:45:00', lvl: 'WARN',  src: 'byma',        msg: 'rate limit warning (28/30) - slowing down' },
  { ts: '16:44:58', lvl: 'INFO',  src: 'alert',       msg: 'rule#1 "AMD breakout" matched, observed=170.20' },
  { ts: '16:44:54', lvl: 'INFO',  src: 'downsampler', msg: 'aggregated 312 rows into prices_1h' },
  { ts: '16:44:42', lvl: 'ERROR', src: 'byma',        msg: 'GET /chart/AAPL: 502 - retry in 2s (backoff)' },
  { ts: '16:44:40', lvl: 'INFO',  src: 'pnl',         msg: 'recomputed P&L for 13 positions in 4ms' },
  { ts: '16:44:32', lvl: 'INFO',  src: 'sse',         msg: 'client connected, subs=[AAPL,AMD,NVDA,BTC,*]' },
  { ts: '16:44:25', lvl: 'DEBUG', src: 'scheduler',   msg: 'next tick group=us_stock in 35s' },
];

/* ===== Seeded RNG and synthetic time-series helpers ===== */

function rng(seed: number): () => number {
  let s = seed | 0;
  return () => {
    s = (s * 16807) % 2147483647;
    return (s & 0x7fffffff) / 0x7fffffff;
  };
}

function series(sym: string, base: number, n: number, vol: number, drift: number): number[] {
  const seed = sym.split('').reduce((a, c) => a + c.charCodeAt(0), 7) + n;
  const r = rng(seed);
  const arr: number[] = [];
  let v = base * 0.92;
  for (let i = 0; i < n; i++) {
    const step = (r() - 0.5) * vol * base + drift * base;
    v += step;
    arr.push(v);
  }
  // Anchor end value to base (the "current" price)
  const last = arr[arr.length - 1];
  const adj = base / last;
  return arr.map((x) => x * adj);
}

/** Sparkline series per symbol (40 points). */
export const SPARKS: Record<string, number[]> = (() => {
  const out: Record<string, number[]> = {};
  for (const sym of Object.keys(LATEST)) {
    const sign = LATEST[sym]!.dpct >= 0 ? 1 : -1;
    out[sym] = series(sym, LATEST[sym]!.px, 40, 0.012, sign * 0.0008);
  }
  return out;
})();

function portfolioHistory(days: number): PortfolioPoint[] {
  const r = rng(404);
  const start = 21500;
  const end = 26840;
  const arr: PortfolioPoint[] = [];
  let v = start;
  for (let i = 0; i < days; i++) {
    const drift = (end - start) / days;
    const noise = (r() - 0.5) * 220;
    v += drift + noise;
    arr.push({ t: NOW - (days - 1 - i) * 86400000, v });
  }
  arr[arr.length - 1]!.v = end;
  return arr;
}

export const PORTFOLIO_90: PortfolioPoint[] = portfolioHistory(90);

function candles(sym: string, base: number, n: number, vol: number): Candle[] {
  const r = rng(sym.charCodeAt(0) * 31 + n);
  const arr: Candle[] = [];
  let close = base * 0.78;
  for (let i = 0; i < n; i++) {
    const o = close;
    const drift = (r() - 0.45) * vol * base * 0.6;
    const c = o + drift;
    const hi = Math.max(o, c) + r() * vol * base * 0.5;
    const lo = Math.min(o, c) - r() * vol * base * 0.5;
    const v = Math.round((0.7 + r() * 0.6) * 1.2e7);
    arr.push({ t: NOW - (n - 1 - i) * 86400000, o, h: hi, l: lo, c, v });
    close = c;
  }
  // Anchor close to base
  const final = arr[arr.length - 1]!;
  final.c = base;
  final.h = Math.max(final.h, base);
  final.l = Math.min(final.l, base);
  return arr;
}

export const AMD_CANDLES: Candle[] = candles('AMD', 168.45, 60, 0.025);

/* ===== Position aggregation (simplified FIFO + multi-currency cost basis) ===== */

interface Lot {
  qty: number;
  costARS: number;
  costUSD: number;
}

function aggregate(): Record<string, Lot> {
  const lots: Record<string, Lot> = {};
  const sorted = OPERATIONS.slice().sort(
    (a, b) => new Date(a.ts).getTime() - new Date(b.ts).getTime(),
  );
  for (const op of sorted) {
    const sym = op.ticker;
    if (!lots[sym]) lots[sym] = { qty: 0, costARS: 0, costUSD: 0 };
    const lot = lots[sym]!;
    const totalNative = op.px * op.qty + op.comm + op.fees;
    if (op.type === 'BUY') {
      lot.qty += op.qty;
      if (op.ccy === 'ARS') {
        lot.costARS += totalNative;
        lot.costUSD += totalNative / FX.ccl;
      } else {
        lot.costARS += totalNative * FX.ccl;
        lot.costUSD += totalNative;
      }
    } else {
      const ratioSold = op.qty / lot.qty;
      lot.costARS *= 1 - ratioSold;
      lot.costUSD *= 1 - ratioSold;
      lot.qty -= op.qty;
    }
  }
  return lots;
}

/** Build position rows with current market value and P&L in ARS and USD. */
export function positions(): Position[] {
  const lots = aggregate();
  const out: Position[] = [];
  for (const sym of Object.keys(lots)) {
    const p = lots[sym]!;
    if (p.qty <= 0.0000001) continue;
    const t = TICKERS.find((x) => x.symbol === sym);
    const latest = LATEST[sym];
    if (!t || !latest) continue;

    let mvUSD: number;
    let mvARS: number;
    if (t.type === 'cedear' && t.byma !== undefined) {
      mvARS = t.byma * p.qty;
      mvUSD = mvARS / FX.ccl;
    } else if (t.type === 'us_stock' || t.type === 'crypto' || t.type === 'bond') {
      mvUSD = latest.px * p.qty;
      mvARS = mvUSD * FX.ccl;
    } else {
      mvARS = latest.px * p.qty;
      mvUSD = mvARS / FX.ccl;
    }

    const avgCostARS = p.costARS / (p.qty || 1);
    const avgCostUSD = p.costUSD / (p.qty || 1);
    const plARS = mvARS - p.costARS;
    const plUSD = mvUSD - p.costUSD;
    const plPctARS = (plARS / p.costARS) * 100;
    const plPctUSD = (plUSD / p.costUSD) * 100;

    out.push({
      ticker: sym,
      type: t.type,
      name: t.name,
      ratio: t.ratio,
      qty: p.qty,
      avgCostARS,
      avgCostUSD,
      mvARS,
      mvUSD,
      plARS,
      plUSD,
      plPctARS,
      plPctUSD,
      d1Pct: latest.dpct,
      d1Abs: latest.d1,
      lastPx: latest.px,
      lastCcy: latest.ccy,
      byma: t.byma,
    });
  }
  return out.sort((a, b) => b.mvUSD - a.mvUSD);
}

export const MOVERS: Mover[] = Object.keys(LATEST)
  .map((s) => {
    const t = TICKERS.find((x) => x.symbol === s);
    return {
      symbol: s,
      dpct: LATEST[s]!.dpct,
      px: LATEST[s]!.px,
      type: t!.type,
    };
  })
  .filter((x) => x.type !== 'fx' && x.type !== 'bond');
