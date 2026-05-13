import type { Route } from '../types/domain';

export const ROUTE_TITLES: Record<Route, string[]> = {
  overview:   ['Cartera', 'Overview'],
  portfolio:  ['Cartera', 'Posiciones'],
  operations: ['Cartera', 'Operaciones'],
  tickers:    ['Mercado', 'Tickers'],
  charts:     ['Mercado', 'Gráficos'],
  alerts:     ['Sistema', 'Alertas'],
  config:     ['Sistema', 'Configuración'],
};
