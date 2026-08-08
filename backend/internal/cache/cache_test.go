package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryCache(t *testing.T) {
	ctx := context.Background()
	c := NewMemory()

	_, err := c.Get(ctx, "missing")
	assert.Error(t, err)

	require.NoError(t, c.Set(ctx, "k", "v", time.Minute))
	v, err := c.Get(ctx, "k")
	require.NoError(t, err)
	assert.Equal(t, "v", v)

	require.NoError(t, c.Del(ctx, "k"))
	_, err = c.Get(ctx, "k")
	assert.Error(t, err)
}

func TestMemoryExpiry(t *testing.T) {
	ctx := context.Background()
	c := NewMemory()
	require.NoError(t, c.Set(ctx, "k", "v", 10*time.Millisecond))
	time.Sleep(30 * time.Millisecond)
	_, err := c.Get(ctx, "k")
	assert.Error(t, err)
}

func TestMemoryIncr(t *testing.T) {
	ctx := context.Background()
	c := NewMemory()
	n1, err := c.Incr(ctx, "counter", time.Minute)
	require.NoError(t, err)
	n2, err := c.Incr(ctx, "counter", time.Minute)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n1)
	assert.Equal(t, int64(2), n2)
}

func TestMemoryLock(t *testing.T) {
	ctx := context.Background()
	c := NewMemory()
	ok, err := c.Lock(ctx, "resource", "owner-1", time.Minute)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = c.Lock(ctx, "resource", "owner-2", time.Minute)
	require.NoError(t, err)
	assert.False(t, ok)

	require.NoError(t, c.Unlock(ctx, "resource", "owner-2")) // wrong owner
	ok, _ = c.Lock(ctx, "resource", "owner-2", time.Minute)
	assert.False(t, ok) // still held by owner-1

	require.NoError(t, c.Unlock(ctx, "resource", "owner-1")) // correct owner
	ok, _ = c.Lock(ctx, "resource", "owner-2", time.Minute)
	assert.True(t, ok)
}
