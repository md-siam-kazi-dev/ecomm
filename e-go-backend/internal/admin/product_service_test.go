package admin

import (
	"context"
	"testing"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

type fakeProductRepo struct {
	lastQuery     ReviewQuery
	records       []ReviewProduct
	total         int64
	lastProductID string
	lastDecision  Decision
	lastNote      string
}

func (f *fakeProductRepo) ListForReview(_ context.Context, q ReviewQuery) ([]ReviewProduct, int64, error) {
	f.lastQuery = q
	return f.records, f.total, nil
}

func (f *fakeProductRepo) GetForReview(_ context.Context, productID string) (ReviewProductDetail, error) {
	return ReviewProductDetail{ID: productID, VendorName: "Vendor Demo Co"}, nil
}

func (f *fakeProductRepo) SetApproval(_ context.Context, productID, _ string, decision Decision, note string) error {
	f.lastProductID, f.lastDecision, f.lastNote = productID, decision, note
	return nil
}

func TestProductService_ListNormalizes(t *testing.T) {
	repo := &fakeProductRepo{total: 25}
	_, pagination, err := NewProductService(repo).List(context.Background(), ReviewQuery{
		Status: "NONSENSE", Page: -1, Limit: 7,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastQuery.Status != "" || repo.lastQuery.Limit != defaultLimit || repo.lastQuery.Page != 1 {
		t.Fatalf("query not normalized: %+v", repo.lastQuery)
	}
	if pagination.Total != 25 {
		t.Errorf("total = %d, want 25", pagination.Total)
	}
}

func TestProductService_Decisions(t *testing.T) {
	repo := &fakeProductRepo{}
	svc := NewProductService(repo)

	if err := svc.Approve(context.Background(), "p-1", "admin-1"); err != nil {
		t.Fatalf("approve: %v", err)
	}
	if repo.lastDecision != DecisionApprove || repo.lastProductID != "p-1" {
		t.Fatalf("approve not forwarded: %+v", repo)
	}

	if err := svc.Reject(context.Background(), "p-1", "Not as described"); err != nil {
		t.Fatalf("reject: %v", err)
	}
	if repo.lastDecision != DecisionReject || repo.lastNote != "Not as described" {
		t.Fatalf("reject not forwarded: %+v", repo)
	}

	if err := svc.RequestChanges(context.Background(), "p-1", "Add clearer photos"); err != nil {
		t.Fatalf("changes: %v", err)
	}
	if repo.lastDecision != DecisionChanges || repo.lastNote != "Add clearer photos" {
		t.Fatalf("changes not forwarded: %+v", repo)
	}
}

func TestProductService_RequiresReason(t *testing.T) {
	svc := NewProductService(&fakeProductRepo{})

	if err := svc.Reject(context.Background(), "p-1", "   "); apperrCode(err) != apperr.ErrValidation.Code {
		t.Errorf("reject without reason should be invalid, got %v", err)
	}
	if err := svc.RequestChanges(context.Background(), "p-1", ""); apperrCode(err) != apperr.ErrValidation.Code {
		t.Errorf("changes without note should be invalid, got %v", err)
	}
}

func apperrCode(err error) string {
	if appErr, ok := apperr.As(err); ok {
		return appErr.Code
	}
	return ""
}
