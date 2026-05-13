import type { AppContext } from '../App';

export interface ViewProps {
  ctx: AppContext;
}

export function Config(_: ViewProps) {
  return (
    <div className="page">
      <div className="page-head">
        <div>
          <h1 className="page-title">Configuración</h1>
          <p className="page-sub">Phase 4c reemplaza este stub con secciones de General/Polling/Broker/Alertas/Retención/Auth/Sistema.</p>
        </div>
      </div>
      <div className="card">
        <div className="card-head"><div className="card-title">Placeholder</div></div>
        <div className="card-body empty">
          <div className="title">Stubbed view</div>
          <div className="sub">Implementado en Phase 4c (post-MVP).</div>
        </div>
      </div>
    </div>
  );
}
