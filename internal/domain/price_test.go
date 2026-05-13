package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joajo13/go-tracker/internal/domain"
)

func TestNewPrice_ValidUSD(t *testing.T) {
	t.Parallel()

	ts := time.Date(2026, 5, 14, 13, 30, 0, 0, time.UTC)
	p, err := domain.NewPrice(&domain.PriceInput{
		TickerID: 7,
		Source:   "yahoo",
		Price:    "172.45",
		Currency: domain.CurrencyUSD,
		Ts:       ts,
	})

	require.NoError(t, err)
	assert.Equal(t, domain.TickerID(7), p.TickerID)
	assert.Equal(t, "yahoo", p.Source)
	assert.Equal(t, "172.45", p.Price.String())
	assert.Equal(t, domain.CurrencyUSD, p.Currency)
	assert.Equal(t, ts, p.Ts)
}

func TestNewPrice_RejectsEmptyPrice(t *testing.T) {
	t.Parallel()
	_, err := domain.NewPrice(&domain.PriceInput{
		TickerID: 1, Source: "yahoo", Price: "",
		Currency: domain.CurrencyUSD, Ts: time.Now().UTC(),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPrice)
}

func TestNewPrice_RejectsBadCurrency(t *testing.T) {
	t.Parallel()
	_, err := domain.NewPrice(&domain.PriceInput{
		TickerID: 1, Source: "yahoo", Price: "10",
		Currency: "EUR", Ts: time.Now().UTC(),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPrice)
}

func TestNewPrice_RejectsZeroTimestamp(t *testing.T) {
	t.Parallel()
	_, err := domain.NewPrice(&domain.PriceInput{
		TickerID: 1, Source: "yahoo", Price: "10",
		Currency: domain.CurrencyUSD, Ts: time.Time{},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPrice)
}

func TestNewPrice_RejectsZeroTickerID(t *testing.T) {
	t.Parallel()
	_, err := domain.NewPrice(&domain.PriceInput{
		TickerID: 0, Source: "yahoo", Price: "10",
		Currency: domain.CurrencyUSD, Ts: time.Now().UTC(),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPrice)
}

func TestNewPriceEvent_Wraps(t *testing.T) {
	t.Parallel()
	p, err := domain.NewPrice(&domain.PriceInput{
		TickerID: 1, Source: "yahoo", Price: "10",
		Currency: domain.CurrencyUSD,
		Ts:       time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	ev := domain.PriceEvent{Price: p}
	assert.Equal(t, domain.TickerID(1), ev.Price.TickerID)
}
