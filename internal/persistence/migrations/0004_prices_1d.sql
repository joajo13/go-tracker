-- +goose Up
CREATE TABLE prices_1d (
    ticker_id INTEGER NOT NULL REFERENCES tickers(id) ON DELETE CASCADE,
    source TEXT NOT NULL,
    open TEXT,
    high TEXT,
    low TEXT,
    close TEXT,
    ccl_close TEXT,
    oficial_close TEXT,
    tarjeta_close TEXT,
    ts DATE NOT NULL,
    PRIMARY KEY (ticker_id, source, ts)
);

-- +goose Down
DROP TABLE IF EXISTS prices_1d;
