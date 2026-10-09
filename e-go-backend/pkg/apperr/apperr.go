package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

// Error is a centralized application error carrying a stable machine-readable
// code and the HTTP status it maps to. Handlers return these; the response
// layer translates them into the PRD error envelope.
type Error struct {
	Status  int
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// WithCause returns a copy of the error wrapping an underlying cause. The
// underlying cause is logged but never exposed to clients.
func (e *Error) WithCause(err error) *Error {
	clone := *e
	clone.Err = err
	return &clone
}

func New(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

// As extracts an *Error from an error chain, if present.
func As(err error) (*Error, bool) {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// Predefined errors used across modules.
var (
	ErrInternal          = New(http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong")
	ErrUnauthorized      = New(http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
	ErrForbidden         = New(http.StatusForbidden, "FORBIDDEN", "You do not have permission to perform this action")
	ErrNotFound          = New(http.StatusNotFound, "NOT_FOUND", "Resource not found")
	ErrConflict          = New(http.StatusConflict, "CONFLICT", "Resource already exists")
	ErrValidation        = New(http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Validation failed")
	ErrRateLimited       = New(http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests")
	ErrInsufficientStock = New(http.StatusConflict, "INSUFFICIENT_STOCK", "Insufficient stock")
)
