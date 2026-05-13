package sqlite_test

import (
	"context"
	"path/filepath"
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

	dsn := filepath.Join(t.TempDir(), "idem.db")

	db1, err := sqlite.Open(context.Background(), dsn)
	require.NoError(t, err)
	require.NoError(t, db1.Close())

	// Re-open the SAME database file. goose should detect that all
	// migrations have already been applied and return without error.
	db2, err := sqlite.Open(context.Background(), dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db2.Close() })

	// Verify the table is still there from the first migration run.
	var name string
	row := db2.QueryRowContext(context.Background(),
		`SELECT name FROM sqlite_master WHERE type='table' AND name='tickers'`)
	require.NoError(t, row.Scan(&name))
	assert.Equal(t, "tickers", name)
}
