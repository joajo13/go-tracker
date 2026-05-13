package sqlite_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joajo13/go-tracker/internal/domain"
	"github.com/joajo13/go-tracker/internal/persistence/sqlite"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sqlite.Open(context.Background(), ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestTickerRepo_InsertAndGet(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	repo := sqlite.NewTickerRepo(db)

	in, err := domain.NewTicker(&domain.TickerInput{
		Symbol: "AAPL", Name: "Apple", Type: domain.AssetTypeUSStock,
		PollInterval: time.Minute, Sources: []string{"yahoo"},
	})
	require.NoError(t, err)

	id, err := repo.Insert(context.Background(), &in)
	require.NoError(t, err)
	assert.NotZero(t, id)

	got, err := repo.Get(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, "AAPL", got.Symbol)
	assert.Equal(t, time.Minute, got.PollInterval)
	assert.Equal(t, []string{"yahoo"}, got.Sources)
	assert.True(t, got.Active)
	assert.False(t, got.CreatedAt.IsZero())
}

func TestTickerRepo_GetMissingReturnsErrNotFound(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	repo := sqlite.NewTickerRepo(db)

	_, err := repo.Get(context.Background(), 999)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestTickerRepo_ListActiveOnlyReturnsActive(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	repo := sqlite.NewTickerRepo(db)

	active, err := domain.NewTicker(&domain.TickerInput{
		Symbol: "AAPL", Name: "Apple", Type: domain.AssetTypeUSStock,
		PollInterval: time.Minute, Sources: []string{"yahoo"},
	})
	require.NoError(t, err)
	inactive, err := domain.NewTicker(&domain.TickerInput{
		Symbol: "MSFT", Name: "Microsoft", Type: domain.AssetTypeUSStock,
		PollInterval: time.Minute, Sources: []string{"yahoo"},
	})
	require.NoError(t, err)
	inactive.Active = false

	_, err = repo.Insert(context.Background(), &active)
	require.NoError(t, err)
	_, err = repo.Insert(context.Background(), &inactive)
	require.NoError(t, err)

	got, err := repo.ListActive(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "AAPL", got[0].Symbol)
}
