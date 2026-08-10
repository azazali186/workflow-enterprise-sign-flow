package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/ctxval"
)

var dbSeq int

func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbSeq++
	dsn := fmt.Sprintf("file:auditmem%d?mode=memory&cache=shared&_loc=UTC", dbSeq)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Log{}, &LoginLog{}))
	return db
}

func TestDiffKeysOnlyChanged(t *testing.T) {
	before := map[string]any{"status": "draft", "title": "Same", "secret": "plaintext"}
	after := map[string]any{"status": "signed", "title": "Same", "note": "new"}
	diff := diffKeys(before, after)
	assert.Equal(t, map[string]any{"status": "signed", "note": "new"}, diff)
}

func TestRecordPersistsBeforeAfter(t *testing.T) {
	db := newDB(t)
	svc := New(db)
	ctx := ctxval.SetUserID(ctxval.SetRequestID(context.Background(), "req-1"), "actor-1")

	err := svc.Record(ctx, "contract.patched", "contract", "c-1",
		map[string]any{"status": "draft"},
		map[string]any{"status": "signed"})
	require.NoError(t, err)

	var log Log
	require.NoError(t, db.First(&log).Error)
	assert.Equal(t, "contract.patched", log.Action)
	assert.Equal(t, "c-1", log.EntityID)
	assert.Equal(t, "actor-1", log.ActorUserID)
	assert.Equal(t, "req-1", log.RequestID)

	var changed map[string]any
	require.NoError(t, json.Unmarshal(log.ChangedFields, &changed))
	assert.Equal(t, "signed", changed["status"])

	var after map[string]any
	require.NoError(t, json.Unmarshal(log.AfterData, &after))
	assert.Equal(t, "signed", after["status"])
}

func TestPurgeOlderThanPhysicallyDeletes(t *testing.T) {
	db := newDB(t)
	svc := New(db)
	ctx := context.Background()
	require.NoError(t, svc.Record(ctx, "a.b", "t", "1", nil, map[string]any{"x": 1}))
	require.NoError(t, svc.RecordLogin(ctx, "old@example.com", false, "no"))

	// Backdate both rows beyond the retention cutoff.
	old := time.Now().Add(-100 * 24 * time.Hour)
	require.NoError(t, db.Model(&Log{}).Where("1 = 1").Update("created_at", old).Error)
	require.NoError(t, db.Model(&LoginLog{}).Where("1 = 1").Update("login_at", old).Error)

	require.NoError(t, svc.PurgeOlderThan(ctx, time.Now().Add(-90*24*time.Hour)))

	// Unscoped count: rows must be GONE from disk, not merely soft-deleted.
	var n int64
	db.Unscoped().Model(&Log{}).Count(&n)
	assert.Zero(t, n, "audit log must be physically deleted")
	db.Unscoped().Model(&LoginLog{}).Count(&n)
	assert.Zero(t, n, "login log must be physically deleted")
}

func TestPurgeZeroRetentionKeepsEverything(t *testing.T) {
	db := newDB(t)
	svc := New(db)
	ctx := context.Background()
	require.NoError(t, svc.Record(ctx, "a.b", "t", "1", nil, map[string]any{"x": 1}))
	require.NoError(t, svc.RecordLogin(ctx, "old@example.com", false, "no"))

	// Backdate rows so any purge would remove them.
	old := time.Now().Add(-200 * 24 * time.Hour)
	require.NoError(t, db.Model(&Log{}).Where("1 = 1").Update("created_at", old).Error)
	require.NoError(t, db.Model(&LoginLog{}).Where("1 = 1").Update("login_at", old).Error)

	// A zero/negative retention must never delete anything.
	svc.StartCleaner(ctx, time.Nanosecond, 0)
	time.Sleep(50 * time.Millisecond)

	var n int64
	db.Unscoped().Model(&Log{}).Count(&n)
	assert.Equal(t, int64(1), n, "zero retention must keep audit rows")
	db.Unscoped().Model(&LoginLog{}).Count(&n)
	assert.Equal(t, int64(1), n, "zero retention must keep login rows")
}

func TestPurgeKeepsFreshRows(t *testing.T) {
	db := newDB(t)
	svc := New(db)
	ctx := context.Background()
	require.NoError(t, svc.Record(ctx, "a.b", "t", "1", nil, map[string]any{"x": 1}))
	require.NoError(t, svc.RecordLogin(ctx, "new@example.com", true, "ok"))

	require.NoError(t, svc.PurgeOlderThan(ctx, time.Now().Add(-90*24*time.Hour)))

	var n int64
	db.Unscoped().Model(&Log{}).Count(&n)
	assert.Equal(t, int64(1), n)
	db.Unscoped().Model(&LoginLog{}).Count(&n)
	assert.Equal(t, int64(1), n)
}

func TestRecordLoginSuccessAndFailure(t *testing.T) {
	db := newDB(t)
	svc := New(db)
	ctx := ctxval.SetIP(context.Background(), "10.0.0.1")

	require.NoError(t, svc.RecordLogin(ctx, "user@example.com", true, "login success"))
	require.NoError(t, svc.RecordLogin(ctx, "user@example.com", false, "invalid password"))

	var rows []LoginLog
	require.NoError(t, db.Order("created_at asc").Find(&rows).Error)
	require.Len(t, rows, 2)
	assert.True(t, rows[0].Success)
	assert.False(t, rows[1].Success)
	// Credentials/plaintext codes must never be persisted by the audit layer.
	assert.NotContains(t, rows[0].Message, "password")
	assert.Equal(t, "10.0.0.1", rows[0].IP)
}
