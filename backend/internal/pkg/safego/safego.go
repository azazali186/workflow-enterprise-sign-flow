// Package safego runs background goroutines with panic containment. An
// unrecovered panic in any goroutine would crash the whole process, so every
// worker (outbox relay, event bus, NATS reconnect, WS pumps) must go through
// Go. The panic is logged with a stack trace; the caller stays alive.
package safego

import (
	"runtime/debug"

	"go.uber.org/zap"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/logger"
)

// Go runs fn in a new goroutine, recovering panics so a transient bug in one
// worker cannot take down the process.
func Go(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.L().Error("panic in background goroutine",
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())),
				)
			}
		}()
		fn()
	}()
}
