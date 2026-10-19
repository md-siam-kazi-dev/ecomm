package catalog

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

const (
	// defaultSitemapLimit / maxSitemapLimit clamp the `limit` query parameter.
	defaultSitemapLimit = 1000
	maxSitemapLimit     = 5000
)

// SitemapProducts handles GET /api/v1/sitemap/products — a public, read-only
// keyset page of indexable product slugs.
func (h *Handler) SitemapProducts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, page, err := h.svc.SitemapProducts(r.Context(), r.URL.Query().Get("cursor"), limit)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, sitemapPagination(page.Limit, page.Total, len(items)))
}

// SitemapStores handles GET /api/v1/sitemap/stores — a public, read-only keyset
// page of active store slugs.
func (h *Handler) SitemapStores(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, page, err := h.svc.SitemapStores(r.Context(), r.URL.Query().Get("cursor"), limit)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, sitemapPagination(page.Limit, page.Total, len(items)))
}

// SitemapProducts returns a keyset page of indexable products. A malformed
// cursor is a 422 (validation), never a silent full scan.
func (s *Service) SitemapProducts(ctx context.Context, rawCursor string, limit int) ([]SitemapEntryDTO, SitemapPage, error) {
	cursor, err := parseSitemapCursor(rawCursor)
	if err != nil {
		return nil, SitemapPage{}, err
	}
	limit = clampLimit(limit, defaultSitemapLimit, maxSitemapLimit)

	records, total, err := s.repo.SitemapProducts(ctx, cursor, limit)
	if err != nil {
		return nil, SitemapPage{}, err
	}
	return toSitemapEntryDTOs(records), SitemapPage{Limit: limit, Total: total}, nil
}

// SitemapStores returns a keyset page of active stores. A malformed cursor is a
// 422 (validation), never a silent full scan.
func (s *Service) SitemapStores(ctx context.Context, rawCursor string, limit int) ([]SitemapEntryDTO, SitemapPage, error) {
	cursor, err := parseSitemapCursor(rawCursor)
	if err != nil {
		return nil, SitemapPage{}, err
	}
	limit = clampLimit(limit, defaultSitemapLimit, maxSitemapLimit)

	records, total, err := s.repo.SitemapStores(ctx, cursor, limit)
	if err != nil {
		return nil, SitemapPage{}, err
	}
	return toSitemapEntryDTOs(records), SitemapPage{Limit: limit, Total: total}, nil
}

// parseSitemapCursor decodes an optional cursor, turning a malformed one into a
// 422 application error.
func parseSitemapCursor(raw string) (SitemapCursor, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return SitemapCursor{}, nil
	}
	cursor, err := DecodeSitemapCursor(raw)
	if err != nil {
		return SitemapCursor{}, apperr.ErrValidation.WithCause(err)
	}
	return cursor, nil
}

// sitemapPagination expresses keyset pagination through the shared
// response.Pagination, which derives has_next from page/total_pages. A keyset
// page has no page number, so it is reported as page 1 while it is full; a short
// final page is reported as the last page, which makes has_next false. `total`
// is always the count of indexable rows.
func sitemapPagination(limit int, total int64, returned int) response.Pagination {
	pages := 0
	if limit > 0 {
		pages = int((total + int64(limit) - 1) / int64(limit))
	}
	page := 1
	if returned < limit && pages > 1 {
		page = pages
	}
	return response.Pagination{Page: page, Limit: limit, Total: total}
}
