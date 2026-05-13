# Portfolio Agent — SRS y Specifications

> **Software Requirements Specification (SRS) y Technical Specs**
> Agente de monitoreo de cartera de inversiones (CEDEARs, USA, cripto, bonos, MEP/CCL)
> Autor: Juan Giupponi
> Fecha: 2026-05-12
> Versión: 1.0 (draft inicial)

---

## 1. Introducción

### 1.1 Propósito del documento

Este documento describe los requisitos funcionales, no funcionales y la arquitectura técnica del **Portfolio Agent**, un agente de monitoreo de inversiones personal escrito en Go. Cubre el alcance completo desde la captura de precios desde fuentes externas hasta la visualización en un dashboard web embebido.

### 1.2 Alcance del producto

Portfolio Agent es una herramienta de uso personal que corre 24/7 en un VPS, mantiene una base de datos SQLite actualizada con precios y posiciones de la cartera del usuario, y expone un dashboard web para consultar estado actual, P&L multi-moneda, alertas disparadas e histórico de precios con resoluciones múltiples.

El sistema **no** ejecuta operaciones de trading. Es estrictamente read-only sobre el mundo financiero.

### 1.3 Objetivos

- Tener en un solo lugar el estado de toda la cartera, sin depender de varios apps de brokers.
- P&L realista con FIFO + comisiones, en ARS, USD CCL, USD oficial y USD tarjeta.
- Detectar spreads entre CEDEAR local y subyacente USA para arbitraje informativo.
- Sistema de alertas configurables persistidas (no notificaciones push).
- Aprender Go a fondo construyendo algo útil y publicable en GitHub.

### 1.4 Definiciones y acrónimos

| Término | Significado |
|---|---|
| CEDEAR | Certificado de Depósito Argentino, representa una acción extranjera operable en ByMA |
| CCL | Contado con Liquidación, dólar implícito en operaciones bursátiles |
| MEP | Mercado Electrónico de Pagos, otro dólar bursátil |
| FIFO | First In First Out, método de cálculo de costo de venta |
| P&L | Profit and Loss, ganancia/pérdida |
| ByMA | Bolsas y Mercados Argentinos |
| Ratio | Cantidad de CEDEARs equivalentes a 1 acción USA (ej AAPL ratio 20:1) |
| TUI | Text User Interface |
| SSE | Server-Sent Events |
| TDD | Test Driven Development |

---

## 2. Visión general del sistema

### 2.1 Perspectiva del producto

Sistema monolítico modular, binario único en Go que contiene:
- Worker pool concurrente que pollea fuentes de datos externas a intervalos configurables.
- Motor de cálculo de P&L con soporte FIFO multi-moneda.
- Motor de evaluación de reglas de alertas (precio absoluto, variación %, indicadores técnicos, portfolio global).
- Capa de persistencia con SQLite y estrategia de downsampling para histórico.
- API REST + SSE para el dashboard.
- Frontend React + Tailwind embebido en el binario vía `embed.FS`.

### 2.2 Funciones del producto (alto nivel)

1. **Ingesta de precios** desde múltiples fuentes en paralelo y con intervalos diferenciados por tipo de activo.
2. **Gestión de posiciones** con eventos de compra/venta, integración con Bull Market Brokers y carga manual.
3. **Cálculo de P&L** FIFO con comisiones, en 4 monedas (ARS, USD CCL, USD oficial, USD tarjeta).
4. **Motor de alertas** configurable, persistido en SQLite, con condiciones componibles.
5. **Dashboard web** para consultar todo + CRUD completo de tickers, alertas y posiciones.
6. **Histórico con resoluciones múltiples** (1m última semana, 1h último mes, 1d siempre).

### 2.3 Características de los usuarios

Usuario único, el autor. Perfil técnico avanzado, no requiere UX para no-técnicos pero sí pulido visual porque el frontend va a ser parte del portfolio público.

### 2.4 Restricciones

- Lenguaje backend, Go (idiomático, sin frameworks pesados).
- DB, SQLite (un solo archivo, deploy trivial).
- Deploy, VPS Linux con systemd, expuesto detrás de Caddy o Nginx con HTTPS.
- Sin servicios cloud pagos (todo gratis o casi).
- Free tiers de APIs financieras, respetar rate limits.
- Sin trading automatizado, read-only.

### 2.5 Supuestos y dependencias

- El VPS tiene conectividad estable a internet.
- Las APIs gratuitas siguen disponibles (Yahoo Finance, Finnhub, dolarapi).
- Bull Market Brokers no implementa contramedidas anti-scraping agresivas durante el desarrollo del MVP.

---

## 3. Requisitos funcionales

### 3.1 Gestión de activos (RF-AS)

**RF-AS-01.** El sistema debe permitir registrar tickers de los siguientes tipos:
- CEDEARs argentinos
- Acciones USA directas
- Criptomonedas
- Bonos argentinos (incluye AL30, GD30, etc)
- Tipos de cambio (CCL, MEP, oficial, tarjeta)

**RF-AS-02.** Cada ticker debe almacenar, símbolo, nombre legible, tipo, fuente(s) de datos, intervalo de polling, ratio (solo CEDEARs), ticker del subyacente (solo CEDEARs).

**RF-AS-03.** El sistema debe permitir CRUD completo de tickers desde el dashboard web.

### 3.2 Ingesta de precios (RF-IN)

**RF-IN-01.** El sistema debe pollear precios con intervalos configurables por tipo de activo. Valores por defecto sugeridos:

| Tipo | Intervalo en horario de mercado | Intervalo fuera de mercado |
|---|---|---|
| Cripto | 1 min | 1 min |
| Acciones USA | 1 min | 1 hora |
| CEDEARs (subyacente) | 1 min | 1 hora |
| CEDEARs (ByMA local) | 1 min | sin polling |
| Bonos ARG | 5 min | sin polling |
| MEP/CCL/Oficial/Tarjeta | 5 min | 15 min |

**RF-IN-02.** Para CEDEARs, el sistema debe obtener tanto el precio local de ByMA como el precio calculado (subyacente USA × CCL / ratio) y persistir ambos con el spread entre ellos.

**RF-IN-03.** Cada precio persistido debe incluir, ticker, fuente, precio, moneda, timestamp, valor del CCL/oficial/tarjeta vigente en ese instante.

**RF-IN-04.** Los horarios de mercado deben respetar feriados argentinos y de USA (NYSE/NASDAQ). El sistema mantiene un calendario interno actualizable.

**RF-IN-05.** Si una fuente falla, el sistema debe registrar el error, reintentar con backoff exponencial, y continuar operando para los demás tickers sin caerse.

**RF-IN-06.** El sistema respeta rate limits de cada API, encolando requests si es necesario.

### 3.3 Gestión de posiciones (RF-PO)

**RF-PO-01.** Las posiciones se modelan como una secuencia de **operaciones** (eventos de compra/venta), no como balances.

**RF-PO-02.** Cada operación contiene, ticker, tipo (BUY/SELL), fecha, cantidad, precio unitario, moneda, comisión, derechos de mercado, broker, notas opcionales.

**RF-PO-03.** El sistema debe permitir carga de operaciones por tres vías:
1. **Integración con Bull Market Brokers**, scraping del portal autenticado (best-effort).
2. **Import CSV** con formato definido, exportable manualmente desde el broker.
3. **Carga manual** desde el dashboard web con formulario.

**RF-PO-04.** Las posiciones actuales (cantidad neta, costo promedio FIFO) se calculan a demanda sobre las operaciones, no se persisten como estado mutable.

**RF-PO-05.** El sistema debe permitir editar y eliminar operaciones cargadas manualmente. Operaciones importadas del broker quedan marcadas como tales y son editables solo con un flag explícito.

### 3.4 Cálculo de P&L (RF-PL)

**RF-PL-01.** El P&L se calcula con metodología **FIFO**, descontando comisiones y aranceles de las compras (suman al costo) y de las ventas (restan del producido).

**RF-PL-02.** El P&L se expresa en **cuatro monedas en paralelo**:
- ARS
- USD CCL
- USD oficial
- USD tarjeta

**RF-PL-03.** Para conversiones a USD, se usa el valor del dólar correspondiente en el timestamp de cada operación (no el actual). Esto preserva la fidelidad histórica del P&L.

**RF-PL-04.** El P&L se desglosa en:
- **Realizado**, ganancia/pérdida de operaciones cerradas.
- **No realizado**, ganancia/pérdida de posiciones abiertas valuadas al último precio disponible.
- **Total**, suma de ambos.

**RF-PL-05.** El sistema debe exponer P&L por ticker, por tipo de activo y por portfolio global.

### 3.5 Motor de alertas (RF-AL)

**RF-AL-01.** El sistema debe soportar los siguientes tipos de condiciones de alerta:

| Tipo | Ejemplo |
|---|---|
| Precio absoluto | "AMD > 150 USD" |
| Variación porcentual | "AMD cae 5% en el día" |
| Cruce de media móvil | "TSM cruza MA50 de arriba hacia abajo" |
| RSI | "AAPL RSI > 70" |
| Volumen anómalo | "AMD volumen > 2x promedio 20 días" |
| Spread CEDEAR vs subyacente | "Spread AAPL > 2%" |
| Portfolio global | "Cartera total cae 3% en el día" |

**RF-AL-02.** Las condiciones deben ser componibles con operadores `AND` y `OR`.

**RF-AL-03.** Las alertas disparadas se persisten en SQLite con, timestamp, regla que la disparó, valores observados, estado (nueva/vista/archivada).

**RF-AL-04.** El sistema no emite notificaciones push (ni Telegram ni mail). Las alertas se consultan exclusivamente desde el dashboard.

**RF-AL-05.** Las alertas pueden ser de disparo único o recurrente (con cooldown configurable, ej "no disparar de nuevo en X minutos").

**RF-AL-06.** CRUD completo de reglas de alertas desde el dashboard.

### 3.6 Histórico de precios (RF-HI)

**RF-HI-01.** El sistema mantiene precios en **tres resoluciones**:
- **1 minuto**, últimos 7 días
- **1 hora**, últimos 90 días
- **1 día**, sin límite

**RF-HI-02.** Un proceso de downsampling corre periódicamente (cada hora) agregando datos viejos a granularidad menor y eliminando los datos de alta resolución que ya fueron agregados.

**RF-HI-03.** Las agregaciones calculan OHLCV (Open, High, Low, Close, Volume si aplica) para cada bucket.

**RF-HI-04.** El histórico es consultable desde el dashboard con gráficos.

### 3.7 Dashboard web (RF-DA)

**RF-DA-01.** El dashboard es servido por el mismo binario Go, accesible vía navegador.

**RF-DA-02.** El dashboard debe incluir las siguientes vistas:
- **Overview**, P&L global multi-moneda, gráfico de evolución de cartera, top movers del día.
- **Cartera**, listado de posiciones abiertas con P&L individual.
- **Operaciones**, historial completo de compras/ventas, con CRUD para las manuales.
- **Tickers**, CRUD de tickers monitoreados, con configuración de polling.
- **Alertas**, vista de alertas disparadas + CRUD de reglas.
- **Gráficos**, vista detallada por ticker con histórico en múltiples resoluciones.
- **Config**, ajustes globales (intervalos default, claves de API, credenciales del broker, etc).

**RF-DA-03.** El dashboard actualiza datos en tiempo real (precios cambiando sin refresh) usando Server-Sent Events.

**RF-DA-04.** El dashboard debe estar protegido por autenticación simple (usuario + password configurable en env vars, sesión con cookie).

### 3.8 Integración con broker (RF-BR)

**RF-BR-01.** El sistema debe integrarse con Bull Market Brokers para importar el portfolio actual y el historial de operaciones.

**RF-BR-02.** Dado que Bull no tiene API pública, la integración se implementa como scraping del portal web o intercepción de los requests de su frontend. Esta integración es **best-effort**, si se rompe debe degradarse a modo manual sin afectar el resto del sistema.

**RF-BR-03.** Las credenciales del broker se almacenan cifradas en disco o se leen desde env vars en runtime.

**RF-BR-04.** Sincronización con el broker, manual a demanda + automática cada 12 hs.

---

## 4. Requisitos no funcionales

### 4.1 Performance

- **NFR-P-01.** Polling de hasta 50 tickers simultáneos sin saturación.
- **NFR-P-02.** Latencia de respuesta del dashboard < 200 ms para consultas básicas.
- **NFR-P-03.** Uso de RAM en VPS < 100 MB en operación estándar.
- **NFR-P-04.** Uso de CPU < 5% promedio en VPS de 1 vCPU.

### 4.2 Confiabilidad

- **NFR-R-01.** El sistema debe reiniciarse automáticamente si crashea (systemd unit con `Restart=always`).
- **NFR-R-02.** Datos persistidos no se pierden ante reinicios.
- **NFR-R-03.** Backup automático diario de SQLite a un directorio configurable.

### 4.3 Seguridad

- **NFR-S-01.** El dashboard expone HTTPS detrás de un reverse proxy (Caddy/Nginx).
- **NFR-S-02.** Las credenciales (broker, API keys, password del dashboard) nunca quedan committeadas en el repo.
- **NFR-S-03.** Secrets se leen desde variables de entorno o un archivo `.env` ignorado por git.
- **NFR-S-04.** El password del dashboard se hashea con bcrypt antes de comparar.

### 4.4 Observabilidad

- **NFR-O-01.** Logs estructurados con `log/slog` (stdlib), formato JSON en producción.
- **NFR-O-02.** Niveles de log configurables (DEBUG, INFO, WARN, ERROR).
- **NFR-O-03.** Endpoint `/healthz` para monitoreo externo (UptimeRobot, etc).
- **NFR-O-04.** Métricas básicas expuestas (cantidad de polls, errores, latencia de fuentes).

### 4.5 Mantenibilidad

- **NFR-M-01.** Código sigue convenciones de Go (gofmt, golangci-lint).
- **NFR-M-02.** Cobertura de tests > 70% en paquetes de lógica de negocio.
- **NFR-M-03.** Arquitectura modular con separación clara entre dominios (precios, posiciones, alertas, etc).
- **NFR-M-04.** README claro con instrucciones de setup y deploy.

### 4.6 Portabilidad

- **NFR-PO-01.** El binario compila para Linux amd64 y arm64 (cubre VPS y Raspberry).
- **NFR-PO-02.** Frontend embebido en el binario, deploy = copiar un archivo.

---

## 5. Arquitectura técnica

### 5.1 Diagrama de alto nivel

```
+--------------------------------------------------------------+
|                     Portfolio Agent (binario Go)             |
|                                                              |
|  +------------+   +-------------+   +-------------------+    |
|  | Scheduler  |-->| Worker Pool |-->| Source Adapters   |    |
|  | (cron-like)|   | (goroutines)|   | (Yahoo, dolarapi, |    |
|  +------------+   +-------------+   |  Finnhub, ByMA,   |    |
|                          |          |  Bull, etc)       |    |
|                          v          +-------------------+    |
|                  +---------------+                           |
|                  | Price Channel |                           |
|                  +---------------+                           |
|                          |                                   |
|              +-----------+-----------+                       |
|              v                       v                       |
|     +-----------------+    +-------------------+             |
|     | Persistence     |    | Alert Evaluator   |             |
|     | (SQLite)        |    | (rule engine)     |             |
|     +-----------------+    +-------------------+             |
|              |                       |                       |
|              +-----------+-----------+                       |
|                          v                                   |
|              +-------------------------+                     |
|              | HTTP API + SSE          |                     |
|              | (chi router)            |                     |
|              +-------------------------+                     |
|                          ^                                   |
|              +-------------------------+                     |
|              | embed.FS                |                     |
|              | (React + Tailwind dist) |                     |
|              +-------------------------+                     |
+--------------------------------------------------------------+
                           ^
                           |
                  Caddy/Nginx (HTTPS)
                           ^
                           |
                       Browser
```

### 5.2 Componentes principales

#### 5.2.1 Scheduler
Goroutine que dispara polls según los intervalos configurados por ticker. Usa `time.Ticker` por grupo de intervalo, no un ticker por activo.

#### 5.2.2 Worker pool
Pool de goroutines (configurable, default 10) que consumen jobs de una cola y ejecutan polls. Permite limitar concurrencia y respetar rate limits.

#### 5.2.3 Source adapters
Interface `PriceSource` con implementaciones por fuente. Cada adapter encapsula su propia lógica de auth, rate limiting y parsing.

```go
type PriceSource interface {
    Name() string
    Fetch(ctx context.Context, symbol string) (*Price, error)
    RateLimit() rate.Limiter
}
```

#### 5.2.4 Persistence layer
- Driver, `modernc.org/sqlite` (puro Go, sin CGO).
- Migrations con `golang-migrate` o `goose`.
- Repositorios por dominio (`PriceRepo`, `OperationRepo`, `AlertRepo`, etc).
- **Decimal everywhere**, todos los montos usan `shopspring/decimal` o se almacenan como integers en centavos.

#### 5.2.5 Alert evaluator
Motor de reglas que evalúa cada precio nuevo contra las reglas activas. Para indicadores técnicos (RSI, medias móviles, volumen), consulta el histórico via repositorio. Las alertas disparadas se persisten con cooldown para evitar spam.

#### 5.2.6 P&L calculator
Módulo puro (sin side-effects) que toma una lista de operaciones + precios actuales + dólares y devuelve el P&L estructurado. Por ser puro es altamente testeable.

#### 5.2.7 HTTP API
- Router, `go-chi/chi`.
- Auth middleware con cookie de sesión.
- Endpoints REST + endpoint SSE para streaming de precios al frontend.
- Validación con `go-playground/validator`.

#### 5.2.8 Frontend
- React + Vite + TypeScript + Tailwind CSS.
- Diseño generado con Claude Design / frontend-design plugin.
- Build embebido en el binario via `embed.FS`.
- Cliente SSE para updates en tiempo real.

### 5.3 Modelo de datos (esquema SQLite simplificado)

```sql
-- Tickers
CREATE TABLE tickers (
  id INTEGER PRIMARY KEY,
  symbol TEXT NOT NULL,
  name TEXT NOT NULL,
  type TEXT NOT NULL CHECK(type IN ('cedear','us_stock','crypto','bond','fx')),
  underlying_symbol TEXT,     -- solo CEDEARs
  ratio TEXT,                  -- solo CEDEARs, decimal
  poll_interval_seconds INTEGER NOT NULL,
  sources TEXT NOT NULL,       -- JSON array
  active INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Precios alta resolución
CREATE TABLE prices_1m (
  ticker_id INTEGER REFERENCES tickers(id),
  source TEXT NOT NULL,
  price TEXT NOT NULL,         -- decimal
  currency TEXT NOT NULL,
  ccl TEXT, oficial TEXT, tarjeta TEXT, mep TEXT,
  ts DATETIME NOT NULL,
  PRIMARY KEY (ticker_id, source, ts)
);

CREATE TABLE prices_1h (
  ticker_id INTEGER REFERENCES tickers(id),
  source TEXT NOT NULL,
  open TEXT, high TEXT, low TEXT, close TEXT,
  ccl_close TEXT, oficial_close TEXT, tarjeta_close TEXT,
  ts DATETIME NOT NULL,
  PRIMARY KEY (ticker_id, source, ts)
);

CREATE TABLE prices_1d (
  ticker_id INTEGER REFERENCES tickers(id),
  source TEXT NOT NULL,
  open TEXT, high TEXT, low TEXT, close TEXT,
  ccl_close TEXT, oficial_close TEXT, tarjeta_close TEXT,
  ts DATE NOT NULL,
  PRIMARY KEY (ticker_id, source, ts)
);

-- Operaciones
CREATE TABLE operations (
  id INTEGER PRIMARY KEY,
  ticker_id INTEGER REFERENCES tickers(id),
  type TEXT NOT NULL CHECK(type IN ('BUY','SELL')),
  ts DATETIME NOT NULL,
  quantity TEXT NOT NULL,
  unit_price TEXT NOT NULL,
  currency TEXT NOT NULL,
  commission TEXT NOT NULL DEFAULT '0',
  market_fees TEXT NOT NULL DEFAULT '0',
  broker TEXT,
  source TEXT NOT NULL CHECK(source IN ('manual','broker_sync','csv_import')),
  notes TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Reglas de alertas
CREATE TABLE alert_rules (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  ticker_id INTEGER REFERENCES tickers(id), -- NULL = portfolio global
  expression TEXT NOT NULL,    -- JSON con AST de la condición
  cooldown_seconds INTEGER DEFAULT 3600,
  active INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Alertas disparadas
CREATE TABLE alerts (
  id INTEGER PRIMARY KEY,
  rule_id INTEGER REFERENCES alert_rules(id),
  ts DATETIME NOT NULL,
  observed_values TEXT NOT NULL, -- JSON
  status TEXT NOT NULL CHECK(status IN ('new','seen','archived')) DEFAULT 'new'
);
```

### 5.4 Estructura de carpetas

```
portfolio-agent/
├── cmd/
│   └── agent/
│       └── main.go                  # entrypoint
├── internal/
│   ├── config/                      # carga de config + env vars
│   ├── domain/                      # entidades de dominio puras
│   │   ├── ticker.go
│   │   ├── operation.go
│   │   ├── price.go
│   │   └── alert.go
│   ├── sources/                     # adapters de fuentes externas
│   │   ├── yahoo.go
│   │   ├── finnhub.go
│   │   ├── dolarapi.go
│   │   ├── byma.go
│   │   └── bull.go
│   ├── persistence/                 # repositorios SQLite
│   ├── scheduler/
│   ├── workers/
│   ├── pnl/                         # cálculo de P&L (lógica pura)
│   ├── alerts/                      # motor de reglas
│   ├── indicators/                  # RSI, MA, etc
│   ├── downsampler/                 # agregación de histórico
│   ├── api/                         # handlers HTTP
│   │   ├── auth.go
│   │   ├── tickers.go
│   │   ├── operations.go
│   │   ├── alerts.go
│   │   └── sse.go
│   └── web/                         # embed.FS del frontend
├── web/                             # frontend React
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── migrations/
├── scripts/
├── .github/
│   └── workflows/
│       └── ci.yml
├── go.mod
├── Makefile
├── README.md
└── LICENSE
```

---

## 6. Stack tecnológico

### Backend
- **Lenguaje**, Go 1.22+
- **Router HTTP**, `go-chi/chi`
- **DB**, SQLite via `modernc.org/sqlite` (puro Go, sin CGO)
- **Migrations**, `pressly/goose`
- **Decimales**, `shopspring/decimal`
- **Validación**, `go-playground/validator`
- **Logs**, `log/slog` (stdlib)
- **Tests**, stdlib + `stretchr/testify`
- **Mocks**, `uber-go/mock`
- **Rate limiting**, `golang.org/x/time/rate`
- **Scraping**, `chromedp` o `colly` para Bull Market
- **Env vars**, `caarlos0/env`

### Frontend
- **Framework**, React 18 + TypeScript
- **Build**, Vite
- **CSS**, Tailwind CSS
- **Charts**, Recharts o Lightweight Charts (TradingView)
- **State**, React Query + Zustand (o solo React Query)
- **Diseño**, generado con Claude Design / frontend-design plugin

### Infraestructura
- **VPS**, Hetzner CPX11 o similar (~3 EUR/mes)
- **Reverse proxy**, Caddy (HTTPS automático con Let's Encrypt)
- **Process manager**, systemd
- **Monitoring externo**, UptimeRobot (gratuito)

### CI/CD
- **CI**, GitHub Actions
- **Linting**, `golangci-lint` con config estricta
- **Test coverage**, `go test -cover` + reporte en Codecov

---

## 7. Plan de trabajo (roadmap por fases)

Considerando un par de tardes/noches por semana (~10 hs/semana) y MVP en 1 mes.

### Fase 0, setup (semana 1, ~5 hs)
- Crear repo público en GitHub con README, LICENSE (MIT), `.gitignore`, `.editorconfig`.
- Configurar `golangci-lint` con preset estricto.
- GitHub Actions, lint + test en cada PR.
- Scaffolding inicial de carpetas según sección 5.4.
- Setup de Vite + React + Tailwind en `web/`.
- Hello world end-to-end, binario que sirve `index.html` embebido.

### Fase 1, ingesta y persistencia (semana 1-2, ~10 hs)
- Modelo de datos + migrations.
- Repositorios con tests (TDD).
- Adapters Yahoo Finance y dolarapi.
- Scheduler básico + worker pool.
- Persistencia de precios en `prices_1m`.

### Fase 2, P&L y operaciones (semana 2-3, ~10 hs)
- Domain `Operation` + repositorio.
- Endpoints CRUD de operaciones.
- Módulo `pnl` con FIFO multi-moneda (alta cobertura de tests, lógica crítica).
- Import CSV.

### Fase 3, alertas básicas (semana 3, ~5 hs)
- Domain `AlertRule` + repositorio.
- Motor de evaluación para alertas de precio absoluto y variación %.
- Persistencia de alertas disparadas con cooldown.

### Fase 4, dashboard MVP (semana 3-4, ~10 hs)
- Auth simple (cookie de sesión).
- Vistas Overview, Cartera, Operaciones, Alertas.
- SSE para precios en tiempo real.
- Diseño aplicado con Claude Design.

### Post-MVP (después del mes)
- Integración Bull Market Brokers (scraping).
- Indicadores técnicos (RSI, MA, volumen).
- Vista de gráficos detallada por ticker.
- Downsampler de histórico.
- Detección de spreads CEDEAR vs subyacente.
- CRUD de tickers y alertas desde dashboard.
- Backup automático.
- Deploy en Hetzner con Caddy.

---

## 8. Estrategia de testing

### 8.1 Pirámide
- **Unitarios (mayoría)**, lógica de dominio, P&L, indicadores, parsers de fuentes.
- **Integración (medio)**, repositorios contra SQLite real en memoria, motor de alertas end-to-end.
- **E2E (pocos)**, smoke tests de los endpoints HTTP principales.

### 8.2 TDD
- Todo módulo de lógica pura (P&L, indicadores, evaluador de reglas) se escribe con TDD estricto.
- Fuentes externas se testean con mocks de HTTP.

### 8.3 Coverage
- Mínimo 70% global.
- Mínimo 90% en `pnl/`, `alerts/`, `indicators/` (lógica crítica).
- Reportado en cada PR vía GitHub Actions.

---

## 9. Riesgos y mitigaciones

| Riesgo | Probabilidad | Impacto | Mitigación |
|---|---|---|---|
| Yahoo Finance corta acceso | Media | Alto | Soporte multi-source, fallback a Finnhub o Twelve Data |
| Bull Market cambia su web | Alta | Medio | Integración aislada, degradación elegante a manual |
| Rate limits free tier insuficientes | Media | Medio | Worker pool con limitación + intervalos configurables |
| Errores de precisión con floats | Baja | Alto | `shopspring/decimal` en todo el sistema desde el día 1 💀 |
| Crashes silenciosos en goroutines | Media | Alto | `recover` en cada worker, logs estructurados, healthcheck |
| VPS comprometido | Baja | Alto | SSH keys only, fail2ban, secrets en env vars no en disk |

---

## 10. Definición de "Done" para el MVP

El MVP se considera entregado cuando:

- [ ] Repo público en GitHub con README, LICENSE, CI pasando.
- [ ] El binario corre en el VPS como servicio systemd con reinicio automático.
- [ ] Dashboard accesible vía HTTPS con auth funcionando.
- [ ] Al menos 10 tickers monitoreados de los 4 tipos (CEDEARs, USA, cripto, bonos/FX).
- [ ] P&L FIFO multi-moneda calculado correctamente, validado con tests.
- [ ] Carga de operaciones manual + import CSV funcionando.
- [ ] Alertas básicas (precio absoluto y variación %) persistidas y visibles.
- [ ] Vistas Overview, Cartera, Operaciones y Alertas en el dashboard.
- [ ] Cobertura de tests > 70% global, > 90% en módulos críticos.
- [ ] Linting limpio.
- [ ] README con instrucciones de setup y deploy reproducibles.

---

## 11. Apéndices

### A. Decisiones de diseño (ADRs resumidos)

**ADR-001, SQLite sobre Postgres.**
Justificación, uso personal, un solo proceso escritor, deploy trivial. Si crece, migrar es directo.

**ADR-002, modernc.org/sqlite sobre mattn/go-sqlite3.**
Justificación, puro Go sin CGO, cross-compilation trivial, binarios más portables. Trade-off, levemente más lento, irrelevante para este uso.

**ADR-003, Sin notificaciones push.**
Justificación, pull-based reduce ruido, evita dependencias de Telegram/SMTP, simplifica el MVP.

**ADR-004, Frontend embebido con embed.FS.**
Justificación, deploy = un archivo. No hay que servir archivos estáticos por separado ni configurar CORS.

**ADR-005, Operaciones como eventos, no balance.**
Justificación, event sourcing limitado permite recalcular P&L con cualquier metodología y auditar correcciones.

### B. Referencias

- Go effective patterns, https://go.dev/doc/effective_go
- chi router, https://github.com/go-chi/chi
- shopspring/decimal, https://github.com/shopspring/decimal
- Claude Design, https://support.claude.com/en/articles/14604416-get-started-with-claude-design
- Frontend Design plugin, https://claude.com/plugins/frontend-design

---

**Fin del documento.**

*Versión 1.0 — sujeto a iteración durante el desarrollo. Cambios sustanciales requieren bump de versión y nota de cambio.*
