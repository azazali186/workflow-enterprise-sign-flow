// Package breaker wraps external calls with a circuit breaker.
package breaker

import (
	"errors"
	"time"

	"github.com/sony/gobreaker"
)

// Config tunes the circuit breaker.
type Config struct {
	Name             string
	MaxRequests      uint32
	Interval         time.Duration
	Timeout          time.Duration
	FailureThreshold uint32
}

// Defaults returns sane defaults.
func Defaults(name string) Config {
	return Config{
		Name:             name,
		MaxRequests:      5,
		Interval:         10 * time.Second,
		Timeout:          15 * time.Second,
		FailureThreshold: 5,
	}
}

// Breaker wraps gobreaker with a typed API.
type Breaker struct {
	cb *gobreaker.CircuitBreaker
}

// New builds a breaker from a config.
func New(cfg Config) *Breaker {
	return &Breaker{cb: gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        cfg.Name,
		MaxRequests: cfg.MaxRequests,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= cfg.FailureThreshold
		},
	})}
}

// ErrOpen is returned when the circuit is open.
var ErrOpen = errors.New("circuit breaker open")

// Execute runs fn guarded by the breaker. A rejected call returns ErrOpen.
func (b *Breaker) Execute(fn func() error) error {
	_, err := b.cb.Execute(func() (interface{}, error) { return nil, fn() })
	if errors.Is(err, gobreaker.ErrOpenState) {
		return ErrOpen
	}
	return err
}

// State exposes the current breaker state (for metrics).
func (b *Breaker) State() string { return b.cb.State().String() }
