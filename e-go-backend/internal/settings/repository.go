package settings

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/internal/audit"
	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// SettingsRepository is the persistence boundary for storefront settings.
type SettingsRepository interface {
	Get(ctx context.Context) (Settings, error)
	Update(ctx context.Context, actorID string, enabled bool, message string) (Settings, error)
}

// Repository persists storefront settings in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// Get returns the singleton settings row.
func (r *Repository) Get(ctx context.Context) (Settings, error) {
	var s Settings
	err := r.db.QueryRow(ctx, `
		SELECT announcement_enabled, announcement_message, updated_at
		FROM site_settings
		WHERE id = $1`, SingletonRowID,
	).Scan(&s.AnnouncementEnabled, &s.AnnouncementMessage, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Settings{}, apperr.ErrNotFound
		}
		return Settings{}, apperr.ErrInternal.WithCause(err)
	}
	return s, nil
}

// Update writes the singleton settings row and records an audit event in the
// same transaction so the change and its audit trail commit atomically.
func (r *Repository) Update(ctx context.Context, actorID string, enabled bool, message string) (Settings, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Settings{}, apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var s Settings
	err = tx.QueryRow(ctx, `
		UPDATE site_settings
		SET announcement_enabled = $1,
		    announcement_message = $2,
		    updated_by = $3,
		    updated_at = NOW()
		WHERE id = $4
		RETURNING announcement_enabled, announcement_message, updated_at`,
		enabled, message, nullableUUID(actorID), SingletonRowID,
	).Scan(&s.AnnouncementEnabled, &s.AnnouncementMessage, &s.UpdatedAt)
	if err != nil {
		return Settings{}, apperr.ErrInternal.WithCause(err)
	}

	if err := audit.Record(ctx, tx, audit.Event{
		ActorID:    &actorID,
		ActorRole:  audit.RoleAdmin,
		Action:     audit.ActionAdminSettingsUpdated,
		EntityType: audit.EntitySettings,
		Result:     audit.ResultSuccess,
		NewValues: map[string]any{
			"announcement_enabled": enabled,
			"announcement_message": message,
		},
	}); err != nil {
		return Settings{}, apperr.ErrInternal.WithCause(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Settings{}, apperr.ErrInternal.WithCause(err)
	}
	return s, nil
}

// nullableUUID converts an actor id string to a uuid.UUID (or SQL NULL when it
// is empty or not a valid UUID).
func nullableUUID(id string) any {
	if id == "" {
		return nil
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil
	}
	return parsed
}
