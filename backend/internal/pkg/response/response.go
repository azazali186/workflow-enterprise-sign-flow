// Package response builds the unified snake_case JSON API envelope.
package response

import (
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
)

// Body is the unified API response envelope.
type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// OK writes a 200 response with data.
func OK(c *app.RequestContext, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: 0, Message: "ok", Data: data})
}

// Created writes a 201 response with data.
func Created(c *app.RequestContext, data interface{}) {
	c.JSON(http.StatusCreated, Body{Code: 0, Message: "created", Data: data})
}

// Fail writes a 400 response with a safe message.
func Fail(c *app.RequestContext, message string) {
	c.JSON(http.StatusBadRequest, Body{Code: 40000, Message: message})
}

// Error writes the JSON body for a typed error.
func Error(c *app.RequestContext, err error) {
	e := errs.From(err)
	if e == nil {
		e = errs.New(http.StatusInternalServerError, 50000, "internal server error")
	}
	c.JSON(e.Status, Body{Code: e.Code, Message: e.Message})
}
