package breaker

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errBoom = errors.New("boom")

func TestExecuteSuccess(t *testing.T) {
	b := New(Defaults("test"))
	require.NoError(t, b.Execute(func() error { return nil }))
}

func TestBreakerOpensAfterFailures(t *testing.T) {
	b := New(Config{Name: "t", MaxRequests: 1, Interval: time.Minute, Timeout: time.Minute, FailureThreshold: 2})
	for i := 0; i < 2; i++ {
		_ = b.Execute(func() error { return errBoom })
	}
	assert.Equal(t, "open", b.State())

	// While open, the call is rejected without invoking fn.
	called := false
	err := b.Execute(func() error { called = true; return nil })
	assert.ErrorIs(t, err, ErrOpen)
	assert.False(t, called, "function must not run while the circuit is open")
}

func TestBreakerHalfOpenAllowsTrial(t *testing.T) {
	b := New(Config{Name: "t", MaxRequests: 1, Interval: time.Minute, Timeout: 80 * time.Millisecond, FailureThreshold: 1})
	for i := 0; i < 2; i++ {
		_ = b.Execute(func() error { return errBoom })
	}
	assert.Equal(t, "open", b.State())
	// Wait well past the timeout so the breaker is guaranteed half-open.
	time.Sleep(250 * time.Millisecond)

	require.NoError(t, b.Execute(func() error { return nil }))
	assert.Equal(t, "closed", b.State())
}
