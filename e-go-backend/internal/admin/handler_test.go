package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

type listResponse struct {
	Success    bool       `json:"success"`
	Data       []UserItem `json:"data"`
	Pagination struct {
		Page        int   `json:"page"`
		Limit       int   `json:"limit"`
		Total       int64 `json:"total"`
		TotalPages  int   `json:"total_pages"`
		HasNext     bool  `json:"has_next"`
		HasPrevious bool  `json:"has_previous"`
	} `json:"pagination"`
}

func doList(t *testing.T, repo *fakeUserRepo, target string) listResponse {
	t.Helper()

	r := chi.NewRouter()
	r.Get("/users", NewHandler(NewService(repo)).ListUsers)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body listResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	return body
}

func TestListUsers_ReturnsPaginationMetadata(t *testing.T) {
	repo := &fakeUserRepo{
		records: []UserRecord{sampleRecord("u-1", "a@x.com", "Alpha", []string{"user"})},
		total:   25,
	}

	body := doList(t, repo, "/users?page=2&limit=10&role=customer&search=siam")

	if !body.Success {
		t.Error("expected success=true")
	}
	if body.Pagination.Page != 2 || body.Pagination.Limit != 10 {
		t.Errorf("page/limit = %d/%d, want 2/10", body.Pagination.Page, body.Pagination.Limit)
	}
	if body.Pagination.Total != 25 || body.Pagination.TotalPages != 3 {
		t.Errorf("total/totalPages = %d/%d, want 25/3", body.Pagination.Total, body.Pagination.TotalPages)
	}
	if !body.Pagination.HasNext || !body.Pagination.HasPrevious {
		t.Errorf("hasNext/hasPrevious = %v/%v, want true/true", body.Pagination.HasNext, body.Pagination.HasPrevious)
	}

	// The role alias and search reach the repository normalized.
	if repo.lastQuery.Role != "user" || repo.lastQuery.Search != "siam" {
		t.Errorf("repo got role=%q search=%q, want user/siam", repo.lastQuery.Role, repo.lastQuery.Search)
	}
}

func TestListUsers_LastPageHasNoNext(t *testing.T) {
	repo := &fakeUserRepo{total: 25}
	body := doList(t, repo, "/users?page=3&limit=10")

	if body.Pagination.HasNext {
		t.Error("last page must not report a next page")
	}
	if !body.Pagination.HasPrevious {
		t.Error("last page must report a previous page")
	}
}

func TestListUsers_EmptyResult(t *testing.T) {
	repo := &fakeUserRepo{total: 0}
	body := doList(t, repo, "/users")

	if len(body.Data) != 0 {
		t.Errorf("expected no users, got %d", len(body.Data))
	}
	if body.Pagination.TotalPages != 0 {
		t.Errorf("expected 0 total pages, got %d", body.Pagination.TotalPages)
	}
}
