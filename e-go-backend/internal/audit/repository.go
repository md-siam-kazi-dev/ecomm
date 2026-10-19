package audit

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

func isNoRows(err error) bool { return errors.Is(err, pgx.ErrNoRows) }

// ListQuery is the normalised filter/pagination input for the audit list.
type ListQuery struct {
	Page       int
	Limit      int
	Search     string
	Action     string
	ActorRole  string
	EntityType string
	EntityID   string
	Result     string
	From       *time.Time
	To         *time.Time
}

// EventRecord is a single audit row (list projection: no payloads).
type EventRecord struct {
	ID         string
	CreatedAt  time.Time
	Action     string
	ActorID    *string
	ActorName  string
	ActorEmail *string
	ActorRole  *string
	EntityType *string
	EntityID   *string
	Result     *string
	Reason     *string
}

// DetailRecord is a full audit row including payloads and request metadata.
type DetailRecord struct {
	EventRecord
	OldValues map[string]any
	NewValues map[string]any
	Metadata  map[string]any
	IPAddress *string
	UserAgent *string
	RequestID *string
}

// SummaryCounts aggregates events for the selected time range.
type SummaryCounts struct {
	Total          int64
	AdminActions   int64
	VendorActions  int64
	SecurityEvents int64
}

// Repository is the read model for audit_logs. It is read-only by design: rows
// are inserted through Recorder and never updated or deleted through the API.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// List returns one page of audit events plus the total matching count.
func (r *Repository) List(ctx context.Context, q ListQuery) ([]EventRecord, int64, error) {
	where, args := buildFilter(q)

	var total int64
	countSQL := `SELECT count(*) FROM audit_logs al LEFT JOIN users u ON u.id = al.user_id LEFT JOIN user_profiles p ON p.user_id = al.user_id WHERE ` + where
	if err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []EventRecord{}, 0, nil
	}

	limitPlace := "$" + strconv.Itoa(len(args)+1)
	offsetPlace := "$" + strconv.Itoa(len(args)+2)
	listArgs := append(append([]any{}, args...), q.Limit, (q.Page-1)*q.Limit)

	query := `
		SELECT
			al.id,
			al.created_at,
			al.action,
			u.id::text,
			COALESCE(NULLIF(TRIM(CONCAT_WS(' ', p.first_name, p.last_name)), ''), split_part(u.email, '@', 1), 'System') AS actor_name,
			u.email,
			al.actor_role,
			al.entity_type,
			al.entity_id::text,
			al.result,
			al.reason
		FROM audit_logs al
		LEFT JOIN users u ON u.id = al.user_id
		LEFT JOIN user_profiles p ON p.user_id = al.user_id
		WHERE ` + where + `
		ORDER BY al.created_at DESC, al.id DESC
		LIMIT ` + limitPlace + ` OFFSET ` + offsetPlace

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	out := make([]EventRecord, 0, q.Limit)
	for rows.Next() {
		var rec EventRecord
		if err := rows.Scan(
			&rec.ID, &rec.CreatedAt, &rec.Action, &rec.ActorID, &rec.ActorName,
			&rec.ActorEmail, &rec.ActorRole, &rec.EntityType, &rec.EntityID,
			&rec.Result, &rec.Reason,
		); err != nil {
			return nil, 0, apperr.ErrInternal.WithCause(err)
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	return out, total, nil
}

// Get returns a single audit event with its full payload, or apperr.ErrNotFound.
func (r *Repository) Get(ctx context.Context, id string) (DetailRecord, error) {
	query := `
		SELECT
			al.id,
			al.created_at,
			al.action,
			u.id::text,
			COALESCE(NULLIF(TRIM(CONCAT_WS(' ', p.first_name, p.last_name)), ''), split_part(u.email, '@', 1), 'System') AS actor_name,
			u.email,
			al.actor_role,
			al.entity_type,
			al.entity_id::text,
			al.result,
			al.reason,
			al.old_values,
			al.new_values,
			al.metadata,
			host(al.ip_address),
			al.user_agent,
			al.request_id
		FROM audit_logs al
		LEFT JOIN users u ON u.id = al.user_id
		LEFT JOIN user_profiles p ON p.user_id = al.user_id
		WHERE al.id = $1`

	var (
		rec                     DetailRecord
		oldRaw, newRaw, metaRaw []byte
	)
	err := r.db.QueryRow(ctx, query, id).Scan(
		&rec.ID, &rec.CreatedAt, &rec.Action, &rec.ActorID, &rec.ActorName,
		&rec.ActorEmail, &rec.ActorRole, &rec.EntityType, &rec.EntityID,
		&rec.Result, &rec.Reason, &oldRaw, &newRaw, &metaRaw,
		&rec.IPAddress, &rec.UserAgent, &rec.RequestID,
	)
	if err != nil {
		if isNoRows(err) {
			return DetailRecord{}, apperr.ErrNotFound
		}
		return DetailRecord{}, apperr.ErrInternal.WithCause(err)
	}
	rec.OldValues = decodeObject(oldRaw)
	rec.NewValues = decodeObject(newRaw)
	rec.Metadata = decodeObject(metaRaw)
	return rec, nil
}

// Summary aggregates totals for the selected range, ignoring list-only filters
// (search/action/actor) so the header always reflects the whole window.
func (r *Repository) Summary(ctx context.Context, q ListQuery) (SummaryCounts, error) {
	where, args := buildDateFilter(q)
	query := `
		SELECT
			count(*),
			count(*) FILTER (WHERE actor_role = 'admin'),
			count(*) FILTER (WHERE actor_role = 'vendor'),
			count(*) FILTER (WHERE action LIKE 'auth.%' OR result = 'failure')
		FROM audit_logs al
		WHERE ` + where

	var s SummaryCounts
	if err := r.db.QueryRow(ctx, query, args...).Scan(&s.Total, &s.AdminActions, &s.VendorActions, &s.SecurityEvents); err != nil {
		return SummaryCounts{}, apperr.ErrInternal.WithCause(err)
	}
	return s, nil
}

// buildFilter returns the WHERE clause and args for the list/count queries.
func buildFilter(q ListQuery) (string, []any) {
	conditions, args := buildDateFilter(q)
	conds := []string{conditions}

	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		p := "$" + strconv.Itoa(len(args))
		conds = append(conds, "(al.action ILIKE "+p+" OR u.email ILIKE "+p+" OR u.id::text ILIKE "+p+" OR p.first_name ILIKE "+p+" OR p.last_name ILIKE "+p+" OR al.entity_id::text ILIKE "+p+")")
	}
	if q.Action != "" {
		args = append(args, q.Action)
		conds = append(conds, "al.action = $"+strconv.Itoa(len(args)))
	}
	if q.ActorRole != "" {
		args = append(args, q.ActorRole)
		conds = append(conds, "al.actor_role = $"+strconv.Itoa(len(args)))
	}
	if q.EntityType != "" {
		args = append(args, q.EntityType)
		conds = append(conds, "al.entity_type = $"+strconv.Itoa(len(args)))
	}
	if q.EntityID != "" {
		if _, err := uuid.Parse(q.EntityID); err == nil {
			args = append(args, q.EntityID)
			conds = append(conds, "al.entity_id = $"+strconv.Itoa(len(args))+"::uuid")
		}
	}
	if q.Result != "" {
		args = append(args, q.Result)
		conds = append(conds, "al.result = $"+strconv.Itoa(len(args)))
	}
	return strings.Join(conds, " AND "), args
}

// buildDateFilter returns the date-range predicate shared by list and summary.
func buildDateFilter(q ListQuery) (string, []any) {
	conds := []string{"TRUE"}
	args := []any{}
	if q.From != nil {
		args = append(args, *q.From)
		conds = append(conds, "al.created_at >= $"+strconv.Itoa(len(args)))
	}
	if q.To != nil {
		args = append(args, *q.To)
		conds = append(conds, "al.created_at < $"+strconv.Itoa(len(args)))
	}
	return strings.Join(conds, " AND "), args
}

func decodeObject(raw []byte) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	return m
}
