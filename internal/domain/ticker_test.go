package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joajo13/go-tracker/internal/domain"
)

func TestNewTicker_ValidCEDEAR(t *testing.T) {
	t.Parallel()

	tk, err := domain.NewTicker(&domain.TickerInput{
		Symbol:           "AAPL",
		Name:             "Apple CEDEAR",
		Type:             domain.AssetTypeCEDEAR,
		UnderlyingSymbol: "AAPL",
		Ratio:            "10",
		PollInterval:     time.Minute,
		Sources:          []string{"byma", "yahoo"},
	})

	require.NoError(t, err)
	assert.Equal(t, "AAPL", tk.Symbol)
	assert.Equal(t, domain.AssetTypeCEDEAR, tk.Type)
	assert.Equal(t, "10", tk.Ratio)
	assert.Equal(t, time.Minute, tk.PollInterval)
	assert.True(t, tk.Active)
}

func TestNewTicker_RejectsEmptySymbol(t *testing.T) {
	t.Parallel()

	_, err := domain.NewTicker(&domain.TickerInput{
		Symbol:       "",
		Name:         "x",
		Type:         domain.AssetTypeUSStock,
		PollInterval: time.Minute,
		Sources:      []string{"yahoo"},
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidTicker)
}

func TestNewTicker_RejectsUnknownType(t *testing.T) {
	t.Parallel()

	_, err := domain.NewTicker(&domain.TickerInput{
		Symbol:       "AAPL",
		Name:         "Apple",
		Type:         "weird",
		PollInterval: time.Minute,
		Sources:      []string{"yahoo"},
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidTicker)
}

func TestNewTicker_RejectsZeroPollInterval(t *testing.T) {
	t.Parallel()

	_, err := domain.NewTicker(&domain.TickerInput{
		Symbol:       "AAPL",
		Name:         "Apple",
		Type:         domain.AssetTypeUSStock,
		PollInterval: 0,
		Sources:      []string{"yahoo"},
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidTicker)
}

func TestNewTicker_RejectsEmptySources(t *testing.T) {
	t.Parallel()

	_, err := domain.NewTicker(&domain.TickerInput{
		Symbol:       "AAPL",
		Name:         "Apple",
		Type:         domain.AssetTypeUSStock,
		PollInterval: time.Minute,
		Sources:      []string{},
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidTicker)
}

func TestNewTicker_CEDEARWithoutRatioFails(t *testing.T) {
	t.Parallel()

	_, err := domain.NewTicker(&domain.TickerInput{
		Symbol:           "AAPL",
		Name:             "Apple CEDEAR",
		Type:             domain.AssetTypeCEDEAR,
		UnderlyingSymbol: "AAPL",
		Ratio:            "",
		PollInterval:     time.Minute,
		Sources:          []string{"byma"},
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidTicker)
}
