// Package broadcaster fans price events out to multiple consumers without
// coupling them. Slow consumers do NOT block the producer — events are dropped
// per-subscriber when their channel is full.
package broadcaster

import (
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/joajo13/go-tracker/internal/domain"
)

// Broadcaster is the interface declared by package consumers.
type Broadcaster interface {
	Subscribe(name string, buf int) <-chan domain.PriceEvent
	Unsubscribe(name string)
	Publish(ev *domain.PriceEvent)
}

type subscriber struct {
	ch      chan domain.PriceEvent
	dropped atomic.Int64
}

// Hub is the concrete Broadcaster.
type Hub struct {
	mu   sync.RWMutex
	subs map[string]*subscriber
}

// New returns a ready-to-use Hub.
func New() *Hub {
	return &Hub{subs: make(map[string]*subscriber)}
}

// Subscribe registers a named subscriber and returns the receive channel.
func (h *Hub) Subscribe(name string, buf int) <-chan domain.PriceEvent {
	if buf <= 0 {
		buf = 1
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	sub := &subscriber{ch: make(chan domain.PriceEvent, buf)}
	h.subs[name] = sub
	return sub.ch
}

// Unsubscribe removes a named subscriber and closes its channel.
func (h *Hub) Unsubscribe(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if sub, ok := h.subs[name]; ok {
		close(sub.ch)
		delete(h.subs, name)
	}
}

// Publish fans the event out non-blockingly to every subscriber. Events that
// can't be delivered are counted as dropped for that subscriber.
func (h *Hub) Publish(ev *domain.PriceEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for name, sub := range h.subs {
		select {
		case sub.ch <- *ev:
		default:
			sub.dropped.Add(1)
			slog.Debug("broadcaster_drop", "subscriber", name)
		}
	}
}

// DroppedFor reports how many events were dropped for the named subscriber.
// Returns 0 for an unknown subscriber.
func (h *Hub) DroppedFor(name string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if sub, ok := h.subs[name]; ok {
		return int(sub.dropped.Load())
	}
	return 0
}
