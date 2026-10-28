package security

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// ListQuery is the normalized filter set for the event history.
type ListQuery struct {
	Page   int
	Limit  int
	Method string
	Status int
	IP     string
	UserID string
	// Endpoint matches the route pattern exactly.
	Endpoint  string
	EventType string
	Severity  string
	From      time.Time
	To        time.Time
}

const listColumns = `
	id::text,
	event_id,
	request_id,
	event_type,
	severity,
	user_id::text,
	user_role,
	host(ip_address),
	method,
	endpoint,
	status_code,
	duration_ms,
	user_agent,
	resource_type,
	resource_id::text,
	metadata,
	created_at`

// Repository reads the persisted security events. It is read-only: rows are
// written exclusively through the batched Writer.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository builds a read repository.
func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// List returns one page of events and the total matching count.
func (r *Repository) List(ctx context.Context, q ListQuery) ([]EventRecord, int64, error) {
	where, args := buildFilter(q)

	var total int64
	countSQL := `SELECT count(*) FROM security_events WHERE ` + where
	if err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []EventRecord{}, 0, nil
	}

	limitPlace := "$" + strconv.Itoa(len(args)+1)
	offsetPlace := "$" + strconv.Itoa(len(args)+2)
	listArgs := append(append([]any{}, args...), q.Limit, (q.Page-1)*q.Limit)

	query := `SELECT ` + listColumns + `
		FROM security_events
		WHERE ` + where + `
		ORDER BY created_at DESC, id DESC
		LIMIT ` + limitPlace + ` OFFSET ` + offsetPlace

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	out := make([]EventRecord, 0, q.Limit)
	for rows.Next() {
		rec, err := scanRecord(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	return out, total, nil
}

// Get returns one event by its row id.
func (r *Repository) Get(ctx context.Context, id string) (EventRecord, error) {
	query := `SELECT ` + listColumns + ` FROM security_events WHERE id = $1::uuid`
	row := r.db.QueryRow(ctx, query, id)
	rec, err := scanRecord(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EventRecord{}, apperr.ErrNotFound
		}
		return EventRecord{}, err
	}
	return rec, nil
}

// scanner is satisfied by both pgx.Row and pgx.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanRecord(s scanner) (EventRecord, error) {
	var (
		rec EventRecord
		raw []byte
	)
	err := s.Scan(
		&rec.ID, &rec.EventID, &rec.RequestID, &rec.EventType, &rec.Severity,
		&rec.UserID, &rec.UserRole, &rec.IPAddress, &rec.Method, &rec.Endpoint,
		&rec.StatusCode, &rec.DurationMS, &rec.UserAgent, &rec.ResourceType,
		&rec.ResourceID, &raw, &rec.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EventRecord{}, err
		}
		return EventRecord{}, apperr.ErrInternal.WithCause(err)
	}
	rec.Metadata = decodeObject(raw)
	return rec, nil
}

func decodeObject(raw []byte) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

// buildFilter assembles a parameterized WHERE clause. Every value is passed as a
// bind parameter; nothing is interpolated into the SQL.
func buildFilter(q ListQuery) (string, []any) {
	conds := make([]string, 0, 8)
	args := make([]any, 0, 8)

	add := func(format string, value any) {
		args = append(args, value)
		conds = append(conds, fmt.Sprintf(format, len(args)))
	}

	if !q.From.IsZero() {
		add("created_at >= $%d", q.From)
	}
	if !q.To.IsZero() {
		add("created_at < $%d", q.To)
	}
	if q.Method != "" {
		add("method = $%d", q.Method)
	}
	if q.Status != 0 {
		add("status_code = $%d", q.Status)
	}
	if q.IP != "" {
		add("ip_address = $%d::inet", q.IP)
	}
	if q.UserID != "" {
		add("user_id = $%d::uuid", q.UserID)
	}
	if q.Endpoint != "" {
		add("endpoint = $%d", q.Endpoint)
	}
	if q.EventType != "" {
		add("event_type = $%d", q.EventType)
	}
	if q.Severity != "" {
		add("severity = $%d", q.Severity)
	}

	if len(conds) == 0 {
		return "TRUE", args
	}
	return strings.Join(conds, " AND "), args
}

// validUUID reports whether a value can be safely cast to uuid in SQL.
func validUUID(value string) bool {
	_, err := uuid.Parse(value)
	return err == nil
}
