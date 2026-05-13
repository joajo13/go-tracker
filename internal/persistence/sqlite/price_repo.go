package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/joajo13/go-tracker/internal/domain"
)

// PriceRepo is the SQLite-backed implementation of domain.PriceRepo.
type PriceRepo struct {
	db *sql.DB
}

// NewPriceRepo constructs a PriceRepo.
func NewPriceRepo(db *sql.DB) *PriceRepo {
	return &PriceRepo{db: db}
}

// Insert writes a Price into prices_1m. Conflicts on (ticker_id, source, ts)
// are treated as no-ops so ingest can replay safely.
func (r *PriceRepo) Insert(ctx context.Context, p *domain.Price) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO prices_1m
			(ticker_id, source, price, currency, ccl, oficial, tarjeta, mep, ts)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(ticker_id, source, ts) DO NOTHING`,
		int64(p.TickerID), p.Source, p.Price.String(), string(p.Currency),
		nullableDecimal(p.CCL), nullableDecimal(p.Oficial),
		nullableDecimal(p.Tarjeta), nullableDecimal(p.MEP),
		p.Ts.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("price insert: %w", err)
	}
	return nil
}

// LatestByTicker returns the most recent Price for the given ticker.
func (r *PriceRepo) LatestByTicker(ctx context.Context, id domain.TickerID) (domain.Price, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT ticker_id, source, price, currency, ccl, oficial, tarjeta, mep, ts
		FROM prices_1m
		WHERE ticker_id = ?
		ORDER BY ts DESC LIMIT 1`, id)
	return scanPrice(row)
}

func scanPrice(s scanner) (domain.Price, error) {
	var (
		tickerID                          int64
		source, priceStr, currency, tsStr string
		ccl, oficial, tarjeta, mep        sql.NullString
	)
	err := s.Scan(&tickerID, &source, &priceStr, &currency, &ccl, &oficial, &tarjeta, &mep, &tsStr)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Price{}, fmt.Errorf("%w: price", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Price{}, fmt.Errorf("price scan: %w", err)
	}
	ts, err := parseTime(tsStr)
	if err != nil {
		return domain.Price{}, fmt.Errorf("price ts: %w", err)
	}
	return domain.NewPrice(&domain.PriceInput{
		TickerID: domain.TickerID(tickerID),
		Source:   source,
		Price:    priceStr,
		Currency: domain.Currency(currency),
		Ts:       ts,
		CCL:      ccl.String,
		Oficial:  oficial.String,
		Tarjeta:  tarjeta.String,
		MEP:      mep.String,
	})
}

func nullableDecimal(d decimal.Decimal) any {
	if d.IsZero() {
		return nil
	}
	return d.String()
}
