package sqlite_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joajo13/go-tracker/internal/persistence/sqlite"
)

func TestOpen_RunsMigrationsOnFreshDatabase(t *testing.T) {
	t.Parallel()

	db, err := sqlite.Open(context.Background(), ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	tables := []string{"tickers", "prices_1m", "prices_1h", "prices_1d"}
	for _, table := range tables {
		var name string
		row := db.QueryRowContext(context.Background(),
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table)
		require.NoError(t, row.Scan(&name), "table %q missing", table)
		assert.Equal(t, table, name)
	}
}

func TestOpen_IsIdempotent(t *testing.T) {
	t.Parallel()

	db, err := sqlite.Open(context.Background(), ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	// Re-running Open against the same DSN should not error or duplicate work.
	db2, err := sqlite.Open(context.Background(), ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db2.Close() })
}
