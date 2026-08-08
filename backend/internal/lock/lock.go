// Package lock provides Redis-based distributed locking helpers.
package lock

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
)

// Lock is a held distributed lock that must be released.
type Lock struct {
	key string
	val string
	c   cache.Cache
}

// Acquire tries to acquire a lock, failing fast with ErrLocked if held.
func Acquire(ctx context.Context, c cache.Cache, key string, ttl time.Duration) (*Lock, error) {
	val := uuid.NewString()
	ok, err := c.Lock(ctx, "lock:"+key, val, ttl)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errs.ErrLocked
	}
	return &Lock{key: "lock:" + key, val: val, c: c}, nil
}

// Release releases the lock (no-op safe).
func (l *Lock) Release(ctx context.Context) {
	if l == nil {
		return
	}
	_ = l.c.Unlock(ctx, l.key, l.val)
}
