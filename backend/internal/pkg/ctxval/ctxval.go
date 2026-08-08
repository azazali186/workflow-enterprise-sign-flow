// Package ctxval defines request-scoped context values shared by middleware,
// audit logging and services.
package ctxval

import (
	"context"
)

type key int

const (
	keyUserID key = iota
	keyUserName
	keyIP
	keyUserAgent
	keyRequestID
	keyToken
)

// SetUserID stores the authenticated user id.
func SetUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyUserID, id)
}

// UserID returns the authenticated user id (empty if anonymous).
func UserID(ctx context.Context) string {
	if v, ok := ctx.Value(keyUserID).(string); ok {
		return v
	}
	return ""
}

// SetUserName stores the actor display name.
func SetUserName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, keyUserName, name)
}

// UserName returns the actor name.
func UserName(ctx context.Context) string {
	if v, ok := ctx.Value(keyUserName).(string); ok {
		return v
	}
	return ""
}

// SetIP stores the client ip.
func SetIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, keyIP, ip)
}

// IP returns the client ip.
func IP(ctx context.Context) string {
	if v, ok := ctx.Value(keyIP).(string); ok {
		return v
	}
	return ""
}

// SetUserAgent stores the client user agent.
func SetUserAgent(ctx context.Context, ua string) context.Context {
	return context.WithValue(ctx, keyUserAgent, ua)
}

// UserAgent returns the client user agent.
func UserAgent(ctx context.Context) string {
	if v, ok := ctx.Value(keyUserAgent).(string); ok {
		return v
	}
	return ""
}

// SetRequestID stores the request id.
func SetRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyRequestID, id)
}

// RequestID returns the request id.
func RequestID(ctx context.Context) string {
	if v, ok := ctx.Value(keyRequestID).(string); ok {
		return v
	}
	return ""
}
