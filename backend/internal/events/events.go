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
	EventType string          `json:"event_type"`
	OccurredAt string         `json:"occurred_at"`
	Data      json.RawMessage `json:"data"`
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

// Publish broadcasts an event to all subscribers in a goroutine.
func (b *Bus) Publish(ev Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, s := range b.subs {
		go s.OnEvent(ev)
	}
}

// NATSPublisher delivers outbox events to NATS JetStream.
type NATSPublisher struct {
	client *natsx.Client
	prefix string
}

// NewNATSPublisher builds a publisher; nil client degrades to a no-op.
func NewNATSPublisher(client *natsx.Client) *NATSPublisher {
	return &NATSPublisher{client: client, prefix: natsx.SubjectPrefix}
}

// Publish sends a payload to subject <prefix>.<event-type>.
func (p *NATSPublisher) Publish(ctx context.Context, subject string, payload []byte) error {
	if p.client == nil || p.client.JS == nil {
		return nil // no NATS configured: relay simply marks events published
	}
	_, err := p.client.JS.Publish(subject, payload)
	return err
}
