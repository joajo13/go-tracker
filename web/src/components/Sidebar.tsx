import type { Route } from '../types/domain';
import { ALERTS } from '../mock/data';
import { Icon, type IconName } from './Icon';

export interface SidebarProps {
  active: Route;
  onNavigate: (to: Route) => void;
}

interface NavItem {
  id: Route;
  label: string;
  icon: IconName;
}

interface NavGroup {
  group: string;
  items: NavItem[];
}

const GROUPS: NavGroup[] = [
  {
    group: 'Cartera',
    items: [
      { id: 'overview',   label: 'Overview',    icon: 'grid' },
      { id: 'portfolio',  label: 'Cartera',     icon: 'pie' },
      { id: 'operations', label: 'Operaciones', icon: 'list' },
    ],
  },
  {
    group: 'Mercado',
    items: [
      { id: 'charts',  label: 'Gráficos', icon: 'candle' },
      { id: 'tickers', label: 'Tickers',  icon: 'activity' },
    ],
  },
  {
    group: 'Sistema',
    items: [
      { id: 'alerts', label: 'Alertas', icon: 'bell' },
      { id: 'config', label: 'Config',  icon: 'settings' },
    ],
  },
];

export function Sidebar({ active, onNavigate }: SidebarProps) {
  const newAlerts = ALERTS.filter((a) => a.status === 'new').length;

  return (
    <aside className="sidebar">
      <div className="sb-brand">
        <div className="logo" />
        <div className="col" style={{ lineHeight: 1.1 }}>
          <span>go-tracker</span>
          <span
            style={{
              fontSize: 9,
              color: 'var(--fg-4)',
              fontFamily: 'var(--font-mono)',
              letterSpacing: '0.06em',
            }}
          >
            v1.0.0 · go 1.22
          </span>
        </div>
      </div>
      {GROUPS.map((g) => (
        <div className="sb-group" key={g.group}>
          <div className="sb-label">{g.group}</div>
          <div className="col gap-2" style={{ gap: 1 }}>
            {g.items.map((it) => {
              const showAlertCount = it.id === 'alerts' && newAlerts > 0;
              return (
                <div
                  key={it.id}
                  className={'sb-item' + (active === it.id ? ' active' : '')}
                  onClick={() => onNavigate(it.id)}
                >
                  <Icon name={it.icon} size={14} />
                  <span>{it.label}</span>
                  {showAlertCount && <span className="count">{newAlerts}</span>}
                </div>
              );
            })}
          </div>
        </div>
      ))}
      <div className="sb-footer">
        <div className="row items-center gap-3">
          <span className="live-dot" />
          <span>SSE conectado · 1 cliente</span>
        </div>
        <div>uptime 18d 4h 22m</div>
        <div>db 14.2 MB · 318k filas</div>
      </div>
    </aside>
  );
}
