package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/ctxval"
)

// RequestID assigns every request a correlation id (honouring a client-supplied
// X-Request-ID) before any handler runs, so public routes and audit entries
// carry one too. Must be registered before RequestLog.
func RequestID() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		ctx = ctxval.SetRequestID(ctx, requestID(c))
		c.Response.Header.Set("X-Request-ID", ctxval.RequestID(ctx))
		c.Next(ctx)
	}
}
