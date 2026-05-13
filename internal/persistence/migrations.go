// Package persistence exposes the embedded migrations FS used by every backend.
package persistence

import "embed"

//go:embed migrations/*.sql
var Migrations embed.FS

// MigrationsDir is the path goose should treat as the migrations root inside
// the embedded filesystem.
const MigrationsDir = "migrations"
