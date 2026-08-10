package lock

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
)

func TestAcquireAndRelease(t *testing.T) {
	c := cache.NewMemory()
	ctx := context.Background()

	l, err := Acquire(ctx, c, "job:sign", time.Second)
	require.NoError(t, err)
	require.NotNil(t, l)

	// A second acquire of the same resource must fail fast.
	_, err = Acquire(ctx, c, "job:sign", time.Second)
	assert.ErrorIs(t, err, errs.ErrLocked)

	l.Release(ctx)
	l2, err := Acquire(ctx, c, "job:sign", time.Second)
	require.NoError(t, err)
	l2.Release(ctx)
}

func TestReleaseIsNoopForNil(t *testing.T) {
	var l *Lock
	require.NotPanics(t, func() { l.Release(context.Background()) })
}

func TestDifferentResourcesDoNotBlock(t *testing.T) {
	c := cache.NewMemory()
	ctx := context.Background()

	l1, err := Acquire(ctx, c, "job:a", time.Second)
	require.NoError(t, err)
	defer l1.Release(ctx)

	l2, err := Acquire(ctx, c, "job:b", time.Second)
	require.NoError(t, err)
	defer l2.Release(ctx)
}
