package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// Token purposes stored in the auth_tokens.purpose enum.
const (
	purposeEmailVerification = "email_verification"
	purposePasswordReset     = "password_reset"
)

// errInvalidToken is returned for a missing, expired, or already-used one-time
// token. The same error is used for every case so the reason is not leaked.
var errInvalidToken = apperr.New(http.StatusBadRequest, "INVALID_TOKEN", "The link is invalid or has expired")

// TokenParams describes a one-time token to persist (hash only).
type TokenParams struct {
	UserID    string
	Purpose   string
	TokenHash string
	ExpiresAt time.Time
}

// TokenRepository persists one-time verification / password-reset tokens.
type TokenRepository interface {
	Create(ctx context.Context, p TokenParams) error
	// Consume atomically validates and marks a token used, returning its user
	// ID. Returns errInvalidToken when missing/expired/used.
	Consume(ctx context.Context, tokenHash, purpose string) (string, error)
	// InvalidateUnused marks any outstanding tokens of a purpose as used.
	InvalidateUnused(ctx context.Context, userID, purpose string) error
}

// TokenRepositoryImpl is the PostgreSQL-backed implementation.
type TokenRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewTokenRepository(db *pgxpool.Pool) *TokenRepositoryImpl {
	return &TokenRepositoryImpl{db: db}
}

func (r *TokenRepositoryImpl) Create(ctx context.Context, p TokenParams) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO auth_tokens (user_id, purpose, token_hash, expires_at)
		VALUES ($1, $2::auth_token_purpose, $3, $4)
	`, p.UserID, p.Purpose, p.TokenHash, p.ExpiresAt)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

// Consume locks the row, validates it, then marks it used before committing —
// so two concurrent requests cannot both redeem the same token.
func (r *TokenRepositoryImpl) Consume(ctx context.Context, tokenHash, purpose string) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		userID    string
		expiresAt time.Time
		usedAt    *time.Time
	)
	err = tx.QueryRow(ctx, `
		SELECT user_id, expires_at, used_at
		FROM auth_tokens
		WHERE token_hash = $1 AND purpose = $2::auth_token_purpose
		FOR UPDATE
	`, tokenHash, purpose).Scan(&userID, &expiresAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", errInvalidToken
	}
	if err != nil {
		return "", apperr.ErrInternal.WithCause(err)
	}
	if usedAt != nil || time.Now().After(expiresAt) {
		return "", errInvalidToken
	}

	if _, err := tx.Exec(ctx, `UPDATE auth_tokens SET used_at = NOW() WHERE token_hash = $1`, tokenHash); err != nil {
		return "", apperr.ErrInternal.WithCause(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return "", apperr.ErrInternal.WithCause(err)
	}
	return userID, nil
}

func (r *TokenRepositoryImpl) InvalidateUnused(ctx context.Context, userID, purpose string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE auth_tokens SET used_at = NOW()
		WHERE user_id = $1 AND purpose = $2::auth_token_purpose AND used_at IS NULL
	`, userID, purpose)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}
