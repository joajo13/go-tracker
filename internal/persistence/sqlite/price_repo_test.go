package sqlite_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joajo13/go-tracker/internal/domain"
	"github.com/joajo13/go-tracker/internal/persistence/sqlite"
)

func seedTicker(t *testing.T, db *sql.DB) domain.TickerID {
	t.Helper()
	repo := sqlite.NewTickerRepo(db)
	tk, err := domain.NewTicker(&domain.TickerInput{
		Symbol: "AAPL", Name: "Apple", Type: domain.AssetTypeUSStock,
		PollInterval: time.Minute, Sources: []string{"yahoo"},
	})
	require.NoError(t, err)
	id, err := repo.Insert(context.Background(), &tk)
	require.NoError(t, err)
	return id
}

func TestPriceRepo_InsertAndLatest(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	tid := seedTicker(t, db)
	repo := sqlite.NewPriceRepo(db)

	now := time.Date(2026, 5, 14, 13, 30, 0, 0, time.UTC)
	p1, err := domain.NewPrice(&domain.PriceInput{
		TickerID: tid, Source: "yahoo", Price: "172.45",
		Currency: domain.CurrencyUSD, Ts: now,
	})
	require.NoError(t, err)
	require.NoError(t, repo.Insert(context.Background(), &p1))

	p2, err := domain.NewPrice(&domain.PriceInput{
		TickerID: tid, Source: "yahoo", Price: "173.10",
		Currency: domain.CurrencyUSD, Ts: now.Add(time.Minute),
	})
	require.NoError(t, err)
	require.NoError(t, repo.Insert(context.Background(), &p2))

	got, err := repo.LatestByTicker(context.Background(), tid)
	require.NoError(t, err)
	assert.True(t, got.Price.Equal(decimal.RequireFromString("173.10")))
	assert.Equal(t, now.Add(time.Minute), got.Ts)
}

func TestPriceRepo_InsertIdempotentOnPK(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	tid := seedTicker(t, db)
	repo := sqlite.NewPriceRepo(db)

	now := time.Date(2026, 5, 14, 13, 30, 0, 0, time.UTC)
	p, err := domain.NewPrice(&domain.PriceInput{
		TickerID: tid, Source: "yahoo", Price: "172.45",
		Currency: domain.CurrencyUSD, Ts: now,
	})
	require.NoError(t, err)
	require.NoError(t, repo.Insert(context.Background(), &p))

	// Re-inserting the same (ticker, source, ts) must NOT error — repo upserts
	// on conflict to keep ingest idempotent.
	require.NoError(t, repo.Insert(context.Background(), &p))

	got, err := repo.LatestByTicker(context.Background(), tid)
	require.NoError(t, err)
	assert.True(t, got.Price.Equal(decimal.RequireFromString("172.45")))
}

func TestPriceRepo_LatestMissingReturnsErrNotFound(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	repo := sqlite.NewPriceRepo(db)

	_, err := repo.LatestByTicker(context.Background(), 999)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
