// Package domain holds the pure domain types of the portfolio agent.
// No type in this package performs IO.
package domain

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

// ErrInvalidAmount is returned when ParseAmount cannot decode a string.
var ErrInvalidAmount = errors.New("invalid amount")

// Money is a precise decimal monetary value. It wraps shopspring/decimal so
// callers cannot accidentally use float arithmetic on financial data.
type Money struct {
	amount decimal.Decimal
}

// ParseAmount parses a decimal string into a Money value.
// Empty strings and non-numeric inputs return ErrInvalidAmount.
func ParseAmount(s string) (Money, error) {
	if s == "" {
		return Money{}, fmt.Errorf("%w: empty string", ErrInvalidAmount)
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return Money{}, fmt.Errorf("%w: %q: %w", ErrInvalidAmount, s, err)
	}
	return Money{amount: d}, nil
}

// String returns the canonical decimal representation. The zero Money is "0".
func (m Money) String() string {
	return m.amount.String()
}
