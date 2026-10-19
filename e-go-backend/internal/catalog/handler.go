package catalog

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

// Handler translates public catalog HTTP requests into service calls.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// ListProducts handles GET /api/v1/products.
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	query, ids := queryFrom(r)
	items, pagination, err := h.svc.ListProducts(r.Context(), query, ids)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, pagination)
}

// GetProduct handles GET /api/v1/products/{slug}.
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	product, err := h.svc.GetProduct(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, product, "")
}

// Related handles GET /api/v1/products/{slug}/related.
func (h *Handler) Related(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.svc.Related(r.Context(), chi.URLParam(r, "slug"), limit)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, items, "")
}

// ListReviews handles GET /api/v1/products/{slug}/reviews.
func (h *Handler) ListReviews(w http.ResponseWriter, r *http.Request) {
	page, limit := pageParams(r)
	items, pagination, err := h.svc.Reviews(r.Context(), chi.URLParam(r, "slug"), page, limit)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, pagination)
}

// Rating handles GET /api/v1/products/{slug}/rating.
func (h *Handler) Rating(w http.ResponseWriter, r *http.Request) {
	summary, err := h.svc.Rating(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, summary, "")
}

// Questions handles GET /api/v1/products/{slug}/questions.
func (h *Handler) Questions(w http.ResponseWriter, r *http.Request) {
	page, limit := pageParams(r)
	items, pagination, err := h.svc.Questions(r.Context(), chi.URLParam(r, "slug"), page, limit)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, pagination)
}

// GetStore handles GET /api/v1/stores/{slug}.
func (h *Handler) GetStore(w http.ResponseWriter, r *http.Request) {
	store, err := h.svc.Store(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, store, "")
}

// StoreProducts handles GET /api/v1/stores/{slug}/products.
func (h *Handler) StoreProducts(w http.ResponseWriter, r *http.Request) {
	query, _ := queryFrom(r)
	items, pagination, err := h.svc.StoreProducts(r.Context(), chi.URLParam(r, "slug"), query)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.List(w, items, pagination)
}

// ListCategories handles GET /api/v1/categories.
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.Categories(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, items, "")
}

// TopReviews handles GET /api/v1/reviews/top.
func (h *Handler) TopReviews(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.svc.TopReviews(r.Context(), limit)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, items, "")
}

func pageParams(r *http.Request) (page, limit int) {
	q := r.URL.Query()
	page, _ = strconv.Atoi(q.Get("page"))
	limit, _ = strconv.Atoi(q.Get("limit"))
	return page, limit
}

func queryFrom(r *http.Request) (ListQuery, []string) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	query := ListQuery{
		Category:  q.Get("category"),
		Search:    q.Get("q"),
		MinPrice:  floatPtr(q.Get("min_price")),
		MaxPrice:  floatPtr(q.Get("max_price")),
		MinRating: floatPtr(q.Get("min_rating")),
		InStock:   boolValue(q.Get("in_stock")),
		OnSale:    boolValue(q.Get("on_sale")),
		Sort:      q.Get("sort"),
		Page:      page,
		Limit:     limit,
	}
	return query, splitIDs(q.Get("ids"))
}

func floatPtr(raw string) *float64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil
	}
	return &v
}

func boolValue(raw string) bool {
	return raw == "1" || strings.EqualFold(raw, "true")
}

func splitIDs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	out := make([]string, 0)
	for _, part := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
