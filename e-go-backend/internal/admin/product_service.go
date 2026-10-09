package admin

import (
	"context"
	"strings"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

// allowedStatuses are the valid review-queue status filters.
var allowedStatuses = map[string]struct{}{
	"all": {}, "pending": {}, "rejected": {}, "changes_requested": {},
}

// ProductService owns product moderation (Handler → Service → Repository).
type ProductService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// List returns a page of submitted products for review.
func (s *ProductService) List(ctx context.Context, q ReviewQuery) ([]ReviewProductResponse, response.Pagination, error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	limit := q.Limit
	if _, ok := allowedLimits[limit]; !ok {
		limit = defaultLimit
	}

	status := strings.ToLower(strings.TrimSpace(q.Status))
	if _, ok := allowedStatuses[status]; !ok {
		status = ""
	}
	search := strings.TrimSpace(q.Search)
	if len(search) > maxSearchLen {
		search = search[:maxSearchLen]
	}

	records, total, err := s.repo.ListForReview(ctx, ReviewQuery{
		Status: status, Search: search, Page: page, Limit: limit,
	})
	if err != nil {
		return nil, response.Pagination{}, err
	}

	items := make([]ReviewProductResponse, 0, len(records))
	for _, rec := range records {
		items = append(items, toReviewProductResponse(rec))
	}
	return items, response.Pagination{Page: page, Limit: limit, Total: total}, nil
}

// Get returns the full detail for one product.
func (s *ProductService) Get(ctx context.Context, productID string) (ReviewProductDetailResponse, error) {
	detail, err := s.repo.GetForReview(ctx, productID)
	if err != nil {
		return ReviewProductDetailResponse{}, err
	}
	return toReviewProductDetailResponse(detail), nil
}

// Approve publishes an approved product and records the reviewer.
func (s *ProductService) Approve(ctx context.Context, productID, adminID string) error {
	if productID == "" {
		return errProductNotFound
	}
	return s.repo.SetApproval(ctx, productID, adminID, DecisionApprove, "")
}

// Reject keeps the product unpublished and stores the required reason.
func (s *ProductService) Reject(ctx context.Context, productID, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return apperr.ErrValidation
	}
	if productID == "" {
		return errProductNotFound
	}
	return s.repo.SetApproval(ctx, productID, "", DecisionReject, reason)
}

// RequestChanges keeps the product unpublished and stores the feedback.
func (s *ProductService) RequestChanges(ctx context.Context, productID, note string) error {
	note = strings.TrimSpace(note)
	if note == "" {
		return apperr.ErrValidation
	}
	if productID == "" {
		return errProductNotFound
	}
	return s.repo.SetApproval(ctx, productID, "", DecisionChanges, note)
}
