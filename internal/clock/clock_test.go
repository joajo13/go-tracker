package clock_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/joajo13/go-tracker/internal/clock"
)

func TestRealClock_NowAdvances(t *testing.T) {
	t.Parallel()

	c := clock.Real{}
	a := c.Now()
	time.Sleep(2 * time.Millisecond)
	b := c.Now()
	assert.True(t, b.After(a))
}

func TestFakeClock_NowReturnsInjectedTime(t *testing.T) {
	t.Parallel()

	fixed := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	c := clock.NewFake(fixed)

	assert.Equal(t, fixed, c.Now())
}

func TestFakeClock_AdvanceMovesNowForward(t *testing.T) {
	t.Parallel()

	c := clock.NewFake(time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC))
	c.Advance(90 * time.Second)

	assert.Equal(t, time.Date(2026, 5, 14, 12, 1, 30, 0, time.UTC), c.Now())
}

func TestFakeClock_AdvanceFiresPendingTimers(t *testing.T) {
	t.Parallel()

	c := clock.NewFake(time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC))
	ch := c.After(30 * time.Second)

	select {
	case <-ch:
		t.Fatal("timer fired too early")
	default:
	}

	c.Advance(30 * time.Second)

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("timer should have fired after Advance")
	}
}
