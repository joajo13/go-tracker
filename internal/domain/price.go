package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// ErrInvalidPrice is the sentinel wrapped by NewPrice validation failures.
var ErrInvalidPrice = errors.New("invalid price")

// Price is one observation of a ticker price at a specific moment, from a
// specific source.
type Price struct {
	TickerID TickerID
	Source   string
	Price    decimal.Decimal
	Currency Currency
	Ts       time.Time

	// FX snapshot at observation time. Zero-valued when unknown.
	CCL     decimal.Decimal
	Oficial decimal.Decimal
	Tarjeta decimal.Decimal
	MEP     decimal.Decimal
}

// PriceInput is the unvalidated payload accepted by NewPrice.
type PriceInput struct {
	TickerID TickerID
	Source   string
	Price    string
	Currency Currency
	Ts       time.Time

	CCL     string
	Oficial string
	Tarjeta string
	MEP     string
}

// NewPrice validates and returns a Price. Timestamps are normalised to UTC.
func NewPrice(in *PriceInput) (Price, error) {
	if in.TickerID == 0 {
		return Price{}, fmt.Errorf("%w: ticker id is required", ErrInvalidPrice)
	}
	if strings.TrimSpace(in.Source) == "" {
		return Price{}, fmt.Errorf("%w: source is required", ErrInvalidPrice)
	}
	if !in.Currency.IsValid() {
		return Price{}, fmt.Errorf("%w: unknown currency %q", ErrInvalidPrice, in.Currency)
	}
	if in.Ts.IsZero() {
		return Price{}, fmt.Errorf("%w: timestamp is required", ErrInvalidPrice)
	}

	price, err := parseDecimal(in.Price)
	if err != nil {
		return Price{}, fmt.Errorf("%w: price: %w", ErrInvalidPrice, err)
	}

	ccl, err := parseOptionalDecimal(in.CCL)
	if err != nil {
		return Price{}, fmt.Errorf("%w: ccl: %w", ErrInvalidPrice, err)
	}
	oficial, err := parseOptionalDecimal(in.Oficial)
	if err != nil {
		return Price{}, fmt.Errorf("%w: oficial: %w", ErrInvalidPrice, err)
	}
	tarjeta, err := parseOptionalDecimal(in.Tarjeta)
	if err != nil {
		return Price{}, fmt.Errorf("%w: tarjeta: %w", ErrInvalidPrice, err)
	}
	mep, err := parseOptionalDecimal(in.MEP)
	if err != nil {
		return Price{}, fmt.Errorf("%w: mep: %w", ErrInvalidPrice, err)
	}

	return Price{
		TickerID: in.TickerID,
		Source:   in.Source,
		Price:    price,
		Currency: in.Currency,
		Ts:       in.Ts.UTC(),
		CCL:      ccl,
		Oficial:  oficial,
		Tarjeta:  tarjeta,
		MEP:      mep,
	}, nil
}

func parseDecimal(s string) (decimal.Decimal, error) {
	if strings.TrimSpace(s) == "" {
		return decimal.Decimal{}, errors.New("empty value")
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("decode %q: %w", s, err)
	}
	return d, nil
}

func parseOptionalDecimal(s string) (decimal.Decimal, error) {
	if strings.TrimSpace(s) == "" {
		return decimal.Decimal{}, nil
	}
	return parseDecimal(s)
}

// PriceEvent is what the worker pool publishes onto the broadcaster.
type PriceEvent struct {
	Price Price
}
