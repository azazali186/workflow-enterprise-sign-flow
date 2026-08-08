// Package outbox implements the transactional outbox pattern. Business writes
// enqueue rows in the same transaction; a relay worker publishes them to NATS
// with retry/backoff behind a circuit breaker.
package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/breaker"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/logger"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/metrics"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/model"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/retryutil"
	"go.uber.org/zap"
)

// Event states.
const (
	StatePending    = "pending"
	StateProcessing = "processing"
	StatePublished  = "published"
	StateFailed     = "failed"
)

// Event is an outbox row: an intent to publish a domain event.
type Event struct {
	model.Base
	AggregateType string     `gorm:"size:60;index" json:"aggregate_type"`
	AggregateID   string     `gorm:"size:60;index" json:"aggregate_id"`
	EventType     string     `gorm:"size:120;index" json:"event_type"`
	Payload       string     `gorm:"type:text" json:"payload"`
	Status        string     `gorm:"size:20;index" json:"status"`
	RetryCount    int        `gorm:"default:0" json:"retry_count"`
	AvailableAt   time.Time  `json:"available_at"`
	ClaimToken    string     `gorm:"size:40;index" json:"-"` // multi-instance claim ownership
	ClaimedAt     *time.Time `json:"-"`
	LastError     string     `gorm:"type:text" json:"-"`
	PublishedAt   *time.Time `json:"published_at"`
}

// Publisher delivers a single event (NATS publish, webhook, ...).
type Publisher interface {
	Publish(ctx context.Context, subject string, payload []byte) error
}

// Enqueue inserts an event inside the caller's transaction.
func Enqueue(ctx context.Context, tx *gorm.DB, aggregateType, aggregateID, eventType string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	ev := Event{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		Payload:       string(raw),
		Status:        StatePending,
		AvailableAt:   time.Now(),
	}
	return tx.WithContext(ctx).Create(&ev).Error
}

// Relay delivers pending events with retry, backoff and a circuit breaker.
type Relay struct {
	db         *gorm.DB
	pub        Publisher
	subject    func(ev *Event) string
	breaker    *breaker.Breaker
	metrics    *metrics.Collectors
	poll       time.Duration
	claimToken string // unique per process for multi-instance safe claiming
	stop       chan struct{}
}

// NewRelay builds the relay worker.
func NewRelay(db *gorm.DB, pub Publisher, subject func(ev *Event) string, met *metrics.Collectors, poll time.Duration) *Relay {
	tok, err := model.NewID()
	if err != nil {
		tok = "relay-" + fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return &Relay{
		db:         db,
		pub:        pub,
		subject:    subject,
		breaker:    breaker.New(breaker.Defaults("nats-publish")),
		metrics:    met,
		poll:       poll,
		claimToken: tok,
		stop:       make(chan struct{}),
	}
}

// Start launches the background relay loop.
func (r *Relay) Start() {
	go r.loop()
}

// Stop halts the relay loop.
func (r *Relay) Stop() { close(r.stop) }

func (r *Relay) loop() {
	if r.poll <= 0 {
		r.poll = 2 * time.Second
	}
	ticker := time.NewTicker(r.poll)
	defer ticker.Stop()
	for {
		select {
		case <-r.stop:
			return
		case <-ticker.C:
			r.dispatch(r.poll)
		}
	}
}

// claimStale resets rows stuck in processing (e.g. a crashed replica) so they
// are retried. Runs every poll cycle as a cheap safety net.
func (r *Relay) claimStale(ctx context.Context) {
	threshold := time.Now().Add(-10 * time.Minute)
	r.db.WithContext(ctx).Model(&Event{}).
		Where("status = ? AND claimed_at < ?", StateProcessing, threshold).
		Updates(map[string]any{"status": StatePending, "claim_token": "", "claimed_at": nil})
}

// claim atomically marks pending rows as owned by this relay instance using an
// UPDATE ... WHERE id IN (SELECT ...) inside a transaction, so concurrent
// replicas never claim the same events.
func (r *Relay) claim(ctx context.Context) ([]Event, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		return tx.Model(&Event{}).
			Where("id IN (?)",
				tx.Model(&Event{}).Select("id").
					Where("status = ? AND available_at <= ?", StatePending, now).
					Order("created_at asc").Limit(50),
			).
			Updates(map[string]any{"status": StateProcessing, "claim_token": r.claimToken, "claimed_at": now}).Error
	})
	if err != nil {
		return nil, err
	}
	var events []Event
	err = r.db.WithContext(ctx).Where("status = ? AND claim_token = ?", StateProcessing, r.claimToken).
		Order("created_at asc").Find(&events).Error
	return events, err
}

// dispatch publishes up to 50 events claimed by this instance.
func (r *Relay) dispatch(limit time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	r.claimStale(ctx)
	events, err := r.claim(ctx)
	if err != nil || len(events) == 0 {
		return
	}
	for i := range events {
		ev := &events[i]
		if err := r.publishWithRetry(ctx, ev); err != nil {
			r.db.Model(ev).Updates(map[string]any{
				"status": StateFailed, "last_error": err.Error(), "retry_count": gorm.Expr("retry_count + 1"),
				"claim_token": "", "claimed_at": nil,
			})
			r.metrics.OutboxEvents.WithLabelValues(StateFailed).Inc()
			continue
		}
		now := time.Now()
		r.db.Model(ev).Updates(map[string]any{
			"status": StatePublished, "published_at": now, "last_error": "",
			"claim_token": "", "claimed_at": nil,
		})
		r.metrics.OutboxEvents.WithLabelValues(StatePublished).Inc()
	}
}

func (r *Relay) publishWithRetry(ctx context.Context, ev *Event) error {
	cfg := retryutil.Config{InitialInterval: 500 * time.Millisecond, MaxInterval: 5 * time.Second, MaxElapsedTime: r.poll * 3 / 4}
	return retryutil.Notify(ctx, cfg, func() error {
		return r.breaker.Execute(func() error {
			return r.pub.Publish(ctx, r.subject(ev), []byte(ev.Payload))
		})
	}, func(err error, d time.Duration) {
		logger.L().Warn("outbox publish failed, retrying", zap.Error(err), zap.Duration("after", d), zap.String("event", ev.EventType))
	})
}
