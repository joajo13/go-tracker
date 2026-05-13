// Package sources defines the contract every external price source must
// implement, plus the job envelope the scheduler hands to workers.
package sources

import (
	"context"

	"github.com/joajo13/go-tracker/internal/domain"
)

// PriceSource is the contract every external price adapter must satisfy.
type PriceSource interface {
	// Name uniquely identifies the source (e.g. "yahoo", "dolarapi"). It is
	// persisted on every PriceEvent and must match the sources stored on the
	// Ticker.
	Name() string
	// Fetch returns the latest Price for the given symbol. Adapters own
	// rate-limiting and backoff internally.
	Fetch(ctx context.Context, symbol string) (domain.Price, error)
}

// PollJob is what the scheduler enqueues for the worker pool.
type PollJob struct {
	TickerID domain.TickerID
	Symbol   string
	Source   string
}
