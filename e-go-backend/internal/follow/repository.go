package follow

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// FollowRepository is the persistence boundary for store follows.
type FollowRepository interface {
	ActiveStoreID(ctx context.Context, slug string) (string, error)
	Follow(ctx context.Context, storeID, userID string) error
	Unfollow(ctx context.Context, storeID, userID string) error
	Following(ctx context.Context, storeID, userID string) (bool, error)
	Count(ctx context.Context, storeID string) (int64, error)
}

// Repository persists store follows in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// ActiveStoreID resolves a public store slug to its id.
func (r *Repository) ActiveStoreID(ctx context.Context, slug string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`SELECT id FROM stores WHERE slug = $1 AND status = 'active'`, slug).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", errStoreNotFound
	}
	if err != nil {
		return "", apperr.ErrInternal.WithCause(err)
	}
	return id, nil
}

// Follow is idempotent.
func (r *Repository) Follow(ctx context.Context, storeID, userID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO store_followers (store_id, user_id) VALUES ($1, $2)
		ON CONFLICT (store_id, user_id) DO NOTHING`, storeID, userID)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

// Unfollow is idempotent.
func (r *Repository) Unfollow(ctx context.Context, storeID, userID string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM store_followers WHERE store_id = $1 AND user_id = $2`, storeID, userID)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

// Following reports whether the user follows the store.
func (r *Repository) Following(ctx context.Context, storeID, userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM store_followers WHERE store_id = $1 AND user_id = $2)`,
		storeID, userID).Scan(&exists)
	if err != nil {
		return false, apperr.ErrInternal.WithCause(err)
	}
	return exists, nil
}

// Count returns the store's follower count.
func (r *Repository) Count(ctx context.Context, storeID string) (int64, error) {
	var n int64
	err := r.db.QueryRow(ctx,
		`SELECT count(*) FROM store_followers WHERE store_id = $1`, storeID).Scan(&n)
	if err != nil {
		return 0, apperr.ErrInternal.WithCause(err)
	}
	return n, nil
}
