import { useState, useEffect, useMemo, useCallback } from 'react';
import type { Route, DisplayCurrency, Position } from './types/domain';
import { LATEST, positions as computePositions } from './mock/data';
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
  ccy: DisplayCurrency;
  setCcy: (c: DisplayCurrency) => void;
  range: string;
  setRange: (r: string) => void;
  positions: Position[];
  nav: (to: Route, params?: Record<string, unknown>) => void;
  params: Record<string, unknown>;
  route: Route;
  toast: (msg: string) => void;
  lastTick: number;
}

const LIVE_TICK_INTERVAL_MS = 3500;
const LIVE_TICK_PERTURB_COUNT = 3;
const LIVE_TICK_PERTURB_VOL = 0.0025;

export function App() {
  const [route, setRoute] = useState<Route>('overview');
  const [params, setParams] = useState<Record<string, unknown>>({});
  const [ccy, setCcy] = useState<DisplayCurrency>('USD_CCL');
  const [range, setRange] = useState('1M');
  const [toastMsg, setToastMsg] = useState<string | null>(null);
  const [lastTick, setLastTick] = useState(Date.now());

  // Live tick simulation — Phase 4c replaces this with SSE from the backend.
  useEffect(() => {
    const id = setInterval(() => {
      const syms = Object.keys(LATEST);
      for (let i = 0; i < LIVE_TICK_PERTURB_COUNT; i++) {
        const s = syms[Math.floor(Math.random() * syms.length)]!;
        const l = LATEST[s]!;
        l.px += (Math.random() - 0.5) * LIVE_TICK_PERTURB_VOL * l.px;
      }
      setLastTick(Date.now());
    }, LIVE_TICK_INTERVAL_MS);
    return () => clearInterval(id);
  }, []);

  const positions = useMemo(() => computePositions(), [lastTick]);

  const nav = useCallback((to: Route, p: Record<string, unknown> = {}) => {
    setRoute(to);
    setParams(p);
    document.querySelector('.main')?.scrollTo(0, 0);
  }, []);

  const toast = useCallback((msg: string) => {
    setToastMsg(msg);
    setTimeout(() => setToastMsg(null), 2200);
  }, []);

  const ctx: AppContext = {
    ccy, setCcy, range, setRange, positions, nav, params, route, toast, lastTick,
  };

  return (
    <div className="app">
      <Sidebar active={route} onNavigate={nav} />
      <TopBar route={route} lastTick={lastTick} />
      <main className="main">
        {route === 'overview'   && <Overview   ctx={ctx} />}
        {route === 'portfolio'  && <Portfolio  ctx={ctx} />}
        {route === 'operations' && <Operations ctx={ctx} />}
        {route === 'tickers'    && <Tickers    ctx={ctx} />}
        {route === 'charts'     && <Charts     ctx={ctx} />}
        {route === 'alerts'     && <Alerts     ctx={ctx} />}
        {route === 'config'     && <Config     ctx={ctx} />}
      </main>
      {toastMsg && (
        <div className="toast">
          <span>{toastMsg}</span>
        </div>
      )}
    </div>
  );
}
