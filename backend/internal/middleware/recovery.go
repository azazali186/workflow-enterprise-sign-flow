package middleware

import (
	"context"
	"runtime/debug"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/logger"
)

// Recovery converts panics into JSON 500 responses. Stack traces go to the
// logger only — never to the client.
func Recovery() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				logger.L().Error("panic recovered",
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())),
					zap.String("path", string(c.Request.URI().Path())),
				)
				c.JSON(500, body(50000, "internal server error"))
				c.Abort()
			}
		}()
		c.Next(ctx)
	}
}
