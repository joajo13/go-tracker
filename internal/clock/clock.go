// Package clock decouples business code from wall-clock time so schedulers and
// timeouts can be exercised deterministically in tests.
package clock

import (
	"sync"
	"time"
)

// Clock is the minimal time interface the rest of the codebase uses.
type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

// Subscriber is an optional extension of Clock that lets callers be notified
// whenever the clock's time changes. Fake implements it so that consumers
// (e.g. a scheduler) can re-evaluate their next fire time after Advance.
type Subscriber interface {
	// Subscribe returns a channel that receives a value every time the clock
	// advances. The channel is buffered (capacity 1) so callers that are busy
	// miss at most one event — they will still notice the change on their next
	// iteration.
	Subscribe() <-chan struct{}
}

// Real is a Clock backed by the standard library.
type Real struct{}

// Now returns the current wall time.
func (Real) Now() time.Time { return time.Now() }

// After is time.After.
func (Real) After(d time.Duration) <-chan time.Time { return time.After(d) }

// Fake is a deterministic Clock for tests.
type Fake struct {
	mu          sync.Mutex
	now         time.Time
	pending     []pendingFire
	subscribers []chan struct{}
}

type pendingFire struct {
	at time.Time
	ch chan time.Time
}

// NewFake returns a Fake clock pinned at the given time.
func NewFake(t time.Time) *Fake {
	return &Fake{now: t}
}

// Now returns the current fake time.
func (f *Fake) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

// After returns a channel that receives once the fake clock has been advanced
// at least d past the point where After was called. If the deadline has
// already passed (d <= 0), the channel fires immediately.
func (f *Fake) After(d time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	f.mu.Lock()
	fireAt := f.now.Add(d)
	if !fireAt.After(f.now) {
		// Already due — fire immediately without registering.
		ch <- f.now
		f.mu.Unlock()
		return ch
	}
	fire := pendingFire{at: fireAt, ch: ch}
	f.pending = append(f.pending, fire)
	f.mu.Unlock()
	return ch
}

// Subscribe returns a channel that receives a struct{} every time Advance is
// called. The channel is buffered with capacity 1; rapid advances may coalesce.
func (f *Fake) Subscribe() <-chan struct{} {
	ch := make(chan struct{}, 1)
	f.mu.Lock()
	f.subscribers = append(f.subscribers, ch)
	f.mu.Unlock()
	return ch
}

// Advance moves the fake clock forward by d, firing any pending timers whose
// scheduled time has now passed, and notifying all subscribers.
func (f *Fake) Advance(d time.Duration) {
	f.mu.Lock()
	f.now = f.now.Add(d)
	ready := f.now
	remaining := f.pending[:0]
	var toFire []pendingFire
	for _, p := range f.pending {
		if !p.at.After(ready) {
			toFire = append(toFire, p)
			continue
		}
		remaining = append(remaining, p)
	}
	f.pending = remaining
	subs := make([]chan struct{}, len(f.subscribers))
	copy(subs, f.subscribers)
	f.mu.Unlock()

	for _, p := range toFire {
		p.ch <- ready
	}
	for _, ch := range subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
