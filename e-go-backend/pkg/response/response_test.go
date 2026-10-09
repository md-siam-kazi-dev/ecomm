package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

func TestJSON_SuccessEnvelope(t *testing.T) {
	w := httptest.NewRecorder()

	JSON(w, http.StatusOK, map[string]string{"id": "abc"}, "ok")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("unexpected content-type %q", ct)
	}

	var body Envelope
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if !body.Success {
		t.Error("expected success=true")
	}
	if body.Message != "ok" {
		t.Errorf("expected message ok, got %q", body.Message)
	}
}

func TestError_MapsKnownAppError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/products/1", nil)

	Error(w, r, apperr.ErrNotFound)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}

	var body ErrorEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body.Success {
		t.Error("expected success=false")
	}
	if body.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND code, got %q", body.Error.Code)
	}
}

func TestError_HidesUnknownError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)

	Error(w, r, errBoom)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	var body ErrorEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR, got %q", body.Error.Code)
	}
	if body.Error.Message == errBoom.Error() {
		t.Error("internal error details must not be exposed to the client")
	}
}

func TestList_SetsPagination(t *testing.T) {
	w := httptest.NewRecorder()

	List(w, []string{"a", "b"}, Pagination{Page: 1, Limit: 2, Total: 5})

	var body Envelope
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body.Pagination == nil {
		t.Fatal("expected pagination metadata")
	}
	if body.Pagination.TotalPages != 3 {
		t.Errorf("expected 3 total pages, got %d", body.Pagination.TotalPages)
	}
}

type boomError struct{}

func (boomError) Error() string { return "secret database detail" }

var errBoom = boomError{}
