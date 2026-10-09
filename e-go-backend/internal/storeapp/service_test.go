package storeapp

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

func newService(repo *fakeRepo, cfg Config) *Service {
	return NewService(repo, nil, cfg)
}

func TestSubmit_Validation(t *testing.T) {
	cases := map[string]func(*SubmitRequest){
		"short name":            func(r *SubmitRequest) { r.StoreName = "A" },
		"missing description":   func(r *SubmitRequest) { r.Description = "" },
		"missing category":      func(r *SubmitRequest) { r.CategoryID = "" },
		"bad email":             func(r *SubmitRequest) { r.BusinessEmail = "not-an-email" },
		"missing phone":         func(r *SubmitRequest) { r.ContactPhone = "" },
		"missing division":      func(r *SubmitRequest) { r.Division = "" },
		"missing district":      func(r *SubmitRequest) { r.District = "" },
		"missing address":       func(r *SubmitRequest) { r.FullAddress = "" },
		"bad business type":     func(r *SubmitRequest) { r.BusinessType = "partnership" },
		"registered needs name": func(r *SubmitRequest) { r.BusinessType = BusinessRegistered },
		"bad website":           func(r *SubmitRequest) { r.WebsiteURL = "ftp://x" },
		"terms not accepted":    func(r *SubmitRequest) { r.AgreedToTerms = false },
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			req := validSubmit()
			mutate(&req)

			_, err := newService(newFakeRepo(), Config{}).Submit(context.Background(), "u-1", req)
			appErr, ok := apperr.As(err)
			if !ok || appErr.Code != apperr.ErrValidation.Code {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestSubmit_Success(t *testing.T) {
	repo := newFakeRepo()
	repo.created = Application{
		ID: "app-1", ApplicationNumber: "SA-X", ApplicantUserID: "u-1",
		StoreName: "Demo Store", StoreSlug: "demo-store", Status: StatusPending,
	}
	svc := newService(repo, Config{})

	detail, err := svc.Submit(context.Background(), "u-1", validSubmit())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(repo.lastCreateParams.ApplicationNumber, "SA-") {
		t.Errorf("application number not generated: %q", repo.lastCreateParams.ApplicationNumber)
	}
	if repo.lastCreateParams.StoreSlug != "demo-store" {
		t.Errorf("slug not normalized: %q", repo.lastCreateParams.StoreSlug)
	}
	if detail.Status != StatusPending || detail.NextAction != "await_review" {
		t.Errorf("unexpected detail: %+v", detail)
	}
}

func TestSubmit_RejectsDuplicateActive(t *testing.T) {
	repo := newFakeRepo()
	repo.activeExists = true

	_, err := newService(repo, Config{}).Submit(context.Background(), "u-1", validSubmit())
	if !errors.Is(err, errActiveApplication) {
		t.Fatalf("expected errActiveApplication, got %v", err)
	}
}

func TestSubmit_RejectsSlugTaken(t *testing.T) {
	repo := newFakeRepo()
	repo.slugTaken = true

	_, err := newService(repo, Config{}).Submit(context.Background(), "u-1", validSubmit())
	if !errors.Is(err, errSlugTaken) {
		t.Fatalf("expected errSlugTaken, got %v", err)
	}
}

func TestSubmit_NormalizesSlugFromName(t *testing.T) {
	repo := newFakeRepo()
	repo.created = Application{ID: "app-1", Status: StatusPending}
	req := validSubmit()
	req.StoreSlug = ""
	req.StoreName = "My Cool Store!"

	if _, err := newService(repo, Config{}).Submit(context.Background(), "u-1", req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastCreateParams.StoreSlug != "my-cool-store" {
		t.Errorf("expected my-cool-store, got %q", repo.lastCreateParams.StoreSlug)
	}
}

func TestReject_RequiresReason(t *testing.T) {
	repo := newFakeRepo()
	err := newService(repo, Config{}).Reject(context.Background(), "admin-1", "app-1", "   ")
	appErr, ok := apperr.As(err)
	if !ok || appErr.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestReject_ForwardsReason(t *testing.T) {
	repo := newFakeRepo()
	if err := newService(repo, Config{}).Reject(context.Background(), "admin-1", "app-1", "  incomplete docs  "); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastReject != "incomplete docs" {
		t.Errorf("reason not trimmed: %q", repo.lastReject)
	}
}

func TestApprove_DelegatesToRepo(t *testing.T) {
	repo := newFakeRepo()
	repo.app = approvedApp()
	svc := newService(repo, Config{FeeEnabled: false})

	if _, err := svc.Approve(context.Background(), "admin-1", "app-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.approveCalled {
		t.Error("repo.Approve was not called")
	}
}

func TestAdminListApplications_NormalizesQuery(t *testing.T) {
	repo := newFakeRepo()
	_, pagination, err := newService(repo, Config{}).AdminListApplications(context.Background(), AdminListQuery{
		Page: 0, Limit: 7, Status: "PENDING", Sort: "weird", Search: strings.Repeat("a", 500),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pagination.Page != 1 || pagination.Limit != defaultLimit {
		t.Errorf("pagination not normalized: %+v", pagination)
	}
	if repo.lastListQuery.Status != StatusPending {
		t.Errorf("status not normalized/lowercased: %q", repo.lastListQuery.Status)
	}
	if repo.lastListQuery.Sort != "submitted_at_desc" {
		t.Errorf("bad sort not reset: %q", repo.lastListQuery.Sort)
	}
	if len(repo.lastListQuery.Search) > maxSearchLen {
		t.Errorf("search not truncated: %d", len(repo.lastListQuery.Search))
	}
}

func TestMyStores_MapsScope(t *testing.T) {
	repo := newFakeRepo()
	repo.stores = []Store{{ID: "s-1", Name: "Demo", Slug: "demo", Status: StoreActive, Role: "owner"}}

	items, err := newService(repo, Config{}).MyStores(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Role != "owner" {
		t.Fatalf("unexpected stores: %+v", items)
	}
}
