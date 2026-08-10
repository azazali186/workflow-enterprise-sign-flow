package retryutil

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errFlaky = errors.New("flaky")

func TestDoRetriesUntilSuccess(t *testing.T) {
	attempts := 0
	cfg := Config{InitialInterval: time.Millisecond, MaxInterval: 5 * time.Millisecond, MaxElapsedTime: 2 * time.Second}
	err := Do(context.Background(), cfg, func() error {
		attempts++
		if attempts < 3 {
			return errFlaky
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestDoGivesUpAfterMaxElapsed(t *testing.T) {
	attempts := 0
	cfg := Config{InitialInterval: time.Millisecond, MaxInterval: time.Millisecond, MaxElapsedTime: 30 * time.Millisecond}
	err := Do(context.Background(), cfg, func() error {
		attempts++
		return errFlaky
	})
	assert.ErrorIs(t, err, errFlaky)
	assert.GreaterOrEqual(t, attempts, 2)
}

func TestDoRespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cfg := Config{InitialInterval: time.Millisecond, MaxInterval: time.Millisecond, MaxElapsedTime: time.Minute}
	err := Do(ctx, cfg, func() error { return errFlaky })
	assert.Error(t, err)
}

func TestNotifyInvokesCallback(t *testing.T) {
	notified := 0
	cfg := Config{InitialInterval: time.Millisecond, MaxInterval: time.Millisecond, MaxElapsedTime: 30 * time.Millisecond}
	_ = Notify(context.Background(), cfg, func() error { return errFlaky },
		func(err error, d time.Duration) { notified++ })
	assert.GreaterOrEqual(t, notified, 1)
}
