// Package config loads the agent's environment-driven configuration.
package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Config is the parsed runtime configuration.
type Config struct {
	DBPath         string `env:"DB_PATH,required"`
	HTTPAddr       string `env:"HTTP_ADDR" envDefault:":8080"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat      string `env:"LOG_FORMAT" envDefault:"json"`
	WorkerPoolSize int    `env:"WORKER_POOL_SIZE" envDefault:"10"`

	Yahoo    SourceConfig `envPrefix:"YAHOO_"`
	DolarAPI SourceConfig `envPrefix:"DOLARAPI_"`
}

// SourceConfig configures a single price-source adapter.
type SourceConfig struct {
	RatePerSec float64 `env:"RATE_PER_SEC" envDefault:"2"`
	RateBurst  int     `env:"RATE_BURST" envDefault:"4"`
}

// Load reads env vars and validates them.
func Load() (Config, error) {
	var c Config
	if err := env.Parse(&c); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}
	if strings.TrimSpace(c.DBPath) == "" {
		return Config{}, fmt.Errorf("config: DB_PATH is required")
	}
	return c, nil
}
