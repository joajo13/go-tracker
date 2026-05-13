package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joajo13/go-tracker/internal/domain"
)

func TestNewOperation_ValidBuy(t *testing.T) {
	t.Parallel()

	ts := time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC)
	op, err := domain.NewOperation(&domain.OperationInput{
		TickerID:   1,
		Type:       domain.OperationBuy,
		Ts:         ts,
		Quantity:   "100",
		UnitPrice:  "172.45",
		Currency:   domain.CurrencyUSD,
		Commission: "1.50",
		MarketFees: "0.10",
		Broker:     "bull",
		Source:     domain.OperationSourceManual,
	})

	require.NoError(t, err)
	assert.Equal(t, domain.OperationBuy, op.Type)
	assert.Equal(t, "100", op.Quantity.String())
	assert.Equal(t, "172.45", op.UnitPrice.String())
}

func TestNewOperation_RejectsBadType(t *testing.T) {
	t.Parallel()
	_, err := domain.NewOperation(&domain.OperationInput{
		TickerID: 1, Type: "HOLD", Ts: time.Now().UTC(),
		Quantity: "1", UnitPrice: "1", Currency: domain.CurrencyUSD,
		Source: domain.OperationSourceManual,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidOperation)
}

func TestNewOperation_RejectsNegativeQuantity(t *testing.T) {
	t.Parallel()
	_, err := domain.NewOperation(&domain.OperationInput{
		TickerID: 1, Type: domain.OperationBuy, Ts: time.Now().UTC(),
		Quantity: "-1", UnitPrice: "1", Currency: domain.CurrencyUSD,
		Source: domain.OperationSourceManual,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidOperation)
}

func TestNewOperation_RejectsBadSource(t *testing.T) {
	t.Parallel()
	_, err := domain.NewOperation(&domain.OperationInput{
		TickerID: 1, Type: domain.OperationBuy, Ts: time.Now().UTC(),
		Quantity: "1", UnitPrice: "1", Currency: domain.CurrencyUSD,
		Source: "ouija",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidOperation)
}
