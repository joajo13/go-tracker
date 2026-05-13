package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// AlertRuleID is the persistent identifier of an AlertRule.
type AlertRuleID int64

// ErrInvalidAlertRule is the sentinel wrapped by NewAlertRule failures.
var ErrInvalidAlertRule = errors.New("invalid alert rule")

// DefaultAlertCooldown is the default cooldown applied when none is supplied.
const DefaultAlertCooldown = time.Hour

// AlertRule is a configured rule that the evaluator will check against price
// events. Expression is a JSON AST — its grammar is owned by the evaluator
// package and not this struct.
type AlertRule struct {
	ID         AlertRuleID
	Name       string
	TickerID   TickerID // 0 means "portfolio-global"
	Expression string   // raw JSON AST
	Cooldown   time.Duration
	Active     bool
	CreatedAt  time.Time
}

// AlertRuleInput is the unvalidated payload accepted by NewAlertRule.
type AlertRuleInput struct {
	Name       string
	TickerID   TickerID
	Expression string
	Cooldown   time.Duration
}

// NewAlertRule validates and returns an AlertRule with Active=true.
func NewAlertRule(in *AlertRuleInput) (AlertRule, error) {
	if strings.TrimSpace(in.Name) == "" {
		return AlertRule{}, fmt.Errorf("%w: name required", ErrInvalidAlertRule)
	}
	if strings.TrimSpace(in.Expression) == "" {
		return AlertRule{}, fmt.Errorf("%w: expression required", ErrInvalidAlertRule)
	}
	var probe any
	if err := json.Unmarshal([]byte(in.Expression), &probe); err != nil {
		return AlertRule{}, fmt.Errorf("%w: expression not JSON: %w", ErrInvalidAlertRule, err)
	}
	cooldown := in.Cooldown
	if cooldown <= 0 {
		cooldown = DefaultAlertCooldown
	}
	return AlertRule{
		Name:       in.Name,
		TickerID:   in.TickerID,
		Expression: in.Expression,
		Cooldown:   cooldown,
		Active:     true,
	}, nil
}
