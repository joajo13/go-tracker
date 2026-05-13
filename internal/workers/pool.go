// Package workers runs a bounded pool of goroutines that poll sources and
// publish results onto the broadcaster.
package workers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	"github.com/joajo13/go-tracker/internal/broadcaster"
	"github.com/joajo13/go-tracker/internal/domain"
	"github.com/joajo13/go-tracker/internal/sources"
)

// PoolConfig configures a Pool.
type PoolConfig struct {
	Size        int
	Broadcaster broadcaster.Broadcaster
	Sources     map[string]sources.PriceSource
	JobTimeout  time.Duration
}

// Pool is a fixed-size goroutine pool that consumes PollJobs.
type Pool struct {
	cfg PoolConfig
}

// NewPool builds a Pool. Defaults: Size=10, JobTimeout=10s.
func NewPool(cfg PoolConfig) *Pool {
	if cfg.Size <= 0 {
		cfg.Size = 10
	}
	if cfg.JobTimeout <= 0 {
		cfg.JobTimeout = 10 * time.Second
	}
	return &Pool{cfg: cfg}
}

// Run blocks until ctx is done, consuming jobs and publishing events.
func (p *Pool) Run(ctx context.Context, jobs <-chan sources.PollJob) error {
	var wg sync.WaitGroup
	wg.Add(p.cfg.Size)
	for i := 0; i < p.cfg.Size; i++ {
		go func(id int) {
			defer wg.Done()
			p.worker(ctx, id, jobs)
		}(i)
	}
	wg.Wait()
	return ctx.Err()
}

func (p *Pool) worker(ctx context.Context, id int, jobs <-chan sources.PollJob) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}
			p.handle(ctx, id, job)
		}
	}
}

func (p *Pool) handle(ctx context.Context, id int, job sources.PollJob) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("worker_panic",
				"worker", id,
				"ticker_id", job.TickerID,
				"source", job.Source,
				"panic", fmt.Sprint(r),
				"stack", string(debug.Stack()))
		}
	}()

	src, ok := p.cfg.Sources[job.Source]
	if !ok {
		slog.Warn("unknown_source", "source", job.Source, "ticker_id", job.TickerID)
		return
	}

	fetchCtx, cancel := context.WithTimeout(ctx, p.cfg.JobTimeout)
	defer cancel()

	price, err := src.Fetch(fetchCtx, job.Symbol)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			slog.Warn("fetch_failed",
				"source", job.Source,
				"symbol", job.Symbol,
				"ticker_id", job.TickerID,
				"err", err)
		}
		return
	}

	// The adapter doesn't know the persistent ticker ID; the worker stamps it.
	price.TickerID = job.TickerID

	ev := domain.PriceEvent{Price: price}
	p.cfg.Broadcaster.Publish(&ev)
}
