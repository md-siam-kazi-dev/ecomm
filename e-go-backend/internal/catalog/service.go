package catalog

import (
	"context"
	"strings"

	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

const (
	defaultLimit = 12
	maxLimit     = 60
	maxSearchLen = 100
)

var allowedSorts = map[string]struct{}{
	"featured": {}, "newest": {}, "price-asc": {}, "price-desc": {}, "rating": {}, "discount": {},
}

// Service owns the public catalog business logic (Handler → Service → Repository).
type Service struct {
	repo CatalogRepository
}

func NewService(repo CatalogRepository) *Service { return &Service{repo: repo} }

func clampLimit(value, fallback, max int) int {
	if value < 1 {
		return fallback
	}
	if value > max {
		return max
	}
	return value
}

func normalizeQuery(q ListQuery) ListQuery {
	if q.Page < 1 {
		q.Page = 1
	}
	q.Limit = clampLimit(q.Limit, defaultLimit, maxLimit)
	q.Category = strings.TrimSpace(q.Category)
	q.Search = strings.TrimSpace(q.Search)
	if len(q.Search) > maxSearchLen {
		q.Search = q.Search[:maxSearchLen]
	}
	if _, ok := allowedSorts[q.Sort]; !ok {
		q.Sort = "featured"
	}
	return q
}

// ListProducts returns a page of live products.
func (s *Service) ListProducts(ctx context.Context, q ListQuery, ids []string) ([]ProductDTO, response.Pagination, error) {
	q = normalizeQuery(q)

	records, total, err := s.repo.ListProducts(ctx, q, ids)
	if err != nil {
		return nil, response.Pagination{}, err
	}

	items := make([]ProductDTO, 0, len(records))
	for _, rec := range records {
		items = append(items, toProductDTO(rec))
	}
	return items, response.Pagination{Page: q.Page, Limit: q.Limit, Total: total}, nil
}

// GetProduct returns one live product by slug.
func (s *Service) GetProduct(ctx context.Context, slug string) (ProductDTO, error) {
	rec, err := s.repo.ProductBySlug(ctx, slug)
	if err != nil {
		return ProductDTO{}, err
	}
	return toProductDTO(rec), nil
}

// Related returns other live products in the same category.
func (s *Service) Related(ctx context.Context, slug string, limit int) ([]ProductDTO, error) {
	records, err := s.repo.RelatedProducts(ctx, slug, clampLimit(limit, 4, 12))
	if err != nil {
		return nil, err
	}
	return toProductDTOs(records), nil
}

// Categories returns active categories with live product counts.
func (s *Service) Categories(ctx context.Context) ([]CategoryDTO, error) {
	records, err := s.repo.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]CategoryDTO, 0, len(records))
	for _, rec := range records {
		items = append(items, toCategoryDTO(rec))
	}
	return items, nil
}

// Reviews returns the published reviews for a product.
func (s *Service) Reviews(ctx context.Context, slug string, limit int) ([]ReviewDTO, error) {
	records, err := s.repo.ProductReviews(ctx, slug, clampLimit(limit, 50, 100))
	if err != nil {
		return nil, err
	}
	return toReviewDTOs(records), nil
}

// TopReviews returns the top published reviews for the homepage.
func (s *Service) TopReviews(ctx context.Context, limit int) ([]ReviewDTO, error) {
	records, err := s.repo.TopReviews(ctx, clampLimit(limit, 3, 20))
	if err != nil {
		return nil, err
	}
	return toReviewDTOs(records), nil
}

func toProductDTOs(records []ProductRecord) []ProductDTO {
	items := make([]ProductDTO, 0, len(records))
	for _, rec := range records {
		items = append(items, toProductDTO(rec))
	}
	return items
}

func toReviewDTOs(records []ReviewRecord) []ReviewDTO {
	items := make([]ReviewDTO, 0, len(records))
	for _, rec := range records {
		items = append(items, toReviewDTO(rec))
	}
	return items
}
