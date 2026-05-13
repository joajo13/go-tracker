-- +goose Up
CREATE TABLE prices_1m (
    ticker_id INTEGER NOT NULL REFERENCES tickers(id) ON DELETE CASCADE,
    source TEXT NOT NULL,
    price TEXT NOT NULL,
    currency TEXT NOT NULL,
    ccl TEXT,
    oficial TEXT,
    tarjeta TEXT,
    mep TEXT,
    ts DATETIME NOT NULL,
    PRIMARY KEY (ticker_id, source, ts)
);

CREATE INDEX idx_prices_1m_ts ON prices_1m(ts);

-- +goose Down
DROP INDEX IF EXISTS idx_prices_1m_ts;
DROP TABLE IF EXISTS prices_1m;
