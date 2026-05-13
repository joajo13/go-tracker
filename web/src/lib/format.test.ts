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
    const ts = new Date(2026, 4, 13, 12, 34, 56).getTime();
    expect(formatTicks(ts)).toBe('12:34:56');
  });
  it('zero-pads single-digit values', () => {
    const ts = new Date(2026, 0, 1, 1, 2, 3).getTime();
    expect(formatTicks(ts)).toBe('01:02:03');
  });
});
