// Package natsx wraps NATS JetStream with graceful degradation: if NATS is
// unreachable the application keeps running and events are delivered later.
package natsx

import (
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/logger"
)

const (
	// Stream is the JetStream stream holding all domain events.
	Stream = "SIGNFLOW"
	// SubjectPrefix is the subject prefix for domain events.
	SubjectPrefix = "signflow.events"
)

// Client wraps the NATS connection and JetStream context.
type Client struct {
	Conn *nats.Conn
	JS   nats.JetStreamContext
}

// Connect dials NATS and ensures the stream exists. Returns nil on failure —
// callers should retry with backoff rather than crash.
func Connect(url string) (*Client, error) {
	nc, err := nats.Connect(url,
		nats.Timeout(3*time.Second),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		logger.L().Warn("nats connect failed, event delivery deferred", zap.Error(err))
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}
	if _, err := js.AddStream(&nats.StreamConfig{
		Name:     Stream,
		Subjects: []string{SubjectPrefix + ".>"},
		Storage:  nats.FileStorage,
	}); err != nil && err != nats.ErrStreamNameAlreadyInUse {
		logger.L().Warn("nats stream setup failed", zap.Error(err))
	}
	return &Client{Conn: nc, JS: js}, nil
}
