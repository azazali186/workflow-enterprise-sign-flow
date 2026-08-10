package middleware

import (
	"context"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/stretchr/testify/assert"

	"github.com/aeroxe/sign-flow/backend/internal/cache"
)

// runRateLimit invokes the middleware through the handler chain for a path
// and returns the resulting status code.
func runRateLimit(c cache.Cache, budget int, routes map[string]int, path string) int {
	ctx := app.NewContext(0)
	ctx.Request.SetMethod("POST")
	ctx.SetFullPath(path)
	status := 200
	ctx.SetHandlers([]app.HandlerFunc{
		RateLimit(c, budget, routes),
		func(_ context.Context, c *app.RequestContext) { c.SetStatusCode(200) },
	})
	ctx.Next(context.Background())
	status = ctx.Response.StatusCode()
	return status
}

func TestRateLimitGlobalBudget(t *testing.T) {
	c := cache.NewMemory()
	status := 200
	for i := 0; i < 3; i++ {
		status = runRateLimit(c, 2, nil, "/api/v1/contracts/list")
	}
	assert.Equal(t, 429, status, "third request exceeds the global budget of 2")
}

func TestRateLimitRouteSpecificBudget(t *testing.T) {
	c := cache.NewMemory()
	routes := map[string]int{"/api/v1/auth/login": 1}

	// Login route: its own budget of 1 -> second hit is rejected.
	status := runRateLimit(c, 100, routes, "/api/v1/auth/login")
	assert.Equal(t, 200, status)
	status = runRateLimit(c, 100, routes, "/api/v1/auth/login")
	assert.Equal(t, 429, status, "login route budget of 1 must be enforced independently")

	// Other routes keep the global budget, unaffected by the login hits.
	status = runRateLimit(c, 100, routes, "/api/v1/contracts/list")
	assert.Equal(t, 200, status)
	status = runRateLimit(c, 100, routes, "/api/v1/contracts/list")
	assert.Equal(t, 200, status)
}
