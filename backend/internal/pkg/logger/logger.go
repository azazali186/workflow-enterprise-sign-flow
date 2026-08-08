// Package logger provides the shared zap logger used across the application.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Init configures the global logger. Call once at startup.
func Init(env, level string) {
	cfg := zap.NewProductionConfig()
	if env != "production" {
		cfg = zap.NewDevelopmentConfig()
	}
	if lvl, err := zapcore.ParseLevel(level); err == nil {
		cfg.Level = zap.NewAtomicLevelAt(lvl)
	}
	log = zap.Must(cfg.Build())
}

// L returns the global logger (no-op fallback if not initialised).
func L() *zap.Logger {
	if log == nil {
		log = zap.NewNop()
	}
	return log
}

// Sync flushes buffered log entries.
func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}
