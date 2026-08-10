package middleware

import (
	"context"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/stretchr/testify/assert"
)

func newSecurityCtx(method string) *app.RequestContext {
	c := app.NewContext(0)
	c.Request.SetMethod(method)
	return c
}

func TestSecuritySetsCorsForAllowedOrigin(t *testing.T) {
	mw := Security([]string{"https://app.example.com"}, false)
	c := newSecurityCtx("POST")
	c.Request.Header.Set("Origin", "https://app.example.com")

	mw(context.Background(), c)

	assert.Equal(t, "https://app.example.com", string(c.Response.Header.Peek("Access-Control-Allow-Origin")))
	assert.Equal(t, "POST, PATCH, DELETE, OPTIONS", string(c.Response.Header.Peek("Access-Control-Allow-Methods")))
	assert.Equal(t, "Origin", string(c.Response.Header.Peek("Vary")))
}

func TestSecurityOmitsCorsForUnknownOrigin(t *testing.T) {
	mw := Security([]string{"https://app.example.com"}, false)
	c := newSecurityCtx("POST")
	c.Request.Header.Set("Origin", "https://evil.example.com")

	mw(context.Background(), c)

	assert.Empty(t, string(c.Response.Header.Peek("Access-Control-Allow-Origin")))
}

func TestSecurityAllowAllOrigin(t *testing.T) {
	mw := Security([]string{"*"}, false)
	c := newSecurityCtx("POST")
	c.Request.Header.Set("Origin", "https://anything.example.com")

	mw(context.Background(), c)

	assert.Equal(t, "*", string(c.Response.Header.Peek("Access-Control-Allow-Origin")))
}

func TestSecurityPreflightShortCircuits(t *testing.T) {
	mw := Security([]string{"https://app.example.com"}, false)
	c := newSecurityCtx("OPTIONS")
	c.Request.Header.Set("Origin", "https://app.example.com")
	called := false
	c.SetHandlers([]app.HandlerFunc{
		mw,
		func(ctx context.Context, _ *app.RequestContext) { called = true },
	})

	c.Next(context.Background())

	assert.False(t, called, "preflight must not continue the chain")
	assert.Equal(t, 204, c.Response.StatusCode())
}

func TestSecuritySetsHardeningHeaders(t *testing.T) {
	mw := Security(nil, true) // production
	c := newSecurityCtx("POST")

	mw(context.Background(), c)

	assert.Equal(t, "nosniff", string(c.Response.Header.Peek("X-Content-Type-Options")))
	assert.Equal(t, "DENY", string(c.Response.Header.Peek("X-Frame-Options")))
	assert.Equal(t, "no-referrer", string(c.Response.Header.Peek("Referrer-Policy")))
	assert.Equal(t, "max-age=63072000; includeSubDomains", string(c.Response.Header.Peek("Strict-Transport-Security")))
}

func TestSecurityNoHstsOutsideProduction(t *testing.T) {
	mw := Security(nil, false) // development
	c := newSecurityCtx("POST")

	mw(context.Background(), c)

	assert.Empty(t, string(c.Response.Header.Peek("Strict-Transport-Security")))
}
