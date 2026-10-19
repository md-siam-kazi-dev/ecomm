package catalog

import "time"

// SitemapRecord is one indexable URL row for the sitemap: the resource id and
// slug plus its last change timestamp. The (UpdatedAt, ID) pair is the keyset
// coordinate used to page through the scan.
type SitemapRecord struct {
	ID        string
	Slug      string
	UpdatedAt time.Time
}

// SitemapEntryDTO is one row of the public sitemap envelope. `updated_at` is
// rendered with nanosecond precision (RFC3339Nano) so the frontend can echo it
// verbatim as the first half of the next keyset cursor without losing rows that
// share a second but differ in sub-second precision.
type SitemapEntryDTO struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	UpdatedAt string `json:"updated_at"`
}

// SitemapPage is the keyset pagination metadata for a sitemap page. `Total` is
// the count of every indexable row, independent of the cursor. The `has_next`
// flag is derived by the HTTP layer from whether the page filled `Limit`.
type SitemapPage struct {
	Limit int
	Total int64
}

func toSitemapEntryDTO(r SitemapRecord) SitemapEntryDTO {
	return SitemapEntryDTO{
		ID:        r.ID,
		Slug:      r.Slug,
		UpdatedAt: r.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func toSitemapEntryDTOs(records []SitemapRecord) []SitemapEntryDTO {
	items := make([]SitemapEntryDTO, 0, len(records))
	for _, rec := range records {
		items = append(items, toSitemapEntryDTO(rec))
	}
	return items
}
