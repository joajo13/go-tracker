-- +goose Up
CREATE TABLE tickers (
    id INTEGER PRIMARY KEY,
    symbol TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('cedear','us_stock','crypto','bond','fx')),
    underlying_symbol TEXT,
    ratio TEXT,
    poll_interval_seconds INTEGER NOT NULL,
    sources TEXT NOT NULL,
    active INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_tickers_symbol_type ON tickers(symbol, type);
CREATE INDEX idx_tickers_active ON tickers(active);

-- +goose Down
DROP INDEX IF EXISTS idx_tickers_active;
DROP INDEX IF EXISTS idx_tickers_symbol_type;
DROP TABLE IF EXISTS tickers;
