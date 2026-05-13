package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// OperationID is the persistent identifier of an Operation.
type OperationID int64

// OperationSource discriminates how an Operation entered the system.
type OperationSource string

const (
	// OperationSourceManual indicates the operation was entered by hand.
	OperationSourceManual OperationSource = "manual"
	// OperationSourceBroker indicates the operation was synced from a broker.
	OperationSourceBroker OperationSource = "broker_sync"
	// OperationSourceCSVImport indicates the operation was imported via CSV.
	OperationSourceCSVImport OperationSource = "csv_import"
)

// IsValid reports whether the source is one of the supported values.
func (s OperationSource) IsValid() bool {
	switch s {
	case OperationSourceManual, OperationSourceBroker, OperationSourceCSVImport:
		return true
	default:
		return false
	}
}

// ErrInvalidOperation is the sentinel wrapped by NewOperation failures.
var ErrInvalidOperation = errors.New("invalid operation")

// Operation is a buy or sell event. Positions are derived from a stream of
// Operations (RF-PO-01).
type Operation struct {
	ID         OperationID
	TickerID   TickerID
	Type       OperationType
	Ts         time.Time
	Quantity   decimal.Decimal
	UnitPrice  decimal.Decimal
	Currency   Currency
	Commission decimal.Decimal
	MarketFees decimal.Decimal
	Broker     string
	Source     OperationSource
	Notes      string
	CreatedAt  time.Time
}

// OperationInput is the unvalidated payload accepted by NewOperation.
type OperationInput struct {
	TickerID   TickerID
	Type       OperationType
	Ts         time.Time
	Quantity   string
	UnitPrice  string
	Currency   Currency
	Commission string
	MarketFees string
	Broker     string
	Source     OperationSource
	Notes      string
}

// NewOperation validates and returns an Operation.
func NewOperation(in *OperationInput) (Operation, error) {
	if in.TickerID == 0 {
		return Operation{}, fmt.Errorf("%w: ticker id required", ErrInvalidOperation)
	}
	if !in.Type.IsValid() {
		return Operation{}, fmt.Errorf("%w: type %q", ErrInvalidOperation, in.Type)
	}
	if in.Ts.IsZero() {
		return Operation{}, fmt.Errorf("%w: timestamp required", ErrInvalidOperation)
	}
	if !in.Currency.IsValid() {
		return Operation{}, fmt.Errorf("%w: currency %q", ErrInvalidOperation, in.Currency)
	}
	if !in.Source.IsValid() {
		return Operation{}, fmt.Errorf("%w: source %q", ErrInvalidOperation, in.Source)
	}

	qty, err := parsePositiveDecimal(in.Quantity)
	if err != nil {
		return Operation{}, fmt.Errorf("%w: quantity: %w", ErrInvalidOperation, err)
	}
	price, err := parsePositiveDecimal(in.UnitPrice)
	if err != nil {
		return Operation{}, fmt.Errorf("%w: unit price: %w", ErrInvalidOperation, err)
	}
	commission, err := parseNonNegativeDecimalDefaultZero(in.Commission)
	if err != nil {
		return Operation{}, fmt.Errorf("%w: commission: %w", ErrInvalidOperation, err)
	}
	fees, err := parseNonNegativeDecimalDefaultZero(in.MarketFees)
	if err != nil {
		return Operation{}, fmt.Errorf("%w: market fees: %w", ErrInvalidOperation, err)
	}

	return Operation{
		TickerID:   in.TickerID,
		Type:       in.Type,
		Ts:         in.Ts.UTC(),
		Quantity:   qty,
		UnitPrice:  price,
		Currency:   in.Currency,
		Commission: commission,
		MarketFees: fees,
		Broker:     strings.TrimSpace(in.Broker),
		Source:     in.Source,
		Notes:      in.Notes,
	}, nil
}

func parsePositiveDecimal(s string) (decimal.Decimal, error) {
	d, err := parseDecimal(s)
	if err != nil {
		return decimal.Decimal{}, err
	}
	if !d.IsPositive() {
		return decimal.Decimal{}, fmt.Errorf("must be > 0 (got %s)", d.String())
	}
	return d, nil
}

func parseNonNegativeDecimalDefaultZero(s string) (decimal.Decimal, error) {
	if strings.TrimSpace(s) == "" {
		return decimal.Zero, nil
	}
	d, err := parseDecimal(s)
	if err != nil {
		return decimal.Decimal{}, err
	}
	if d.IsNegative() {
		return decimal.Decimal{}, fmt.Errorf("must be >= 0 (got %s)", d.String())
	}
	return d, nil
}
