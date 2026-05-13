package domain

import (
	"context"
	"errors"
)

// ErrNotFound is returned by repositories when a row does not exist.
var ErrNotFound = errors.New("not found")

//go:generate go run go.uber.org/mock/mockgen -source=ports.go -destination=../persistence/mocks/mock_repos.go -package=mocks

// TickerRepo persists Tickers.
type TickerRepo interface {
	Insert(ctx context.Context, t Ticker) (TickerID, error)
	Get(ctx context.Context, id TickerID) (Ticker, error)
	ListActive(ctx context.Context) ([]Ticker, error)
}

// PriceRepo persists price observations into the high-resolution table.
type PriceRepo interface {
	Insert(ctx context.Context, p Price) error
	LatestByTicker(ctx context.Context, id TickerID) (Price, error)
}
