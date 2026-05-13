package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/joajo13/go-tracker/internal/domain"
)

// TickerRepo is the SQLite-backed implementation of domain.TickerRepo.
type TickerRepo struct {
	db *sql.DB
}

// NewTickerRepo constructs a TickerRepo around an already-open DB.
func NewTickerRepo(db *sql.DB) *TickerRepo {
	return &TickerRepo{db: db}
}

// Insert persists a Ticker and returns the generated ID.
func (r *TickerRepo) Insert(ctx context.Context, t *domain.Ticker) (domain.TickerID, error) {
	sources, err := json.Marshal(t.Sources)
	if err != nil {
		return 0, fmt.Errorf("ticker sources: %w", err)
	}
	active := 0
	if t.Active {
		active = 1
	}
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO tickers (
			symbol, name, type, underlying_symbol, ratio,
			poll_interval_seconds, sources, active, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.Symbol, t.Name, string(t.Type), nullableString(t.UnderlyingSymbol),
		nullableString(t.Ratio), int64(t.PollInterval/time.Second), string(sources),
		active, time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return 0, fmt.Errorf("ticker insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ticker insert last id: %w", err)
	}
	return domain.TickerID(id), nil
}

// Get returns the Ticker with the given ID or domain.ErrNotFound.
func (r *TickerRepo) Get(ctx context.Context, id domain.TickerID) (domain.Ticker, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, symbol, name, type, underlying_symbol, ratio,
		       poll_interval_seconds, sources, active, created_at
		FROM tickers WHERE id = ?`, id)
	return scanTicker(row)
}

// ListActive returns every active Ticker.
func (r *TickerRepo) ListActive(ctx context.Context) ([]domain.Ticker, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, symbol, name, type, underlying_symbol, ratio,
		       poll_interval_seconds, sources, active, created_at
		FROM tickers WHERE active = 1`)
	if err != nil {
		return nil, fmt.Errorf("ticker list: %w", err)
	}
	defer rows.Close()

	var out []domain.Ticker
	for rows.Next() {
		tk, err := scanTicker(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, tk)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ticker rows: %w", err)
	}
	return out, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTicker(s scanner) (domain.Ticker, error) {
	var (
		id           int64
		symbol, name string
		typ          string
		underlying   sql.NullString
		ratio        sql.NullString
		pollSec      int64
		sourcesJSON  string
		active       int64
		createdAt    string
	)
	err := s.Scan(&id, &symbol, &name, &typ, &underlying, &ratio, &pollSec, &sourcesJSON, &active, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Ticker{}, fmt.Errorf("%w: ticker", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Ticker{}, fmt.Errorf("ticker scan: %w", err)
	}
	var sources []string
	if unmarshalErr := json.Unmarshal([]byte(sourcesJSON), &sources); unmarshalErr != nil {
		return domain.Ticker{}, fmt.Errorf("ticker sources decode: %w", unmarshalErr)
	}
	ts, err := parseTime(createdAt)
	if err != nil {
		return domain.Ticker{}, fmt.Errorf("ticker created_at: %w", err)
	}
	return domain.Ticker{
		ID:               domain.TickerID(id),
		Symbol:           symbol,
		Name:             name,
		Type:             domain.AssetType(typ),
		UnderlyingSymbol: underlying.String,
		Ratio:            ratio.String,
		PollInterval:     time.Duration(pollSec) * time.Second,
		Sources:          sources,
		Active:           active == 1,
		CreatedAt:        ts,
	}, nil
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unparseable time %q", s)
}
