package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joajo13/go-tracker/internal/broadcaster"
	"github.com/joajo13/go-tracker/internal/clock"
	"github.com/joajo13/go-tracker/internal/domain"
	"github.com/joajo13/go-tracker/internal/persistence/sqlite"
	"github.com/joajo13/go-tracker/internal/scheduler"
	"github.com/joajo13/go-tracker/internal/sources"
	"github.com/joajo13/go-tracker/internal/workers"
)

func TestPhase1_PriceTickPersistsToPrices1m(t *testing.T) {
	t.Parallel()

	// --- 0. Fake Yahoo HTTP server ---
	yahooResp := []byte(`{
		"chart": {
			"result": [{
				"meta": {
					"symbol": "AAPL",
					"regularMarketPrice": 172.45,
					"currency": "USD",
					"regularMarketTime": 1747227600
				}
			}],
			"error": null
		}
	}`)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(yahooResp)
	}))
	defer srv.Close()

	// --- 1. DB + repos ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := sqlite.Open(ctx, ":memory:")
	require.NoError(t, err)
	defer db.Close()

	tickerRepo := sqlite.NewTickerRepo(db)
	priceRepo := sqlite.NewPriceRepo(db)

	tk, err := domain.NewTicker(&domain.TickerInput{
		Symbol: "AAPL", Name: "Apple", Type: domain.AssetTypeUSStock,
		PollInterval: time.Minute, Sources: []string{"yahoo"},
	})
	require.NoError(t, err)
	tid, err := tickerRepo.Insert(ctx, &tk)
	require.NoError(t, err)
	tk.ID = tid

	// --- 2. Broadcaster + persistence subscriber ---
	hub := broadcaster.New()
	persistCh := hub.Subscribe("persist", 16)
	go func() {
		for ev := range persistCh {
			price := ev.Price
			_ = priceRepo.Insert(ctx, &price)
		}
	}()

	// --- 3. Workers + Yahoo adapter pointed at the fake server ---
	yahoo := sources.NewYahoo(sources.YahooConfig{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		RatePerSec: 100, RateBurst: 100,
	})
	jobs := make(chan sources.PollJob, 8)
	pool := workers.NewPool(workers.PoolConfig{
		Size: 2, Broadcaster: hub,
		Sources:    map[string]sources.PriceSource{"yahoo": yahoo},
		JobTimeout: 2 * time.Second,
	})
	go func() { _ = pool.Run(ctx, jobs) }()

	// --- 4. Scheduler with FakeClock ---
	fake := clock.NewFake(time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC))
	sched := scheduler.New(scheduler.Config{Clock: fake, Jobs: jobs})
	sched.Add(&tk)
	go func() { _ = sched.Run(ctx) }()

	// --- 5. Drive a tick ---
	// Give the scheduler a moment to enter Run loop before advancing the fake clock.
	time.Sleep(50 * time.Millisecond)
	fake.Advance(time.Minute)

	// --- 6. Assert a row landed in prices_1m ---
	require.Eventually(t, func() bool {
		latest, lerr := priceRepo.LatestByTicker(ctx, tid)
		if lerr != nil {
			return false
		}
		// decimal.String strips trailing zeros; "172.45" stays as "172.45".
		return latest.Price.String() == "172.45"
	}, 3*time.Second, 50*time.Millisecond, "price never persisted")

	latest, err := priceRepo.LatestByTicker(ctx, tid)
	require.NoError(t, err)
	assert.Equal(t, "yahoo", latest.Source)
	assert.Equal(t, domain.CurrencyUSD, latest.Currency)
}
