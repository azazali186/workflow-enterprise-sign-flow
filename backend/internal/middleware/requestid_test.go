package middleware

import (
	"context"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/stretchr/testify/assert"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/ctxval"
)

func TestRequestIDGeneratesForEveryRequest(t *testing.T) {
	mw := RequestID()
	c := app.NewContext(0)
	c.Request.SetMethod("POST")

	mw(context.Background(), c)

	assert.NotEmpty(t, string(c.Response.Header.Peek("X-Request-ID")))
}

func TestRequestIDHonoursClientHeader(t *testing.T) {
	mw := RequestID()
	c := app.NewContext(0)
	c.Request.SetMethod("POST")
	c.Request.Header.Set("X-Request-ID", "client-supplied-id")

	ctx := context.Background()
	mw(ctx, c)

	assert.Equal(t, "client-supplied-id", string(c.Response.Header.Peek("X-Request-ID")))
}

func TestRequestIDChainsToHandlers(t *testing.T) {
	mw := RequestID()
	c := app.NewContext(0)
	c.Request.SetMethod("POST")

	var captured string
	c.SetHandlers([]app.HandlerFunc{
		mw,
		func(ctx context.Context, _ *app.RequestContext) { captured = ctxval.RequestID(ctx) },
	})

	c.Next(context.Background())

	assert.NotEmpty(t, captured)
	assert.Equal(t, captured, string(c.Response.Header.Peek("X-Request-ID")))
}
