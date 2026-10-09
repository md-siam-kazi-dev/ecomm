package response

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// Envelope is the standard success response shape.
type Envelope struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data,omitempty"`
	Message    string      `json:"message,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type ErrorEnvelope struct {
	Success bool      `json:"success"`
	Error   ErrorBody `json:"error"`
}

// JSON writes a success envelope with the given status code.
func JSON(w http.ResponseWriter, status int, data interface{}, message string) {
	writeJSON(w, status, Envelope{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// Created writes a 201 response.
func Created(w http.ResponseWriter, data interface{}, message string) {
	JSON(w, http.StatusCreated, data, message)
}

// OK writes a 200 response.
func OK(w http.ResponseWriter, data interface{}, message string) {
	JSON(w, http.StatusOK, data, message)
}

// List writes a 200 response with pagination metadata. TotalPages, HasNext and
// HasPrevious are derived here so callers only supply Page/Limit/Total.
func List(w http.ResponseWriter, data interface{}, p Pagination) {
	if p.Limit > 0 {
		p.TotalPages = int((p.Total + int64(p.Limit) - 1) / int64(p.Limit))
	}
	p.HasNext = p.Page < p.TotalPages
	p.HasPrevious = p.Page > 1
	writeJSON(w, http.StatusOK, Envelope{
		Success:    true,
		Data:       data,
		Pagination: &p,
	})
}

// NoContent writes a 204 response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error maps any error to the standard error envelope. Known application
// errors keep their code/status; unknown errors become a generic 500 so that
// internal details and stack traces are never leaked to clients.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	appErr, ok := apperr.As(err)
	if !ok {
		appErr = apperr.ErrInternal.WithCause(err)
	}

	if appErr.Status >= http.StatusInternalServerError {
		slog.ErrorContext(r.Context(), "request failed",
			slog.String("error", appErr.Error()),
			slog.String("path", r.URL.Path),
		)
	}

	writeJSON(w, appErr.Status, ErrorEnvelope{
		Success: false,
		Error: ErrorBody{
			Code:    appErr.Code,
			Message: appErr.Message,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if status == http.StatusNoContent {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("failed to encode response", slog.String("error", err.Error()))
	}
}
