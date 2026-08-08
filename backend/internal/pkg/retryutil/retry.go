// Package retryutil provides retry with exponential backoff and jitter.
package retryutil

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// Config tunes backoff behaviour.
type Config struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	MaxElapsedTime  time.Duration
}

// Defaults returns sensible retry defaults.
func Defaults() Config {
	return Config{
		InitialInterval: 200 * time.Millisecond,
		MaxInterval:     5 * time.Second,
		MaxElapsedTime:  30 * time.Second,
	}
}

// Do retries fn with exponential backoff until it succeeds or the context
// is done / max elapsed time passes.
func Do(ctx context.Context, cfg Config, fn func() error) error {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = cfg.InitialInterval
	b.MaxInterval = cfg.MaxInterval
	b.MaxElapsedTime = cfg.MaxElapsedTime
	b.RandomizationFactor = 0.3
	return backoff.Retry(func() error { return fn() }, backoff.WithContext(b, ctx))
}

// Notify retries fn and logs each retry attempt via the notify callback.
func Notify(ctx context.Context, cfg Config, fn func() error, notify func(err error, d time.Duration)) error {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = cfg.InitialInterval
	b.MaxInterval = cfg.MaxInterval
	b.MaxElapsedTime = cfg.MaxElapsedTime
	b.RandomizationFactor = 0.3
	return backoff.RetryNotify(fn, backoff.WithContext(b, ctx), notify)
}
