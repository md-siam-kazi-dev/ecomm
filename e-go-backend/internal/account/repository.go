package account

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// errAddressNotFound is returned when an address does not exist or belongs to
// another user.
var errAddressNotFound = apperr.New(http.StatusNotFound, "ADDRESS_NOT_FOUND", "Address not found")

// errProfileNotFound is returned when no account matches the user id.
var errProfileNotFound = apperr.New(http.StatusNotFound, "PROFILE_NOT_FOUND", "Account not found")

// AddressRepository is the persistence boundary for user addresses.
type AddressRepository interface {
	ListAddresses(ctx context.Context, userID string) ([]Address, error)
	CreateAddress(ctx context.Context, userID string, p AddressParams) (Address, error)
	UpdateAddress(ctx context.Context, userID, id string, p AddressParams) (Address, error)
	DeleteAddress(ctx context.Context, userID, id string) error
	SetDefault(ctx context.Context, userID, id string) error
}

// ProfileRepository persists the user's editable profile.
type ProfileRepository interface {
	GetProfile(ctx context.Context, userID string) (Profile, error)
	UpdateProfile(ctx context.Context, userID, fullName, phone string) error
	UpdateAvatar(ctx context.Context, userID, avatarURL string) error
}

// Repository persists addresses and profile data in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

const addressColumns = `id, full_name, phone,
	COALESCE(division, ''), COALESCE(district, ''), COALESCE(upazila, ''),
	COALESCE(village_town, ''), COALESCE(clarification, ''), is_default, created_at`

// ListAddresses returns the user's addresses, default first then newest.
func (r *Repository) ListAddresses(ctx context.Context, userID string) ([]Address, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+addressColumns+`
		FROM user_addresses
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at DESC
	`, userID)
	if err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	addresses := make([]Address, 0)
	for rows.Next() {
		addr, err := scanAddress(rows)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, addr)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return addresses, nil
}

// CreateAddress inserts an address. The first address for a user becomes the
// default automatically; requesting default clears it on the others.
func (r *Repository) CreateAddress(ctx context.Context, userID string, p AddressParams) (Address, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Address{}, apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var existing int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM user_addresses WHERE user_id = $1`, userID).Scan(&existing); err != nil {
		return Address{}, apperr.ErrInternal.WithCause(err)
	}

	makeDefault := p.IsDefault || existing == 0
	if makeDefault {
		if _, err := tx.Exec(ctx, `UPDATE user_addresses SET is_default = FALSE, updated_at = NOW() WHERE user_id = $1`, userID); err != nil {
			return Address{}, apperr.ErrInternal.WithCause(err)
		}
	}

	addr, err := scanAddress(tx.QueryRow(ctx, `
		INSERT INTO user_addresses
			(user_id, full_name, phone, division, district, upazila, village_town, clarification, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING `+addressColumns,
		userID, p.FullName, p.Phone, p.Division, p.District, p.Upazila, p.VillageTown, p.Clarification, makeDefault,
	))
	if err != nil {
		return Address{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Address{}, apperr.ErrInternal.WithCause(err)
	}
	return addr, nil
}

// UpdateAddress updates one of the user's addresses.
func (r *Repository) UpdateAddress(ctx context.Context, userID, id string, p AddressParams) (Address, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Address{}, apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if p.IsDefault {
		if _, err := tx.Exec(ctx, `UPDATE user_addresses SET is_default = FALSE, updated_at = NOW() WHERE user_id = $1`, userID); err != nil {
			return Address{}, apperr.ErrInternal.WithCause(err)
		}
	}

	addr, err := scanAddress(tx.QueryRow(ctx, `
		UPDATE user_addresses
		SET full_name = $3, phone = $4, division = $5, district = $6,
		    upazila = $7, village_town = $8, clarification = $9,
		    is_default = $10, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING `+addressColumns,
		id, userID, p.FullName, p.Phone, p.Division, p.District, p.Upazila, p.VillageTown, p.Clarification, p.IsDefault,
	))
	if errors.Is(err, errAddressNotFound) {
		return Address{}, errAddressNotFound
	}
	if err != nil {
		return Address{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Address{}, apperr.ErrInternal.WithCause(err)
	}
	return addr, nil
}

// DeleteAddress removes an address, promoting another to default if needed.
func (r *Repository) DeleteAddress(ctx context.Context, userID, id string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, `DELETE FROM user_addresses WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if tag.RowsAffected() == 0 {
		return errAddressNotFound
	}

	// Keep exactly one default when others remain.
	if _, err := tx.Exec(ctx, `
		UPDATE user_addresses SET is_default = TRUE
		WHERE id = (
			SELECT id FROM user_addresses
			WHERE user_id = $1 AND is_default = FALSE
			ORDER BY created_at DESC
			LIMIT 1
		)
		AND NOT EXISTS (SELECT 1 FROM user_addresses WHERE user_id = $1 AND is_default = TRUE)
	`, userID); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

// SetDefault makes one address the user's default.
func (r *Repository) SetDefault(ctx context.Context, userID, id string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM user_addresses WHERE id = $1 AND user_id = $2)`, id, userID).Scan(&exists); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if !exists {
		return errAddressNotFound
	}

	if _, err := tx.Exec(ctx, `UPDATE user_addresses SET is_default = FALSE, updated_at = NOW() WHERE user_id = $1`, userID); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE user_addresses SET is_default = TRUE, updated_at = NOW() WHERE id = $1 AND user_id = $2`, id, userID); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

// GetProfile loads the user's name, phone and avatar.
func (r *Repository) GetProfile(ctx context.Context, userID string) (Profile, error) {
	var (
		email     string
		phone     string
		firstName string
		lastName  string
		avatarURL string
	)
	err := r.db.QueryRow(ctx, `
		SELECT u.email,
		       COALESCE(u.phone, ''),
		       COALESCE(p.first_name, ''),
		       COALESCE(p.last_name, ''),
		       COALESCE(p.avatar_url, '')
		FROM users u
		LEFT JOIN user_profiles p ON p.user_id = u.id
		WHERE u.id = $1
	`, userID).Scan(&email, &phone, &firstName, &lastName, &avatarURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return Profile{}, errProfileNotFound
	}
	if err != nil {
		return Profile{}, apperr.ErrInternal.WithCause(err)
	}

	fullName := strings.TrimSpace(firstName + " " + lastName)
	if fullName == "" {
		fullName, _ = splitName(email)
	}
	return Profile{FullName: fullName, Email: email, Phone: phone, AvatarURL: avatarURL}, nil
}

// UpdateProfile writes the display name (user_profiles) and phone (users).
func (r *Repository) UpdateProfile(ctx context.Context, userID, fullName, phone string) error {
	first, last := splitName(fullName)

	var phoneValue *string
	if phone != "" {
		phoneValue = &phone
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `UPDATE users SET phone = $2, updated_at = NOW() WHERE id = $1`, userID, phoneValue); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_profiles (user_id, first_name, last_name)
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''))
		ON CONFLICT (user_id)
		DO UPDATE SET first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name, updated_at = NOW()
	`, userID, first, last); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

// UpdateAvatar stores the hosted image URL.
func (r *Repository) UpdateAvatar(ctx context.Context, userID, avatarURL string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_profiles (user_id, avatar_url)
		VALUES ($1, $2)
		ON CONFLICT (user_id)
		DO UPDATE SET avatar_url = EXCLUDED.avatar_url, updated_at = NOW()
	`, userID, avatarURL)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

// scanner is satisfied by both pgx.Row and pgx.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanAddress(row scanner) (Address, error) {
	var a Address
	err := row.Scan(
		&a.ID, &a.FullName, &a.Phone,
		&a.Division, &a.District, &a.Upazila, &a.VillageTown, &a.Clarification,
		&a.IsDefault, &a.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Address{}, errAddressNotFound
	}
	if err != nil {
		return Address{}, apperr.ErrInternal.WithCause(err)
	}
	return a, nil
}
