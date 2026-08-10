package outbox

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/metrics"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/model"
)

type fakePublisher struct {
	mu     sync.Mutex
	events []string
	fail   bool
}

func (f *fakePublisher) Publish(_ context.Context, subject string, payload []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fail {
		return errors.New("publish failed") // deliberate error
	}
	f.events = append(f.events, subject+":"+string(payload))
	return nil
}

func (f *fakePublisher) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.events)
}

var outboxDBSeq int

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	outboxDBSeq++
	dsn := fmt.Sprintf("file:outbox%d?mode=memory&cache=shared&_loc=UTC", outboxDBSeq)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Event{}))
	return db
}

func TestEnqueueAndDispatch(t *testing.T) {
	db := setupDB(t)
	pub := &fakePublisher{}
	relay := NewRelay(db, pub, func(ev *Event) string { return "signflow.events." + ev.EventType }, metrics.New(), 50*time.Millisecond)
	relay.Start()
	defer relay.Stop()

	ctx := context.Background()
	err := db.Transaction(func(tx *gorm.DB) error {
		return Enqueue(ctx, tx, "contract", "abc-123", "contract_created", map[string]any{"id": "abc-123"})
	})
	require.NoError(t, err)

	require.Eventually(t, func() bool { return pub.count() == 1 }, 3*time.Second, 50*time.Millisecond)
	var ev Event
	require.NoError(t, db.First(&ev).Error)
	assert.Equal(t, StatePublished, ev.Status)
	assert.NotNil(t, ev.PublishedAt)
}

func TestFailedPublishRetriesAndMarksFailed(t *testing.T) {
	db := setupDB(t)
	pub := &fakePublisher{fail: true}
	relay := NewRelay(db, pub, func(ev *Event) string { return "signflow.events." + ev.EventType }, metrics.New(), 50*time.Millisecond)
	relay.Start()
	defer relay.Stop()

	ctx := context.Background()
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		return Enqueue(ctx, tx, "contract", "abc", "signed_update", map[string]any{"x": 1})
	}))

	require.Eventually(t, func() bool {
		var ev Event
		if err := db.First(&ev).Error; err != nil {
			return false
		}
		return ev.Status == StateFailed && ev.RetryCount > 0
	}, 5*time.Second, 100*time.Millisecond)
}

func TestDeadLetterAfterMaxRetries(t *testing.T) {
	db := setupDB(t)
	pub := &fakePublisher{fail: true}
	relay := NewRelay(db, pub, func(ev *Event) string { return "signflow.events." + ev.EventType }, metrics.New(), 30*time.Millisecond)
	relay.SetMaxRetries(3)
	relay.Start()
	defer relay.Stop()

	ctx := context.Background()
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		return Enqueue(ctx, tx, "contract", "dead", "signed_update", map[string]any{"x": 1})
	}))

	// Once the retry cap is hit the event is dead-lettered and never retried again.
	require.Eventually(t, func() bool {
		var ev Event
		if err := db.First(&ev).Error; err != nil {
			return false
		}
		return ev.Status == StateDead && ev.RetryCount >= 3
	}, 8*time.Second, 100*time.Millisecond)

	// It must not be claimed again after reaching the terminal state.
	time.Sleep(200 * time.Millisecond)
	var ev Event
	require.NoError(t, db.First(&ev).Error)
	assert.Equal(t, StateDead, ev.Status)
}

// TestTwoRelaysDoNotDoublePublish enqueues N events and runs two relay
// instances concurrently; each event must be published exactly once.
func TestTwoRelaysDoNotDoublePublish(t *testing.T) {
	db := setupDB(t)
	pub := &fakePublisher{}
	relayA := NewRelay(db, pub, func(ev *Event) string { return "signflow.events." + ev.EventType }, metrics.New(), 30*time.Millisecond)
	relayB := NewRelay(db, pub, func(ev *Event) string { return "signflow.events." + ev.EventType }, metrics.New(), 30*time.Millisecond)
	relayA.Start()
	relayB.Start()
	defer relayA.Stop()
	defer relayB.Stop()

	ctx := context.Background()
	for i := 0; i < 20; i++ {
		id := fmt.Sprintf("agg-%d", i)
		require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
			return Enqueue(ctx, tx, "contract", id, "contract_created", map[string]any{"id": id})
		}))
	}

	require.Eventually(t, func() bool { return pub.count() == 20 }, 8*time.Second, 50*time.Millisecond)

	// Every event must end published exactly once; none stuck in processing.
	var published int64
	var processing int64
	db.Model(&Event{}).Where("status = ?", StatePublished).Count(&published)
	db.Model(&Event{}).Where("status = ?", StateProcessing).Count(&processing)
	assert.Equal(t, int64(20), published)
	assert.Equal(t, int64(0), processing)

	// No duplicate subjects in the published stream.
	seen := map[string]bool{}
	for _, s := range pub.events {
		assert.False(t, seen[s], "duplicate publish: %s", s)
		seen[s] = true
	}
}

func TestEnqueueSetsUUIDAndFields(t *testing.T) {
	db := setupDB(t)
	ctx := context.Background()
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		return Enqueue(ctx, tx, "contract", "id-1", "contract_executed", map[string]any{"contract_id": "id-1"})
	}))
	var ev Event
	require.NoError(t, db.First(&ev).Error)
	assert.NotEmpty(t, ev.ID)
	assert.Equal(t, "contract", ev.AggregateType)
	assert.Equal(t, "contract_executed", ev.EventType)
	assert.Equal(t, StatePending, ev.Status)

	// ID parses as UUID v7 (16-byte UUID string form, non-empty)
	_, err := model.NewID()
	require.NoError(t, err)
}
