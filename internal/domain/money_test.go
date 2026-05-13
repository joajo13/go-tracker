package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joajo13/go-tracker/internal/domain"
)

func TestParseAmount_ValidDecimal(t *testing.T) {
	t.Parallel()

	m, err := domain.ParseAmount("1234.5678")

	require.NoError(t, err)
	assert.Equal(t, "1234.5678", m.String())
}

func TestParseAmount_PreservesPrecision(t *testing.T) {
	t.Parallel()

	m, err := domain.ParseAmount("0.00000001")

	require.NoError(t, err)
	assert.Equal(t, "0.00000001", m.String())
}

func TestParseAmount_NegativeValue(t *testing.T) {
	t.Parallel()

	m, err := domain.ParseAmount("-42.5")

	require.NoError(t, err)
	assert.Equal(t, "-42.5", m.String())
}

func TestParseAmount_InvalidString(t *testing.T) {
	t.Parallel()

	_, err := domain.ParseAmount("not-a-number")

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestParseAmount_EmptyString(t *testing.T) {
	t.Parallel()

	_, err := domain.ParseAmount("")

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestMoney_ZeroValueStringIsZero(t *testing.T) {
	t.Parallel()

	var m domain.Money
	assert.Equal(t, "0", m.String())
}
