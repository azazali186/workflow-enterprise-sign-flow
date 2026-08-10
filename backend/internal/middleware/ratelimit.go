package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/aeroxe/sign-flow/backend/internal/cache"
)

// RateLimit enforces a per-minute budget per client IP using Redis. A global
// budget applies to every route; routeLimits may impose a stricter (or
// looser) budget per exact path — e.g. auth/login gets a lower ceiling so a
// distributed attack cannot hammer credentials from one IP.
func RateLimit(cache cache.Cache, perMinute int, routeLimits map[string]int) app.HandlerFunc {
	if perMinute <= 0 {
		perMinute = 120
	}
	return func(ctx context.Context, c *app.RequestContext) {
		if perMinute >= 1<<20 { // disabled when configured very high
			c.Next(ctx)
			return
		}
		ip := c.ClientIP()
		if ip == "" {
			ip = "unknown"
		}
		// Separate key per route that has a specific budget, so a stricter
		// login limit does not consume (or get consumed by) the global one.
		key := "ratelimit:" + ip
		budget := perMinute
		if rl, ok := routeLimits[c.FullPath()]; ok {
			key += ":" + c.FullPath()
			budget = rl
		}
		n, err := cache.Incr(ctx, key, time.Minute)
		if err == nil && n > int64(budget) {
			c.JSON(429, body(42900, "too many requests"))
			c.Abort()
			return
		}
		c.Next(ctx)
	}
}
