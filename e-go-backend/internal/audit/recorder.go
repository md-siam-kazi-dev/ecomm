package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Querier is satisfied by both *pgxpool.Pool and pgx.Tx, so the same recording
// path works for standalone events and events written inside a transaction.
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

const insertSQL = `
INSERT INTO audit_logs (
    user_id, actor_role, action, entity_type, entity_id,
    old_values, new_values, result, reason,
    ip_address, user_agent, request_id, metadata
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`

// Recorder writes audit events. It is safe to call on a nil *Recorder: the
// write becomes a no-op, which keeps tests and optional wiring simple.
type Recorder struct {
	db *pgxpool.Pool
}

// NewRecorder builds a recorder backed by the given pool. A nil pool yields a
// disabled recorder.
func NewRecorder(db *pgxpool.Pool) *Recorder {
	return &Recorder{db: db}
}

// Enabled reports whether the recorder has a usable database handle.
func (r *Recorder) Enabled() bool {
	return r != nil && r.db != nil
}

// Log records an event using the recorder's pool outside any transaction. It is
// intended for best-effort auditing of operations that do not manage their own
// transaction. Failures are logged and returned but callers should not abort
// the user's request on a logging error.
func (r *Recorder) Log(ctx context.Context, e Event) error {
	if !r.Enabled() {
		return nil
	}
	return Record(ctx, r.db, e)
}

// Record writes an event using the supplied querier. Passing a pgx.Tx makes the
// audit row commit or roll back atomically with the business change. A nil
// querier falls back to the recorder's pool (when called on a Recorder).
func Record(ctx context.Context, q Querier, e Event) error {
	if q == nil {
		return nil
	}
	if info, ok := RequestInfoFromContext(ctx); ok {
		if e.IP == "" {
			e.IP = info.IP
		}
		if e.UserAgent == "" {
			e.UserAgent = info.UserAgent
		}
		if e.RequestID == "" {
			e.RequestID = info.RequestID
		}
	}

	userID := parseUUID(e.ActorID)
	entityID := parseUUID(e.EntityID)
	ip := parseIP(e.IP)
	result := e.Result
	if result == "" {
		result = ResultSuccess
	}

	oldJSON, err := marshalJSONB(Sanitize(e.OldValues))
	if err != nil {
		return err
	}
	newJSON, err := marshalJSONB(Sanitize(e.NewValues))
	if err != nil {
		return err
	}
	meta := Sanitize(e.Metadata)
	// Preserve an unusable entity reference rather than silently dropping it.
	if entityID == nil && e.EntityID != nil && *e.EntityID != "" {
		if meta == nil {
			meta = map[string]any{}
		}
		meta["entity_ref"] = *e.EntityID
	}
	metaJSON, err := marshalJSONB(meta)
	if err != nil {
		return err
	}

	role := e.ActorRole
	if role == "" {
		role = RoleSystem
	}

	_, err = q.Exec(ctx, insertSQL,
		userID,
		nullableString(role),
		e.Action,
		nullableString(e.EntityType),
		entityID,
		oldJSON,
		newJSON,
		nullableString(result),
		nullableString(e.Reason),
		ip,
		nullableString(e.UserAgent),
		nullableString(e.RequestID),
		metaJSON,
	)
	if err != nil {
		slog.ErrorContext(ctx, "audit write failed",
			slog.String("action", e.Action),
			slog.String("error", err.Error()),
		)
	}
	return err
}

// parseUUID converts a string pointer to a uuid.UUID, returning a nil interface
// (SQL NULL) when the value is empty or not a valid UUID.
func parseUUID(v *string) any {
	if v == nil || *v == "" {
		return nil
	}
	id, err := uuid.Parse(*v)
	if err != nil {
		return nil
	}
	return id
}

// parseIP validates an IP for the INET column, dropping host:port and
// unparseable values.
func parseIP(v string) any {
	if v == "" {
		return nil
	}
	host := v
	if h, _, err := net.SplitHostPort(v); err == nil {
		host = h
	}
	if net.ParseIP(host) == nil {
		return nil
	}
	return host
}

func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}

// marshalJSONB returns nil for an empty map so the column stays SQL NULL.
func marshalJSONB(m map[string]any) ([]byte, error) {
	if len(m) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// compile-time assertion that pgx.Tx satisfies Querier.
var _ Querier = (pgx.Tx)(nil)
