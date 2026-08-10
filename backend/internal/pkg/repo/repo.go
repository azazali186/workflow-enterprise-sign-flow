// Package repo provides a generic, cache-aware GORM repository used by every
// entity. All database access goes through GORM — no raw SQL anywhere.
package repo

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
)

// Cache is the distributed cache interface the repository invalidates/reads.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

// Options configures a repository instance.
type Options[T any] struct {
	Slug        string                         // cache slug, e.g. "contract"
	Searchable  []string                       // columns used for full-text-ish search
	Filterable  map[string]string              // request filter key -> db column
	Sortable    map[string]string              // request sort field -> db column
	DateFields  []string                       // whitelist for date_field
	Preloads    []string                       // associations to preload
	Summary     func(tx *gorm.DB) (any, error) // optional per-list summary data
	DefaultSort string                         // e.g. "-created_at"
}

// Repo is a generic CRUD repository over GORM.
type Repo[T any] struct {
	db    *gorm.DB
	cache Cache
	opts  Options[T]
}

// New builds a repository.
func New[T any](db *gorm.DB, c Cache, opts Options[T]) *Repo[T] {
	if len(opts.DateFields) == 0 {
		opts.DateFields = []string{"created_at", "updated_at"}
	}
	if opts.DefaultSort == "" {
		opts.DefaultSort = "-created_at"
	}
	return &Repo[T]{db: db, cache: c, opts: opts}
}

// DB exposes the underlying connection for transactions.
func (r *Repo[T]) DB() *gorm.DB { return r.db }

func (r *Repo[T]) cacheKey(id string) string {
	return "cache:" + r.opts.Slug + ":" + id
}

// Create persists a new record and invalidates nothing (new id).
func (r *Repo[T]) Create(ctx context.Context, m *T) error {
	return r.db.WithContext(ctx).Create(m).Error
}

// CreateTx persists a record inside the given transaction.
func (r *Repo[T]) CreateTx(ctx context.Context, tx *gorm.DB, m *T) error {
	return tx.WithContext(ctx).Create(m).Error
}

// Get loads a record by id. The distributed cache is write-through
// invalidated on mutations (see Patch/Delete); reads always hit the DB so
// they are never stale.
func (r *Repo[T]) Get(ctx context.Context, id string) (*T, error) {
	if id == "" {
		return nil, errs.ErrNotFound
	}
	var out T
	tx := r.db.WithContext(ctx)
	for _, p := range r.opts.Preloads {
		tx = tx.Preload(p)
	}
	if err := tx.First(&out, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	return &out, nil
}

// Patch applies partial updates from a whitelisted column map.
func (r *Repo[T]) Patch(ctx context.Context, id string, updates map[string]any) (*T, error) {
	if _, err := r.Get(ctx, id); err != nil {
		return nil, err
	}
	cols := make([]string, 0, len(updates))
	for k := range updates {
		cols = append(cols, k)
	}
	if len(cols) > 0 {
		if err := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Select(cols).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	_ = r.cache.Del(ctx, r.cacheKey(id))
	return r.Get(ctx, id)
}

// PatchTx applies partial updates inside a transaction.
func (r *Repo[T]) PatchTx(ctx context.Context, tx *gorm.DB, id string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	cols := make([]string, 0, len(updates))
	for k := range updates {
		cols = append(cols, k)
	}
	_ = r.cache.Del(ctx, r.cacheKey(id))
	return tx.WithContext(ctx).Model(new(T)).Where("id = ?", id).Select(cols).Updates(updates).Error
}

// Delete soft-deletes a record.
func (r *Repo[T]) Delete(ctx context.Context, id string) error {
	if _, err := r.Get(ctx, id); err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Delete(new(T), "id = ?", id).Error; err != nil {
		return err
	}
	_ = r.cache.Del(ctx, r.cacheKey(id))
	return nil
}

// Count returns the total count for a query without filters.
func (r *Repo[T]) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(new(T)).Count(&n).Error
	return n, err
}
