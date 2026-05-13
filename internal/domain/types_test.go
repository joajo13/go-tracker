package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/joajo13/go-tracker/internal/domain"
)

func TestAssetType_IsValid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   domain.AssetType
		want bool
	}{
		{domain.AssetTypeCEDEAR, true},
		{domain.AssetTypeUSStock, true},
		{domain.AssetTypeCrypto, true},
		{domain.AssetTypeBond, true},
		{domain.AssetTypeFX, true},
		{"weird", false},
		{"", false},
	}

	for _, c := range cases {
		assert.Equal(t, c.want, c.in.IsValid(), "in=%q", c.in)
	}
}

func TestCurrency_IsValid(t *testing.T) {
	t.Parallel()

	assert.True(t, domain.CurrencyARS.IsValid())
	assert.True(t, domain.CurrencyUSD.IsValid())
	assert.False(t, domain.Currency("EUR").IsValid())
	assert.False(t, domain.Currency("").IsValid())
}

func TestOperationType_IsValid(t *testing.T) {
	t.Parallel()

	assert.True(t, domain.OperationBuy.IsValid())
	assert.True(t, domain.OperationSell.IsValid())
	assert.False(t, domain.OperationType("HOLD").IsValid())
}

func TestAlertStatus_IsValid(t *testing.T) {
	t.Parallel()

	assert.True(t, domain.AlertStatusNew.IsValid())
	assert.True(t, domain.AlertStatusSeen.IsValid())
	assert.True(t, domain.AlertStatusArchived.IsValid())
	assert.False(t, domain.AlertStatus("muted").IsValid())
}
