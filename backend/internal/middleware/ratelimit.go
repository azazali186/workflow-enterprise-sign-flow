package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/aeroxe/sign-flow/backend/internal/cache"
)

// RateLimit enforces a per-minute budget per client IP using Redis.
func RateLimit(cache cache.Cache, perMinute int) app.HandlerFunc {
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
		n, err := cache.Incr(ctx, "ratelimit:"+ip, time.Minute)
		if err == nil && n > int64(perMinute) {
			c.JSON(429, body(42900, "too many requests"))
			c.Abort()
			return
		}
		c.Next(ctx)
	}
}
