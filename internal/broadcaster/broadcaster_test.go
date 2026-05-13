package broadcaster_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joajo13/go-tracker/internal/broadcaster"
	"github.com/joajo13/go-tracker/internal/domain"
)

func mkEvent(t *testing.T, source string) domain.PriceEvent {
	t.Helper()
	p, err := domain.NewPrice(&domain.PriceInput{
		TickerID: 1, Source: source, Price: "1",
		Currency: domain.CurrencyUSD, Ts: time.Now().UTC(),
	})
	require.NoError(t, err)
	return domain.PriceEvent{Price: p}
}

func TestBroadcaster_FanOut(t *testing.T) {
	t.Parallel()

	b := broadcaster.New()
	a := b.Subscribe("a", 4)
	c := b.Subscribe("b", 4)

	ev := mkEvent(t, "yahoo")
	b.Publish(&ev)

	select {
	case got := <-a:
		assert.Equal(t, "yahoo", got.Price.Source)
	case <-time.After(time.Second):
		t.Fatal("subscriber a did not receive")
	}
	select {
	case got := <-c:
		assert.Equal(t, "yahoo", got.Price.Source)
	case <-time.After(time.Second):
		t.Fatal("subscriber b did not receive")
	}
}

func TestBroadcaster_DropOnFullDoesNotBlock(t *testing.T) {
	t.Parallel()

	b := broadcaster.New()
	_ = b.Subscribe("slow", 1) // never drain

	done := make(chan struct{})
	go func() {
		for range 100 {
			ev := mkEvent(t, "yahoo")
			b.Publish(&ev)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("publish blocked on a full subscriber")
	}

	assert.GreaterOrEqual(t, b.DroppedFor("slow"), 99)
}

func TestBroadcaster_UnsubscribeStopsDelivery(t *testing.T) {
	t.Parallel()

	b := broadcaster.New()
	ch := b.Subscribe("x", 1)
	b.Unsubscribe("x")

	ev2 := mkEvent(t, "yahoo")
	b.Publish(&ev2)

	select {
	case _, ok := <-ch:
		assert.False(t, ok, "channel should be closed after unsubscribe")
	case <-time.After(100 * time.Millisecond):
	}
}
