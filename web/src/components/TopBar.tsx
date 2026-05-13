import type { Route } from '../types/domain';
import { ROUTE_TITLES } from '../lib/routing';
import { formatPercent, formatTicks } from '../lib/format';
import { MOVERS } from '../mock/data';
import { Icon } from './Icon';

export interface TopBarProps {
  route: Route;
  lastTick: number;
}

export function TopBar({ route, lastTick }: TopBarProps) {
  const crumbs = ROUTE_TITLES[route];
  const movers = MOVERS.slice(0, 8);

  return (
    <header className="topbar">
      <div className="tb-title">
        {crumbs.map((c, i) => (
          <span
            key={`${c}-${i}`}
            className="row items-center gap-3"
            style={{ display: 'inline-flex' }}
          >
            <span className={i === crumbs.length - 1 ? '' : 'crumb'}>{c}</span>
            {i < crumbs.length - 1 && <span className="sep">/</span>}
          </span>
        ))}
      </div>
      <div className="vr" style={{ height: 20 }} />
      <div className="tb-ticker">
        {movers.map((m, i) => (
          <span className="item" key={`${m.symbol}-${i}`}>
            <span className="sym">{m.symbol}</span>
            <span className="px">
              {m.px < 10
                ? m.px.toFixed(2)
                : m.px.toLocaleString('en-US', { maximumFractionDigits: 2 })}
            </span>
            <span
              className={m.dpct >= 0 ? 'pos' : 'neg'}
              style={{ fontSize: 'var(--t-xs)' }}
            >
              {formatPercent(m.dpct)}
            </span>
          </span>
        ))}
      </div>
      <div className="tb-spacer" />
      <div className="tb-search">
        <Icon name="search" size={12} />
        <span>Buscar ticker, alerta…</span>
        <span className="kbd">⌘K</span>
      </div>
      <div className="tb-status">
        <span className="live-dot" />
        <span>LIVE · {formatTicks(lastTick)}</span>
      </div>
    </header>
  );
}
