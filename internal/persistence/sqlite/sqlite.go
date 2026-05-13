// Package sqlite is the SQLite-backed implementation of the persistence
// interfaces declared in internal/domain.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite" // registers the "sqlite" driver

	"github.com/joajo13/go-tracker/internal/persistence"
)

// Open opens a SQLite database at the given DSN, applies any pending goose
// migrations, and returns the connection. The caller owns Close.
//
// DSN examples:
//   - in-memory: ":memory:"
//   - file:     "/var/lib/portfolio/portfolio.db"
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite ping: %w", err)
	}
	if err := applyMigrations(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func applyMigrations(ctx context.Context, db *sql.DB) error {
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}
	sub, err := fs.Sub(persistence.Migrations, persistence.MigrationsDir)
	if err != nil {
		return fmt.Errorf("migrations sub-fs: %w", err)
	}
	goose.SetBaseFS(sub)
	if err := goose.UpContext(ctx, db, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
