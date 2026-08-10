// Package cache provides the distributed Redis cache with an in-memory
// fallback for tests and local development.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache is the distributed cache interface used across the application.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	GetWithTTL(ctx context.Context, key string) (string, time.Duration, error)
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Incr(ctx context.Context, key string, ttl time.Duration) (int64, error)
	Lock(ctx context.Context, key string, val string, ttl time.Duration) (bool, error)
	Unlock(ctx context.Context, key, val string) error
	// ScanAndDel deletes every key matching a glob pattern (used to purge
	// cache families such as the per-user RBAC grants on role changes).
	ScanAndDel(ctx context.Context, pattern string) error
	// Ping reports whether the backing store is reachable (readiness probes).
	Ping(ctx context.Context) error
}

// Redis implements Cache backed by Redis (also used for distributed locks).
type Redis struct {
	client *redis.Client
}

// NewRedis builds a Redis-backed cache.
func NewRedis(url string) (*Redis, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &Redis{client: client}, nil
}

// NewRedisClient wraps an existing client (used for tests).
func NewRedisClient(client *redis.Client) *Redis { return &Redis{client: client} }

// Ping verifies connectivity to the Redis server.
func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *Redis) GetWithTTL(ctx context.Context, key string) (string, time.Duration, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", 0, err
	}
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return "", 0, err
	}
	return val, ttl, nil
}

func (r *Redis) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	var payload string
	switch v := val.(type) {
	case string:
		payload = v
	case []byte:
		payload = string(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		payload = string(b)
	}
	return r.client.Set(ctx, key, payload, ttl).Err()
}

func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// ScanAndDel purges all keys matching pattern via SCAN + DEL, deleting in
// batches so huge key families are not held in memory. SCAN is non-blocking
// and safe to run against a live store.
func (r *Redis) ScanAndDel(ctx context.Context, pattern string) error {
	const batch = 1000
	var keys []string
	iter := r.client.Scan(ctx, 0, pattern, batch).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= batch {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
			keys = keys[:0]
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}
	return nil
}

func (r *Redis) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	n, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if n == 1 {
		r.client.Expire(ctx, key, ttl)
	}
	return n, nil
}

// Lock acquires a non-blocking distributed lock (SET NX PX).
func (r *Redis) Lock(ctx context.Context, key, val string, ttl time.Duration) (bool, error) {
	ok, err := r.client.SetNX(ctx, key, val, ttl).Result()
	return ok, err
}

// Unlock releases the lock only if val matches (prevents releasing others' locks).
func (r *Redis) Unlock(ctx context.Context, key, val string) error {
	script := `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`
	return r.client.Eval(ctx, script, []string{key}, val).Err()
}

// Memory is an in-memory Cache used by unit/integration tests.
type Memory struct {
	mu    sync.Mutex
	store map[string]item
}

type item struct {
	val string
	exp time.Time
}

// NewMemory builds an empty in-memory cache.
func NewMemory() *Memory { return &Memory{store: map[string]item{}} }

// Ping always succeeds for the in-memory store.
func (m *Memory) Ping(ctx context.Context) error { return nil }

func (m *Memory) Get(ctx context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.store[key]
	if !ok || expired(it.exp) {
		return "", errors.New("redis: nil")
	}
	return it.val, nil
}

func (m *Memory) GetWithTTL(ctx context.Context, key string) (string, time.Duration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.store[key]
	if !ok || expired(it.exp) {
		return "", 0, errors.New("redis: nil")
	}
	ttl := time.Duration(0)
	if !it.exp.IsZero() {
		ttl = time.Until(it.exp)
	}
	return it.val, ttl, nil
}

func (m *Memory) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Mirror Redis semantics: ttl <= 0 persists indefinitely. GetWithTTL then
	// reports 0, which callers only use to decide session renewal — session
	// keys are always created with a positive TTL.
	exp := time.Time{}
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	m.store[key] = item{val: asStr(val), exp: exp}
	return nil
}

// expired reports whether an item has passed its expiry (zero means forever).
func expired(exp time.Time) bool {
	return !exp.IsZero() && time.Now().After(exp)
}

func (m *Memory) Del(ctx context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, k := range keys {
		delete(m.store, k)
	}
	return nil
}

// ScanAndDel deletes every key matching a glob-style pattern. Note: path.Match
// treats '*' as not crossing '/', unlike Redis glob; patterns used in practice
// ("rbac:user:*") contain no slashes, so the two implementations agree.
func (m *Memory) ScanAndDel(ctx context.Context, pattern string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k := range m.store {
		if ok, err := path.Match(pattern, k); err == nil && ok {
			delete(m.store, k)
		}
	}
	return nil
}

func (m *Memory) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.store[key]
	n := int64(0)
	if ok && !expired(it.exp) {
		_ = json.Unmarshal([]byte(it.val), &n)
	}
	n++
	exp := time.Time{}
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	m.store[key] = item{val: itoa(n), exp: exp}
	return n, nil
}

func (m *Memory) Lock(ctx context.Context, key, val string, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.store[key]
	if ok && !expired(it.exp) {
		return false, nil
	}
	exp := time.Time{}
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	m.store[key] = item{val: val, exp: exp}
	return true, nil
}

func (m *Memory) Unlock(ctx context.Context, key, val string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if it, ok := m.store[key]; ok && it.val == val {
		delete(m.store, key)
	}
	return nil
}

func asStr(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}
