// Package errs defines typed application errors mapped to HTTP status codes.
package errs

import (
	"errors"
	"net/http"
)

// Error is a typed application error with an HTTP status and a safe message.
type Error struct {
	Status  int    `json:"-"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// Unwrap exposes the wrapped cause.
func (e *Error) Unwrap() error { return e.Err }

// New builds a typed error.
func New(status, code int, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

// Wrap builds a typed error around a cause (cause is logged, never exposed).
func Wrap(status, code int, message string, err error) *Error {
	return &Error{Status: status, Code: code, Message: message, Err: err}
}

// From converts any error into a typed error, keeping HTTP status if known.
func From(err error) *Error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	if errors.Is(err, ErrNotFound) {
		return ErrNotFound
	}
	if errors.Is(err, ErrConflict) {
		return ErrConflict
	}
	if errors.Is(err, ErrUnauthorized) {
		return ErrUnauthorized
	}
	return Wrap(http.StatusInternalServerError, 50000, "internal server error", err)
}

// Predefined errors used across the application.
var (
	ErrNotFound      = New(http.StatusNotFound, 40400, "resource not found")
	ErrConflict      = New(http.StatusConflict, 40900, "resource already exists")
	ErrUnauthorized  = New(http.StatusUnauthorized, 40100, "unauthorized")
	ErrForbidden     = New(http.StatusForbidden, 40300, "forbidden")
	ErrValidation    = New(http.StatusBadRequest, 40000, "validation failed")
	ErrBadRequest    = New(http.StatusBadRequest, 40001, "bad request")
	ErrRateLimited   = New(http.StatusTooManyRequests, 42900, "too many requests")
	ErrInvalidCursor = New(http.StatusBadRequest, 40002, "invalid pagination cursor")
	ErrLocked        = New(http.StatusConflict, 40901, "resource is locked by another operation")
)

// Is reports whether err matches the target error value.
func Is(err, target error) bool { return errors.Is(err, target) }
