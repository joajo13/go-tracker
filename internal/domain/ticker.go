package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// TickerID is the persistent identifier of a Ticker. Zero means "not yet
// persisted".
type TickerID int64

// ErrInvalidTicker is the sentinel error wrapped by NewTicker validation
// failures.
var ErrInvalidTicker = errors.New("invalid ticker")

// Ticker is the immutable, validated representation of a monitored asset.
type Ticker struct {
	ID               TickerID
	Symbol           string
	Name             string
	Type             AssetType
	UnderlyingSymbol string
	Ratio            string
	PollInterval     time.Duration
	Sources          []string
	Active           bool
	CreatedAt        time.Time
}

// TickerInput is the unvalidated payload accepted by NewTicker.
type TickerInput struct {
	Symbol           string
	Name             string
	Type             AssetType
	UnderlyingSymbol string
	Ratio            string
	PollInterval     time.Duration
	Sources          []string
}

// NewTicker validates input and returns a Ticker with Active=true and a zero
// CreatedAt (the repo assigns it on insert).
func NewTicker(in TickerInput) (Ticker, error) {
	if strings.TrimSpace(in.Symbol) == "" {
		return Ticker{}, fmt.Errorf("%w: symbol is required", ErrInvalidTicker)
	}
	if strings.TrimSpace(in.Name) == "" {
		return Ticker{}, fmt.Errorf("%w: name is required", ErrInvalidTicker)
	}
	if !in.Type.IsValid() {
		return Ticker{}, fmt.Errorf("%w: unknown type %q", ErrInvalidTicker, in.Type)
	}
	if in.PollInterval <= 0 {
		return Ticker{}, fmt.Errorf("%w: poll interval must be positive", ErrInvalidTicker)
	}
	if len(in.Sources) == 0 {
		return Ticker{}, fmt.Errorf("%w: at least one source is required", ErrInvalidTicker)
	}
	if in.Type == AssetTypeCEDEAR && strings.TrimSpace(in.Ratio) == "" {
		return Ticker{}, fmt.Errorf("%w: CEDEARs require a ratio", ErrInvalidTicker)
	}

	sources := make([]string, len(in.Sources))
	copy(sources, in.Sources)

	return Ticker{
		Symbol:           in.Symbol,
		Name:             in.Name,
		Type:             in.Type,
		UnderlyingSymbol: in.UnderlyingSymbol,
		Ratio:            in.Ratio,
		PollInterval:     in.PollInterval,
		Sources:          sources,
		Active:           true,
	}, nil
}
