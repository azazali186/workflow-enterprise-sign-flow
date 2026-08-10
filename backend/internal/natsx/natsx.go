// Package natsx wraps NATS JetStream with graceful degradation: if NATS is
// unreachable the client keeps reconnecting in the background and Publish
// returns an error so the outbox relay retries instead of losing events.
package natsx

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/logger"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/safego"
)

const (
	// Stream is the JetStream stream holding all domain events.
	Stream = "SIGNFLOW"
	// SubjectPrefix is the subject prefix for domain events.
	SubjectPrefix = "signflow.events"
)

// ErrNotConnected is returned by Publish while NATS is unavailable.
var ErrNotConnected = errors.New("nats: not connected")

// Client wraps the NATS connection and JetStream context. The connection is
// swapped atomically by the reconnect loop, so Publish is safe from any goroutine.
type Client struct {
	mu   sync.RWMutex
	conn *nats.Conn
	js   nats.JetStreamContext
}

// NewClient returns a disconnected client; call Connect or
// StartReconnectLoop to establish a connection.
func NewClient() *Client { return &Client{} }

// Connect dials NATS, creates the stream and stores the connection.
func Connect(url string) (*Client, error) {
	c := &Client{}
	if err := c.connect(url); err != nil {
		return nil, err
	}
	return c, nil
}

// Connect establishes a connection on an existing (possibly disconnected) client.
func (c *Client) Connect(url string) error { return c.connect(url) }

func (c *Client) connect(url string) error {
	nc, err := nats.Connect(url,
		nats.Timeout(3*time.Second),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return err
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return err
	}
	if _, err := js.AddStream(&nats.StreamConfig{
		Name:     Stream,
		Subjects: []string{SubjectPrefix + ".>"},
		Storage:  nats.FileStorage,
		// Bound retention: keep ~7 days or up to 1 GiB so the JetStream
		// store cannot grow without limit.
		MaxAge:   7 * 24 * time.Hour,
		MaxBytes: 1 << 30,
	}); err != nil && !errors.Is(err, nats.ErrStreamNameAlreadyInUse) {
		logger.L().Warn("nats stream setup failed", zap.Error(err))
	}
	c.mu.Lock()
	c.conn = nc
	c.js = js
	c.mu.Unlock()
	logger.L().Info("nats connected")
	return nil
}

// Connected reports whether a live connection is held.
func (c *Client) Connected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil && c.conn.IsConnected()
}

// Publish sends a payload to a subject; returns ErrNotConnected while the
// reconnect loop has not yet established a connection.
func (c *Client) Publish(subject string, payload []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.js == nil {
		return ErrNotConnected
	}
	_, err := c.js.Publish(subject, payload)
	return err
}

// Close drains and closes the connection if any.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		_ = c.conn.Drain()
		c.conn = nil
		c.js = nil
	}
}

// StartReconnectLoop keeps dialing until connected, backing off between
// attempts, and never exits after a success: if the connection is later lost
// (and the nats library cannot restore it) this loop re-dials. Stops when
// ctx is cancelled. Safe to call once at boot.
func (c *Client) StartReconnectLoop(ctx context.Context, url string, interval time.Duration) {
	safego.Go(func() {
		if interval <= 0 {
			interval = 5 * time.Second
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			if !c.Connected() {
				if err := c.connect(url); err == nil {
					logger.L().Info("nats reconnected")
				}
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	})
}
