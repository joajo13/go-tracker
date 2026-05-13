// Package scheduler enqueues poll jobs into a buffered channel at the cadence
// declared by each Ticker. Time is injected via clock.Clock so tests can drive
// it deterministically.
package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/joajo13/go-tracker/internal/clock"
	"github.com/joajo13/go-tracker/internal/domain"
	"github.com/joajo13/go-tracker/internal/sources"
)

// Config configures a Scheduler.
type Config struct {
	Clock clock.Clock
	Jobs  chan<- sources.PollJob
}

// group holds all tickers sharing the same PollInterval plus the wall-clock
// time at which the next batch should fire.
type group struct {
	interval time.Duration
	tickers  []*domain.Ticker
	nextAt   time.Time
}

// Scheduler emits PollJobs into Config.Jobs on every interval boundary.
type Scheduler struct {
	cfg Config

	mu     sync.Mutex
	groups map[time.Duration]*group // keyed by PollInterval

	wake chan struct{}
}

// New returns a ready Scheduler.
func New(cfg Config) *Scheduler {
	return &Scheduler{
		cfg:    cfg,
		groups: make(map[time.Duration]*group),
		wake:   make(chan struct{}, 1),
	}
}

// Add registers (or replaces) a ticker. If the ticker's interval group does
// not exist yet, nextAt is set to now+interval so the group fires one interval
// from the moment it was first created.
func (s *Scheduler) Add(t *domain.Ticker) {
	s.mu.Lock()
	g, ok := s.groups[t.PollInterval]
	if !ok {
		now := s.cfg.Clock.Now()
		g = &group{interval: t.PollInterval, nextAt: now.Add(t.PollInterval)}
		s.groups[t.PollInterval] = g
	}
	// Replace existing entry for this ticker ID, or append.
	replaced := false
	for i := range g.tickers {
		if g.tickers[i].ID == t.ID {
			copied := *t
			g.tickers[i] = &copied
			replaced = true
			break
		}
	}
	if !replaced {
		copied := *t
		g.tickers = append(g.tickers, &copied)
	}
	s.mu.Unlock()
	s.notify()
}

// Remove unregisters a ticker by ID.
func (s *Scheduler) Remove(id domain.TickerID) {
	s.mu.Lock()
	for iv, g := range s.groups {
		filtered := g.tickers[:0]
		for i := range g.tickers {
			if g.tickers[i].ID != id {
				filtered = append(filtered, g.tickers[i])
			}
		}
		g.tickers = filtered
		if len(g.tickers) == 0 {
			delete(s.groups, iv)
		}
	}
	s.mu.Unlock()
	s.notify()
}

func (s *Scheduler) notify() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

// snapshot returns a shallow copy of the current groups map so the run loop
// can iterate without holding the lock.
func (s *Scheduler) snapshot() map[time.Duration]*group {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[time.Duration]*group, len(s.groups))
	for k, v := range s.groups {
		out[k] = v
	}
	return out
}

// Run blocks until ctx is done.
//
// Scheduling loop:
//  1. Find the group whose nextAt is soonest.
//  2. Wait until that target passes (via clock.After) or until the ticker set
//     changes (via wake channel).
//  3. Fire all jobs in the due group, then reset its nextAt.
//
// If the injected clock implements clock.Subscriber the scheduler also wakes
// on every Advance() call so deterministic fake-clock tests work without
// real-time sleeps.
func (s *Scheduler) Run(ctx context.Context) error {
	// Subscribe to clock advances if the clock supports it (e.g. Fake).
	var clockChanged <-chan struct{}
	if sub, ok := s.cfg.Clock.(clock.Subscriber); ok {
		clockChanged = sub.Subscribe()
	}

	for {
		groups := s.snapshot()
		nextAt, nextGroup := s.nextFire(groups)

		if nextAt.IsZero() {
			// No tickers — wait for an Add or context cancel.
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-s.wake:
			case <-clockChanged:
			}
			continue
		}

		delay := nextAt.Sub(s.cfg.Clock.Now())
		if delay <= 0 {
			// Already due — fire without blocking.
			s.fireGroup(nextGroup)
			continue
		}

		timer := s.cfg.Clock.After(delay)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.wake:
			// Ticker set changed; loop to re-snapshot.
		case <-clockChanged:
			// Clock was advanced; loop to re-check delay.
		case <-timer:
			s.fireGroup(nextGroup)
		}
	}
}

func (s *Scheduler) nextFire(groups map[time.Duration]*group) (time.Time, *group) {
	var bestAt time.Time
	var bestGroup *group
	for _, g := range groups {
		if bestAt.IsZero() || g.nextAt.Before(bestAt) {
			bestAt = g.nextAt
			bestGroup = g
		}
	}
	return bestAt, bestGroup
}

func (s *Scheduler) fireGroup(g *group) {
	if g == nil {
		return
	}
	s.mu.Lock()
	// Re-fetch the live group so we emit the latest ticker list.
	live := s.groups[g.interval]
	if live == nil {
		s.mu.Unlock()
		return
	}
	tickers := make([]*domain.Ticker, len(live.tickers))
	copy(tickers, live.tickers)
	s.mu.Unlock()

	for i := range tickers {
		t := tickers[i]
		for _, src := range t.Sources {
			job := sources.PollJob{TickerID: t.ID, Symbol: t.Symbol, Source: src}
			select {
			case s.cfg.Jobs <- job:
			default:
				slog.Warn("scheduler_jobs_full", "ticker_id", t.ID, "source", src)
			}
		}
	}

	// Advance nextAt on the live group.
	s.mu.Lock()
	if live2 := s.groups[g.interval]; live2 != nil {
		live2.nextAt = s.cfg.Clock.Now().Add(g.interval)
	}
	s.mu.Unlock()
}
