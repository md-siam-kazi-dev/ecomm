package catalog

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// sitemapBody mirrors the standard response envelope for the sitemap routes.
type sitemapBody struct {
	Success    bool              `json:"success"`
	Data       []SitemapEntryDTO `json:"data"`
	Pagination struct {
		Page       int   `json:"page"`
		Limit      int   `json:"limit"`
		Total      int64 `json:"total"`
		TotalPages int   `json:"total_pages"`
		HasNext    bool  `json:"has_next"`
	} `json:"pagination"`
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

func decodeSitemap(t *testing.T, rec *httptest.ResponseRecorder) sitemapBody {
	t.Helper()
	var body sitemapBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, rec.Body.String())
	}
	return body
}

// uuidN returns a deterministic, valid UUID string for fixtures (the cursor
// decoder requires a real UUID).
func uuidN(n int) string {
	return fmt.Sprintf("00000000-0000-0000-0000-%012d", n)
}

// cursorFor builds the next keyset cursor from a returned row exactly as the
// frontend does: base64url(`<updated_at>|<id>`).
func cursorFor(t *testing.T, row SitemapEntryDTO) string {
	t.Helper()
	if row.ID == "" || row.UpdatedAt == "" {
		t.Fatalf("cannot build cursor from %+v", row)
	}
	return base64.RawURLEncoding.EncodeToString([]byte(row.UpdatedAt + "|" + row.ID))
}

func TestSitemapProducts_ExcludesNonPublic(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepo{sitemapProducts: []sitemapFixture{
		{ID: uuidN(1), Slug: "published", UpdatedAt: base, Status: "active", Approval: "approved"},
		{ID: uuidN(2), Slug: "draft", UpdatedAt: base.Add(time.Hour), Status: "draft", Approval: "pending"},
		{ID: uuidN(3), Slug: "rejected", UpdatedAt: base.Add(2 * time.Hour), Status: "active", Approval: "rejected"},
		{ID: uuidN(4), Slug: "inactive", UpdatedAt: base.Add(3 * time.Hour), Status: "inactive", Approval: "approved"},
		{ID: uuidN(5), Slug: "pending-approval", UpdatedAt: base.Add(4 * time.Hour), Status: "active", Approval: "pending"},
		{ID: uuidN(6), Slug: "also-published", UpdatedAt: base.Add(5 * time.Hour), Status: "active", Approval: "approved"},
	}}

	body := decodeSitemap(t, get(newRouter(repo), "/sitemap/products"))
	if !body.Success {
		t.Fatalf("expected success envelope, got %+v", body)
	}
	if len(body.Data) != 2 || body.Pagination.Total != 2 {
		t.Fatalf("expected only the 2 public products, got %+v (total %d)", body.Data, body.Pagination.Total)
	}
	want := map[string]bool{"published": true, "also-published": true}
	for _, row := range body.Data {
		if !want[row.Slug] {
			t.Errorf("non-public product leaked: %q", row.Slug)
		}
		if row.ID == "" || row.UpdatedAt == "" {
			t.Errorf("row missing id/updated_at: %+v", row)
		}
	}
}

func TestSitemapStores_OnlyActive(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepo{sitemapStores: []sitemapFixture{
		{ID: uuidN(1), Slug: "live-store", UpdatedAt: base, Status: "active"},
		{ID: uuidN(2), Slug: "awaiting", UpdatedAt: base.Add(time.Hour), Status: "pending_activation"},
		{ID: uuidN(3), Slug: "stopped", UpdatedAt: base.Add(2 * time.Hour), Status: "suspended"},
	}}

	rec := get(newRouter(repo), "/sitemap/stores")
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}
	body := decodeSitemap(t, rec)
	if len(body.Data) != 1 || body.Data[0].Slug != "live-store" {
		t.Fatalf("expected only the active store, got %+v", body.Data)
	}
	if body.Pagination.Total != 1 {
		t.Fatalf("total = %d, want 1", body.Pagination.Total)
	}
}

func TestSitemapProducts_KeysetPagination(t *testing.T) {
	base := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	// p2/p3 share a timestamp to exercise the id tie-break in the keyset order.
	repo := &fakeRepo{sitemapProducts: []sitemapFixture{
		{ID: uuidN(1), Slug: "p1", UpdatedAt: base, Status: "active", Approval: "approved"},
		{ID: uuidN(2), Slug: "p2", UpdatedAt: base.Add(time.Minute), Status: "active", Approval: "approved"},
		{ID: uuidN(3), Slug: "p3", UpdatedAt: base.Add(time.Minute), Status: "active", Approval: "approved"},
		{ID: uuidN(4), Slug: "p4", UpdatedAt: base.Add(2 * time.Minute), Status: "active", Approval: "approved"},
		{ID: uuidN(5), Slug: "p5", UpdatedAt: base.Add(3 * time.Minute), Status: "active", Approval: "approved"},
	}}
	router := newRouter(repo)

	seen := map[string]bool{}
	var cursor string

	// Page 1 (no cursor).
	body := decodeSitemap(t, get(router, "/sitemap/products?limit=2"))
	if len(body.Data) != 2 {
		t.Fatalf("page 1 rows = %d, want 2", len(body.Data))
	}
	if body.Pagination.Total != 5 || !body.Pagination.HasNext {
		t.Fatalf("page 1 pagination: %+v", body.Pagination)
	}
	cursor = cursorFor(t, body.Data[len(body.Data)-1])
	for _, row := range body.Data {
		if seen[row.Slug] {
			t.Fatalf("duplicate slug %q within page 1", row.Slug)
		}
		seen[row.Slug] = true
	}

	// Page 2.
	body = decodeSitemap(t, get(router, "/sitemap/products?limit=2&cursor="+url.QueryEscape(cursor)))
	if len(body.Data) != 2 {
		t.Fatalf("page 2 rows = %d, want 2", len(body.Data))
	}
	if body.Pagination.Total != 5 || !body.Pagination.HasNext {
		t.Fatalf("page 2 pagination: %+v", body.Pagination)
	}
	for _, row := range body.Data {
		if seen[row.Slug] {
			t.Fatalf("page 2 overlaps page 1 at %q", row.Slug)
		}
		seen[row.Slug] = true
	}
	cursor = cursorFor(t, body.Data[len(body.Data)-1])

	// Page 3 (final, short).
	body = decodeSitemap(t, get(router, "/sitemap/products?limit=2&cursor="+url.QueryEscape(cursor)))
	if len(body.Data) != 1 {
		t.Fatalf("page 3 rows = %d, want 1", len(body.Data))
	}
	if body.Pagination.Total != 5 || body.Pagination.HasNext {
		t.Fatalf("page 3 pagination: %+v", body.Pagination)
	}
	for _, row := range body.Data {
		if seen[row.Slug] {
			t.Fatalf("page 3 overlaps an earlier page at %q", row.Slug)
		}
		seen[row.Slug] = true
	}
	if len(seen) != 5 {
		t.Fatalf("paged through %d distinct products, want 5", len(seen))
	}
}

func TestSitemap_MalformedCursorIsValidation(t *testing.T) {
	repo := &fakeRepo{}
	valid := time.Now().UTC().Format(time.RFC3339Nano)
	cases := []string{
		"%%%not-base64%%%",
		base64.RawURLEncoding.EncodeToString([]byte("no-separator-here")),
		base64.RawURLEncoding.EncodeToString([]byte("not-a-timestamp|" + uuidN(1))),
		base64.RawURLEncoding.EncodeToString([]byte(valid + "|not-a-uuid")),
	}

	for _, cursor := range cases {
		for _, path := range []string{"/sitemap/products", "/sitemap/stores"} {
			rec := get(newRouter(repo), path+"?cursor="+url.QueryEscape(cursor))
			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("%s cursor %q: got %d, want 422", path, cursor, rec.Code)
			}
			if code := decodeSitemap(t, rec).Error.Code; code != "VALIDATION_ERROR" {
				t.Fatalf("%s cursor %q: error code %q, want VALIDATION_ERROR", path, cursor, code)
			}
		}
	}
}

func TestSitemap_LimitClamping(t *testing.T) {
	repo := &fakeRepo{}
	get(newRouter(repo), "/sitemap/products?limit=999999")
	if repo.lastSitemapLimit != maxSitemapLimit {
		t.Errorf("limit above max = %d, want %d", repo.lastSitemapLimit, maxSitemapLimit)
	}

	get(newRouter(repo), "/sitemap/stores")
	if repo.lastSitemapLimit != defaultSitemapLimit {
		t.Errorf("default limit = %d, want %d", repo.lastSitemapLimit, defaultSitemapLimit)
	}

	get(newRouter(repo), "/sitemap/products?limit=250")
	if repo.lastSitemapLimit != 250 {
		t.Errorf("explicit limit = %d, want 250", repo.lastSitemapLimit)
	}
}

func TestSitemapCursor_RoundTrip(t *testing.T) {
	ts := time.Date(2026, 3, 4, 5, 6, 7, 123456789, time.UTC)
	raw := EncodeSitemapCursor(ts, uuidN(7))

	got, err := DecodeSitemapCursor(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.UpdatedAt.Equal(ts) {
		t.Errorf("updated_at = %s, want %s", got.UpdatedAt, ts)
	}
	if got.ID != uuidN(7) {
		t.Errorf("id = %q, want %q", got.ID, uuidN(7))
	}
	if got.IsZero() {
		t.Error("decoded cursor should not be zero")
	}
}

func TestSitemapProducts_EmptyCatalog(t *testing.T) {
	body := decodeSitemap(t, get(newRouter(&fakeRepo{}), "/sitemap/products"))
	if !body.Success || len(body.Data) != 0 || body.Pagination.Total != 0 || body.Pagination.HasNext {
		t.Fatalf("unexpected empty-catalog response: %+v", body)
	}
}
