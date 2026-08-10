// Package audit records change logs (before/after) and login logs. Sensitive
// fields are excluded before anything is persisted.
package audit

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/ctxval"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/logger"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/model"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/safego"
)

// Log is an audit log row capturing a state change.
type Log struct {
	model.Base
	Action        string          `gorm:"size:60;index" json:"action"`
	EntityType    string          `gorm:"size:60;index" json:"entity_type"`
	EntityID      string          `gorm:"size:60;index" json:"entity_id"`
	ActorUserID   string          `gorm:"size:60;index" json:"actor_user_id"`
	ActorName     string          `gorm:"size:120" json:"actor_name"`
	BeforeData    json.RawMessage `gorm:"type:jsonb" json:"before_data,omitempty"`
	AfterData     json.RawMessage `gorm:"type:jsonb" json:"after_data,omitempty"`
	ChangedFields json.RawMessage `gorm:"type:jsonb" json:"changed_fields,omitempty"`
	IP            string          `gorm:"size:64" json:"ip"`
	UserAgent     string          `gorm:"size:255" json:"user_agent"`
	RequestID     string          `gorm:"size:64;index" json:"request_id"`
}

// LoginLog records authentication attempts (never passwords).
type LoginLog struct {
	model.Base
	Username  string    `gorm:"size:160;index" json:"username"`
	IP        string    `gorm:"size:64" json:"ip"`
	UserAgent string    `gorm:"size:255" json:"user_agent"`
	Success   bool      `gorm:"index" json:"success"`
	Message   string    `gorm:"size:255" json:"message"`
	LoginAt   time.Time `gorm:"index" json:"login_at"`
}

// Service persists audit records.
type Service struct {
	db *gorm.DB
}

// New builds the audit service.
func New(db *gorm.DB) *Service { return &Service{db: db} }

// Record stores a change with before/after maps. Maps are JSON-marshalled
// inside the same goroutine; values must already be safe for logging.
func (s *Service) Record(ctx context.Context, action, entityType, entityID string, before, after map[string]any) error {
	fields := diffKeys(before, after)
	beforeRaw, err := safeJSON(before)
	if err != nil {
		return err
	}
	afterRaw, err := safeJSON(after)
	if err != nil {
		return err
	}
	fieldsRaw, err := safeJSON(fields)
	if err != nil {
		return err
	}
	entry := Log{
		Action:        action,
		EntityType:    entityType,
		EntityID:      entityID,
		ActorUserID:   ctxval.UserID(ctx),
		ActorName:     ctxval.UserName(ctx),
		BeforeData:    beforeRaw,
		AfterData:     afterRaw,
		ChangedFields: fieldsRaw,
		IP:            ctxval.IP(ctx),
		UserAgent:     ctxval.UserAgent(ctx),
		RequestID:     ctxval.RequestID(ctx),
	}
	return s.db.WithContext(ctx).Create(&entry).Error
}

// RecordLogin logs an authentication attempt.
func (s *Service) RecordLogin(ctx context.Context, username string, success bool, message string) error {
	entry := LoginLog{
		Username:  username,
		IP:        ctxval.IP(ctx),
		UserAgent: ctxval.UserAgent(ctx),
		Success:   success,
		Message:   message,
		LoginAt:   time.Now(),
	}
	return s.db.WithContext(ctx).Create(&entry).Error
}

// PurgeOlderThan physically deletes audit and login log rows older than
// cutoff. Unscoped is required (both models carry a soft-delete column);
// otherwise rows would merely be marked deleted and the tables would keep
// growing. Called periodically by the retention cleaner.
func (s *Service) PurgeOlderThan(ctx context.Context, cutoff time.Time) error {
	if err := s.db.WithContext(ctx).Unscoped().Where("created_at < ?", cutoff).Delete(&Log{}).Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Unscoped().Where("login_at < ?", cutoff).Delete(&LoginLog{}).Error
}

// StartCleaner runs a background goroutine that purges audit and login logs
// older than retention, every interval. Panic-safe; stops when ctx is done.
func (s *Service) StartCleaner(ctx context.Context, interval, retention time.Duration) {
	safego.Go(func() {
		if interval <= 0 {
			interval = time.Hour
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// retention <= 0 means "keep everything" — never purge, so a
				// misconfigured zero cannot wipe the entire audit trail.
				if retention <= 0 {
					continue
				}
				cutoff := time.Now().Add(-retention)
				if err := s.PurgeOlderThan(ctx, cutoff); err != nil {
					logger.L().Warn("audit retention purge failed", zap.Error(err))
				}
			}
		}
	})
}

// diffKeys returns the keys that differ between before and after.
func diffKeys(before, after map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range after {
		if bv, ok := before[k]; !ok || !jsonEqual(bv, v) {
			out[k] = v
		}
	}
	keys := make([]string, 0, len(out))
	for k := range out {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	res := map[string]any{}
	for _, k := range keys {
		res[k] = out[k]
	}
	return res
}

func jsonEqual(a, b any) bool {
	ar, err1 := json.Marshal(a)
	br, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return string(ar) == string(br)
}

func safeJSON(m map[string]any) (json.RawMessage, error) {
	if len(m) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return b, nil
}
