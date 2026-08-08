package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/ctxval"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/logger"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/metrics"
)

// RequestLog logs method, path, status, duration and request id. Bodies and
// headers are never logged (sensitive data policy).
func RequestLog(met *metrics.Collectors) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.Request.URI().Path())
		method := string(c.Request.Method())
		c.Next(ctx)
		status := c.Response.StatusCode()
		met.HTTPRequests.WithLabelValues(method, path, itoaStatus(status)).Inc()
		met.HTTPDuration.WithLabelValues(method, path).Observe(time.Since(start).Seconds())
		logger.L().Info("http",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("duration", time.Since(start)),
			zap.String("request_id", ctxval.RequestID(ctx)),
			zap.String("ip", ctxval.IP(ctx)),
			zap.String("user_id", ctxval.UserID(ctx)),
		)
	}
}

func itoaStatus(s int) string {
	switch s {
	case 200:
		return "200"
	case 201:
		return "201"
	case 400:
		return "400"
	case 401:
		return "401"
	case 403:
		return "403"
	case 404:
		return "404"
	case 409:
		return "409"
	case 429:
		return "429"
	case 500:
		return "500"
	default:
		return "other"
	}
}
