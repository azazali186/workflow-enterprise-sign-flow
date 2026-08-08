// Package apikit provides shared HTTP handler helpers.
package apikit

import (
	"github.com/cloudwego/hertz/pkg/app"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/response"
)

// Bind binds and validates a JSON request body. On failure it writes a 400
// response and returns false.
func Bind[T any](c *app.RequestContext, req *T) bool {
	if err := c.BindAndValidate(req); err != nil {
		response.Error(c, errs.Wrap(400, 40000, "validation failed", err))
		return false
	}
	return true
}
