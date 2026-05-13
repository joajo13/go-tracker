import type { DisplayCurrency } from '../types/domain';

type AnyCurrency = DisplayCurrency | 'ARS';

const SYMBOLS: Record<AnyCurrency, string> = {
  ARS: 'AR$',
  USD_CCL: '$',
  USD_OFI: '$',
  USD_TARJ: '$',
};

export function formatMoney(amount: number, ccy: AnyCurrency): string {
  if (!Number.isFinite(amount)) return '—';
  const sym = SYMBOLS[ccy];
  const fractionDigits = ccy === 'ARS' ? 0 : 2;
  const absStr = Math.abs(amount).toLocaleString('en-US', {
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  });
  return `${amount < 0 ? '-' : ''}${sym} ${absStr}`;
}

export function formatPercent(pct: number): string {
  if (!Number.isFinite(pct)) return '—';
  if (pct === 0) return '0.00%';
  const sign = pct > 0 ? '+' : '-';
  return `${sign}${Math.abs(pct).toFixed(2)}%`;
}

export function formatTicks(tsMs: number): string {
  const d = new Date(tsMs);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}
