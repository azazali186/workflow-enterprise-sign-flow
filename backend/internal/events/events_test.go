package events

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aeroxe/sign-flow/backend/internal/natsx"
)

type recorder struct {
	ch chan Event
}

func (r *recorder) OnEvent(ev Event) { r.ch <- ev }

func TestBusDeliversToSubscribers(t *testing.T) {
	bus := NewBus()
	sub := &recorder{ch: make(chan Event, 4)}
	bus.Subscribe(sub)

	ev := Event{EventType: "contract_created", OccurredAt: time.Now().UTC().Format(time.RFC3339), Data: json.RawMessage(`{"contract_id":"c1"}`)}
	bus.Publish(ev)

	select {
	case got := <-sub.ch:
		assert.Equal(t, "contract_created", got.EventType)
	case <-time.After(time.Second):
		t.Fatal("subscriber did not receive the event")
	}
}

func TestNATSPublisherErrorsWithoutClient(t *testing.T) {
	p := NewNATSPublisher(nil)
	// No silent event loss: while NATS is unavailable Publish must error so
	// the outbox relay keeps the event pending and retries with backoff.
	require.ErrorIs(t, p.Publish(context.Background(), "signflow.events.contract_created", []byte(`{}`)), natsx.ErrNotConnected)
}
