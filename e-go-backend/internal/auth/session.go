package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// SessionRecord is a stored refresh-token session.
type SessionRecord struct {
	ID          string
	UserID      string
	RefreshHash string
	ExpiresAt   time.Time
	RevokedAt   *time.Time
}

// SessionParams describes a new session to persist.
type SessionParams struct {
	UserID      string
	RefreshHash string
	UserAgent   string
	IPAddress   string
	ExpiresAt   time.Time
}

// SessionRepository persists refresh-token sessions.
type SessionRepository interface {
	Create(ctx context.Context, p SessionParams) (SessionRecord, error)
	FindByRefreshHash(ctx context.Context, hash string) (SessionRecord, error)
	// Rotate atomically revokes revokeID and inserts the replacement session.
	Rotate(ctx context.Context, revokeID string, p SessionParams) (SessionRecord, error)
	RevokeByHash(ctx context.Context, hash string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

// SessionRepositoryImpl is the PostgreSQL-backed implementation.
type SessionRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepositoryImpl {
	return &SessionRepositoryImpl{db: db}
}

var errSessionNotFound = errors.New("auth: session not found")

const sessionColumns = `id, user_id, refresh_token_hash, expires_at, revoked_at`

func (r *SessionRepositoryImpl) Create(ctx context.Context, p SessionParams) (SessionRecord, error) {
	return scanSession(r.db.QueryRow(ctx, `
		INSERT INTO user_sessions (user_id, refresh_token_hash, user_agent, ip_address, expires_at, last_used_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING `+sessionColumns,
		p.UserID, p.RefreshHash, p.UserAgent, nullableIP(p.IPAddress), p.ExpiresAt,
	))
}

func (r *SessionRepositoryImpl) FindByRefreshHash(ctx context.Context, hash string) (SessionRecord, error) {
	rec, err := scanSession(r.db.QueryRow(ctx,
		`SELECT `+sessionColumns+` FROM user_sessions WHERE refresh_token_hash = $1`, hash))
	if errors.Is(err, errSessionNotFound) {
		return SessionRecord{}, errSessionNotFound
	}
	return rec, err
}

// Rotate revokes the old session and inserts its replacement in one
// transaction, so a refresh can never leave two valid tokens behind.
func (r *SessionRepositoryImpl) Rotate(ctx context.Context, revokeID string, p SessionParams) (SessionRecord, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return SessionRecord{}, apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		UPDATE user_sessions SET revoked_at = NOW(), last_used_at = NOW() WHERE id = $1
	`, revokeID); err != nil {
		return SessionRecord{}, apperr.ErrInternal.WithCause(err)
	}

	rec, err := scanSession(tx.QueryRow(ctx, `
		INSERT INTO user_sessions (user_id, refresh_token_hash, user_agent, ip_address, expires_at, last_used_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING `+sessionColumns,
		p.UserID, p.RefreshHash, p.UserAgent, nullableIP(p.IPAddress), p.ExpiresAt,
	))
	if err != nil {
		return SessionRecord{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return SessionRecord{}, apperr.ErrInternal.WithCause(err)
	}
	return rec, nil
}

func (r *SessionRepositoryImpl) RevokeByHash(ctx context.Context, hash string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE user_sessions SET revoked_at = NOW()
		WHERE refresh_token_hash = $1 AND revoked_at IS NULL
	`, hash)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

func (r *SessionRepositoryImpl) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE user_sessions SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

func scanSession(row pgx.Row) (SessionRecord, error) {
	var rec SessionRecord
	err := row.Scan(&rec.ID, &rec.UserID, &rec.RefreshHash, &rec.ExpiresAt, &rec.RevokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return SessionRecord{}, errSessionNotFound
	}
	if err != nil {
		return SessionRecord{}, apperr.ErrInternal.WithCause(err)
	}
	return rec, nil
}

// nullableIP maps an empty IP to NULL so it fits the INET column.
func nullableIP(ip string) *string {
	if ip == "" {
		return nil
	}
	return &ip
}
