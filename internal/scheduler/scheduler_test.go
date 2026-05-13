package scheduler_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joajo13/go-tracker/internal/clock"
	"github.com/joajo13/go-tracker/internal/domain"
	"github.com/joajo13/go-tracker/internal/scheduler"
	"github.com/joajo13/go-tracker/internal/sources"
)

func mkTicker(t *testing.T, sym string, id domain.TickerID, interval time.Duration) domain.Ticker {
	t.Helper()
	tk, err := domain.NewTicker(&domain.TickerInput{
		Symbol: sym, Name: sym, Type: domain.AssetTypeUSStock,
		PollInterval: interval, Sources: []string{"yahoo"},
	})
	require.NoError(t, err)
	tk.ID = id
	return tk
}

func TestScheduler_EmitsJobOnEachTick(t *testing.T) {
	t.Parallel()

	fake := clock.NewFake(time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC))
	jobs := make(chan sources.PollJob, 8)

	s := scheduler.New(scheduler.Config{
		Clock: fake,
		Jobs:  jobs,
	})
	tk := mkTicker(t, "AAPL", 1, time.Minute)
	s.Add(&tk)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		_ = s.Run(ctx)
		close(done)
	}()

	fake.Advance(time.Minute)

	select {
	case job := <-jobs:
		assert.Equal(t, domain.TickerID(1), job.TickerID)
		assert.Equal(t, "AAPL", job.Symbol)
		assert.Equal(t, "yahoo", job.Source)
	case <-time.After(time.Second):
		t.Fatal("scheduler did not emit job")
	}

	cancel()
	<-done
}

func TestScheduler_GroupsByInterval(t *testing.T) {
	t.Parallel()

	fake := clock.NewFake(time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC))
	jobs := make(chan sources.PollJob, 16)
	s := scheduler.New(scheduler.Config{Clock: fake, Jobs: jobs})
	tkAAPL := mkTicker(t, "AAPL", 1, time.Minute)
	tkMSFT := mkTicker(t, "MSFT", 2, time.Minute)
	tkBTC := mkTicker(t, "BTC", 3, 5*time.Minute)
	s.Add(&tkAAPL)
	s.Add(&tkMSFT)
	s.Add(&tkBTC)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { _ = s.Run(ctx); close(done) }()

	fake.Advance(time.Minute)
	// Expect 2 emissions (AAPL + MSFT) but NOT BTC.
	collected := drainFor(t, jobs, 2, time.Second)
	assert.ElementsMatch(t, []string{"AAPL", "MSFT"}, collected)

	cancel()
	<-done
}

func TestScheduler_RemoveStopsEmission(t *testing.T) {
	t.Parallel()

	fake := clock.NewFake(time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC))
	jobs := make(chan sources.PollJob, 4)
	s := scheduler.New(scheduler.Config{Clock: fake, Jobs: jobs})
	tk := mkTicker(t, "AAPL", 1, time.Minute)
	s.Add(&tk)
	s.Remove(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = s.Run(ctx) }()

	fake.Advance(time.Minute)

	select {
	case j := <-jobs:
		t.Fatalf("unexpected job after Remove: %+v", j)
	case <-time.After(200 * time.Millisecond):
	}
}

func drainFor(t *testing.T, jobs <-chan sources.PollJob, want int, maxWait time.Duration) []string {
	t.Helper()
	deadline := time.After(maxWait)
	var symbols []string
	for len(symbols) < want {
		select {
		case j := <-jobs:
			symbols = append(symbols, j.Symbol)
		case <-deadline:
			return symbols
		}
	}
	return symbols
}
