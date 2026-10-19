package catalog

import (
	"context"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// errInvalidSitemapCursor is the internal cause behind a 422 response when a
// caller sends a cursor the API cannot decode.
var errInvalidSitemapCursor = errors.New("invalid sitemap cursor")

// SitemapCursor is the decoded keyset position: the last row returned by the
// previous page, ordered by (updated_at, id). The zero value means "from the
// start".
type SitemapCursor struct {
	UpdatedAt time.Time
	ID        string
}

// IsZero reports whether the cursor points at no particular row.
func (c SitemapCursor) IsZero() bool { return c.ID == "" }

// EncodeSitemapCursor renders a cursor as an opaque, URL-safe token. The payload
// is `<updated_at RFC3339Nano>|<id>` and is base64url encoded without padding.
func EncodeSitemapCursor(updatedAt time.Time, id string) string {
	payload := updatedAt.UTC().Format(time.RFC3339Nano) + "|" + id
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

// DecodeSitemapCursor reverses EncodeSitemapCursor. Any malformed input —
// invalid base64, a missing separator, an unparseable timestamp or a non-UUID id
// — returns errInvalidSitemapCursor so the caller can surface a 422 instead of
// falling back to a full scan.
func DecodeSitemapCursor(raw string) (SitemapCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(raw))
	if err != nil {
		return SitemapCursor{}, errInvalidSitemapCursor
	}

	parts := strings.SplitN(string(decoded), "|", 2)
	if len(parts) != 2 {
		return SitemapCursor{}, errInvalidSitemapCursor
	}

	ts, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(parts[0]))
	if err != nil {
		return SitemapCursor{}, errInvalidSitemapCursor
	}

	id, err := uuid.Parse(strings.TrimSpace(parts[1]))
	if err != nil {
		return SitemapCursor{}, errInvalidSitemapCursor
	}

	return SitemapCursor{UpdatedAt: ts.UTC(), ID: id.String()}, nil
}

// SitemapProducts keyset-scans the indexable products (status `active` and
// approval `approved`) ordered by (updated_at, id) ascending. It runs a single
// parameterised query plus a count, and returns the total number of indexable
// products regardless of the cursor.
func (r *Repository) SitemapProducts(ctx context.Context, cursor SitemapCursor, limit int) ([]SitemapRecord, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM products p WHERE `+liveProducts).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []SitemapRecord{}, 0, nil
	}

	where := liveProducts
	args := make([]any, 0, 3)
	if !cursor.IsZero() {
		var id pgtype.UUID
		if err := id.Scan(cursor.ID); err != nil {
			return nil, 0, apperr.ErrValidation.WithCause(errInvalidSitemapCursor)
		}
		args = append(args, cursor.UpdatedAt, id)
		where += " AND (p.updated_at, p.id) > ($1, $2)"
	}
	args = append(args, limit)

	query := `SELECT p.id, p.slug, p.updated_at FROM products p WHERE ` + where +
		` ORDER BY p.updated_at ASC, p.id ASC LIMIT $` + strconv.Itoa(len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	records := make([]SitemapRecord, 0, limit)
	for rows.Next() {
		var rec SitemapRecord
		if err := rows.Scan(&rec.ID, &rec.Slug, &rec.UpdatedAt); err != nil {
			return nil, 0, apperr.ErrInternal.WithCause(err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	return records, total, nil
}

// SitemapStores keyset-scans the active stores ordered by (updated_at, id)
// ascending, mirroring SitemapProducts.
func (r *Repository) SitemapStores(ctx context.Context, cursor SitemapCursor, limit int) ([]SitemapRecord, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM stores s WHERE s.status = 'active'`).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []SitemapRecord{}, 0, nil
	}

	where := `s.status = 'active'`
	args := make([]any, 0, 3)
	if !cursor.IsZero() {
		var id pgtype.UUID
		if err := id.Scan(cursor.ID); err != nil {
			return nil, 0, apperr.ErrValidation.WithCause(errInvalidSitemapCursor)
		}
		args = append(args, cursor.UpdatedAt, id)
		where += " AND (s.updated_at, s.id) > ($1, $2)"
	}
	args = append(args, limit)

	query := `SELECT s.id, s.slug, s.updated_at FROM stores s WHERE ` + where +
		` ORDER BY s.updated_at ASC, s.id ASC LIMIT $` + strconv.Itoa(len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	records := make([]SitemapRecord, 0, limit)
	for rows.Next() {
		var rec SitemapRecord
		if err := rows.Scan(&rec.ID, &rec.Slug, &rec.UpdatedAt); err != nil {
			return nil, 0, apperr.ErrInternal.WithCause(err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	return records, total, nil
}
