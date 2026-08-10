package middleware

import (
	"context"
	"slices"

	"github.com/cloudwego/hertz/pkg/app"
)

// Security sets CORS and hardening response headers on every response.
// allowedOrigins is a comma-separated allow-list; empty means same-origin
// only (no CORS headers emitted), "*" allows any origin (development only).
func Security(allowedOrigins []string, production bool) app.HandlerFunc {
	allowAll := slices.Contains(allowedOrigins, "*")
	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.Request.Header.Peek("Origin"))
		if origin != "" && (allowAll || slices.Contains(allowedOrigins, origin)) {
			h := &c.Response.Header
			if allowAll {
				h.Set("Access-Control-Allow-Origin", "*")
			} else {
				h.Set("Access-Control-Allow-Origin", origin)
				h.Set("Vary", "Origin")
			}
			h.Set("Access-Control-Allow-Methods", "POST, PATCH, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			h.Set("Access-Control-Max-Age", "86400")
		}
		// Hardening headers (always applied, even on error responses).
		c.Response.Header.Set("X-Content-Type-Options", "nosniff")
		c.Response.Header.Set("X-Frame-Options", "DENY")
		c.Response.Header.Set("Referrer-Policy", "no-referrer")
		if production {
			c.Response.Header.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		}
		// CORS preflight: answer immediately, never reach auth/RBAC.
		if string(c.Request.Method()) == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next(ctx)
	}
}
