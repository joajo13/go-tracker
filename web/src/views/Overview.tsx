import type { AppContext } from '../App';

export interface ViewProps {
  ctx: AppContext;
}

export function Overview(_: ViewProps) {
  return (
    <div className="page">
      <div className="page-head">
        <div>
          <h1 className="page-title">Overview</h1>
          <p className="page-sub">Phase 4b reemplaza este stub con KPIs, chart de evolución y top movers.</p>
        </div>
      </div>
      <div className="card">
        <div className="card-head"><div className="card-title">Placeholder</div></div>
        <div className="card-body empty">
          <div className="title">Stubbed view</div>
          <div className="sub">Implementado en Phase 4b.</div>
        </div>
      </div>
    </div>
  );
}
