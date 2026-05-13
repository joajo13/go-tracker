package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joajo13/go-tracker/internal/config"
)

func TestLoad_AppliesDefaults(t *testing.T) {
	// NOT parallel — sets process env vars.
	t.Setenv("DB_PATH", "/tmp/x.db")

	c, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "/tmp/x.db", c.DBPath)
	assert.Equal(t, ":8080", c.HTTPAddr)
	assert.Equal(t, 10, c.WorkerPoolSize)
	assert.Equal(t, 2.0, c.Yahoo.RatePerSec)
	assert.Equal(t, 4, c.Yahoo.RateBurst)
}

func TestLoad_FailsWithoutRequired(t *testing.T) {
	t.Setenv("DB_PATH", "")

	_, err := config.Load()
	require.Error(t, err)
}

func TestLoad_RespectsOverrides(t *testing.T) {
	t.Setenv("DB_PATH", "/tmp/x.db")
	t.Setenv("HTTP_ADDR", ":9000")
	t.Setenv("WORKER_POOL_SIZE", "25")
	t.Setenv("YAHOO_RATE_PER_SEC", "5")

	c, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, ":9000", c.HTTPAddr)
	assert.Equal(t, 25, c.WorkerPoolSize)
	assert.Equal(t, 5.0, c.Yahoo.RatePerSec)
}
