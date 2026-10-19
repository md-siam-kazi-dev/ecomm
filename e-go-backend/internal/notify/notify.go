// Package notify writes in-app notifications.
//
// The platform currently delivers notifications in-app only (no email/SMS/push
// worker exists yet), so every writer inserts through this helper to keep the
// row shape consistent. Callers may pass a *pgxpool.Pool or a pgx.Tx so a
// notification can be written inside an existing transaction.
package notify

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgconn"
)

// Querier is satisfied by *pgxpool.Pool and pgx.Tx.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Insert writes one in-app notification. data is optional context (e.g. an
// order number) and is stored as JSONB (NULL when nil).
func Insert(ctx context.Context, q Querier, userID, ntype, title, message string, data any) error {
	var payload any
	if data != nil {
		encoded, err := json.Marshal(data)
		if err != nil {
			return err
		}
		payload = encoded
	}
	_, err := q.Exec(ctx, `
		INSERT INTO notifications (user_id, type, title, message, data)
		VALUES ($1, $2::notification_type, $3, $4, $5::jsonb)`,
		userID, ntype, title, message, payload)
	return err
}
