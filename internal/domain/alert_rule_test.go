package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joajo13/go-tracker/internal/domain"
)

func TestNewAlertRule_ValidPriceAbove(t *testing.T) {
	t.Parallel()

	expr, _ := json.Marshal(map[string]any{
		"op": "price_above", "ticker": "AAPL", "value": "150",
	})
	rule, err := domain.NewAlertRule(&domain.AlertRuleInput{
		Name:       "AAPL above 150",
		TickerID:   7,
		Expression: string(expr),
		Cooldown:   time.Hour,
	})

	require.NoError(t, err)
	assert.Equal(t, "AAPL above 150", rule.Name)
	assert.True(t, rule.Active)
}

func TestNewAlertRule_RejectsEmptyName(t *testing.T) {
	t.Parallel()
	_, err := domain.NewAlertRule(&domain.AlertRuleInput{
		Name: "", Expression: `{"op":"price_above"}`,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidAlertRule)
}

func TestNewAlertRule_RejectsInvalidJSON(t *testing.T) {
	t.Parallel()
	_, err := domain.NewAlertRule(&domain.AlertRuleInput{
		Name: "x", Expression: `not-json`,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidAlertRule)
}

func TestNewAlertRule_DefaultsCooldown(t *testing.T) {
	t.Parallel()
	rule, err := domain.NewAlertRule(&domain.AlertRuleInput{
		Name: "x", Expression: `{"op":"price_above"}`, Cooldown: 0,
	})
	require.NoError(t, err)
	assert.Equal(t, time.Hour, rule.Cooldown)
}
