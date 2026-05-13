package workers_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joajo13/go-tracker/internal/broadcaster"
	"github.com/joajo13/go-tracker/internal/domain"
	"github.com/joajo13/go-tracker/internal/sources"
	"github.com/joajo13/go-tracker/internal/workers"
)

type stubSource struct {
	mu      sync.Mutex
	name    string
	calls   int
	respond func(symbol string) (domain.Price, error)
}

func (s *stubSource) Name() string { return s.name }
func (s *stubSource) Fetch(_ context.Context, symbol string) (domain.Price, error) {
	s.mu.Lock()
	s.calls++
	s.mu.Unlock()
	return s.respond(symbol)
}

func makePrice(t *testing.T, sym string, tickerID domain.TickerID) domain.Price {
	t.Helper()
	p, err := domain.NewPrice(&domain.PriceInput{
		TickerID: tickerID, Source: sym, Price: "10",
		Currency: domain.CurrencyUSD, Ts: time.Now().UTC(),
	})
	require.NoError(t, err)
	return p
}

func TestPool_FetchAndPublish(t *testing.T) {
	t.Parallel()

	src := &stubSource{
		name: "yahoo",
		respond: func(_ string) (domain.Price, error) {
			return makePrice(t, "yahoo", 99), nil
		},
	}
	hub := broadcaster.New()
	out := hub.Subscribe("test", 4)

	pool := workers.NewPool(workers.PoolConfig{
		Size:        2,
		Broadcaster: hub,
		Sources:     map[string]sources.PriceSource{"yahoo": src},
		JobTimeout:  time.Second,
	})

	jobs := make(chan sources.PollJob, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		_ = pool.Run(ctx, jobs)
		close(done)
	}()

	jobs <- sources.PollJob{TickerID: 42, Symbol: "AAPL", Source: "yahoo"}

	select {
	case ev := <-out:
		assert.Equal(t, domain.TickerID(42), ev.Price.TickerID)
		assert.Equal(t, "yahoo", ev.Price.Source)
	case <-time.After(2 * time.Second):
		t.Fatal("never received event")
	}

	cancel()
	<-done
}

func TestPool_RecoversFromPanic(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	src := &stubSource{
		name: "yahoo",
		respond: func(_ string) (domain.Price, error) {
			n := calls.Add(1)
			if n == 1 {
				panic("boom")
			}
			return makePrice(t, "yahoo", 1), nil
		},
	}
	hub := broadcaster.New()
	out := hub.Subscribe("test", 4)

	pool := workers.NewPool(workers.PoolConfig{
		Size: 1, Broadcaster: hub,
		Sources:    map[string]sources.PriceSource{"yahoo": src},
		JobTimeout: time.Second,
	})

	jobs := make(chan sources.PollJob, 2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = pool.Run(ctx, jobs) }()

	jobs <- sources.PollJob{TickerID: 1, Symbol: "A", Source: "yahoo"}
	jobs <- sources.PollJob{TickerID: 2, Symbol: "A", Source: "yahoo"}

	select {
	case <-out:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not recover from panic")
	}
}

func TestPool_FetchErrorIsLoggedNotFatal(t *testing.T) {
	t.Parallel()

	src := &stubSource{
		name: "yahoo",
		respond: func(_ string) (domain.Price, error) {
			return domain.Price{}, errors.New("boom")
		},
	}
	hub := broadcaster.New()
	out := hub.Subscribe("test", 4)

	pool := workers.NewPool(workers.PoolConfig{
		Size: 1, Broadcaster: hub,
		Sources:    map[string]sources.PriceSource{"yahoo": src},
		JobTimeout: time.Second,
	})

	jobs := make(chan sources.PollJob, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = pool.Run(ctx, jobs) }()

	jobs <- sources.PollJob{TickerID: 1, Symbol: "A", Source: "yahoo"}

	select {
	case <-out:
		t.Fatal("event should NOT have been published on fetch error")
	case <-time.After(200 * time.Millisecond):
	}
}
