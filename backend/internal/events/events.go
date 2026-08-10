// Package events provides the in-process event bus (WebSocket fan-out) and
// the NATS publisher used by the outbox relay.
package events

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/aeroxe/sign-flow/backend/internal/natsx"
)

// Event is a domain event delivered on the bus.
type Event struct {
	EventType  string          `json:"event_type"`
	OccurredAt string          `json:"occurred_at"`
	Data       json.RawMessage `json:"data"`
}

// Subscriber receives domain events.
type Subscriber interface {
	OnEvent(ev Event)
}

// Bus fans events out to subscribers (in-process, non-blocking).
type Bus struct {
	mu   sync.RWMutex
	subs []Subscriber
}

// NewBus builds an empty bus.
func NewBus() *Bus { return &Bus{} }

// Subscribe registers a subscriber.
func (b *Bus) Subscribe(s Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs = append(b.subs, s)
}

// Publish broadcasts an event to all subscribers. Delivery runs in a single
// goroutine per event (bounded), so a burst of events cannot spawn unbounded
// goroutines. Subscribers must not block (the WS hub uses buffered channels).
func (b *Bus) Publish(ev Event) {
	b.mu.RLock()
	subs := append([]Subscriber(nil), b.subs...)
	b.mu.RUnlock()
	if len(subs) == 0 {
		return
	}
	go func() {
		for _, s := range subs {
			s.OnEvent(ev)
		}
	}()
}

// NATSPublisher delivers outbox events to NATS JetStream.
type NATSPublisher struct {
	client *natsx.Client
	prefix string
}

// NewNATSPublisher builds a publisher. A nil client is allowed for tests and
// degrades to an error so the outbox relay keeps events pending (no silent loss).
func NewNATSPublisher(client *natsx.Client) *NATSPublisher {
	return &NATSPublisher{client: client, prefix: natsx.SubjectPrefix}
}

// Publish sends a payload to subject <prefix>.<event-type>. Returns an error
// (never nil) while NATS is unavailable so events are retried, not dropped.
func (p *NATSPublisher) Publish(ctx context.Context, subject string, payload []byte) error {
	if p.client == nil {
		return natsx.ErrNotConnected
	}
	return p.client.Publish(subject, payload)
}
