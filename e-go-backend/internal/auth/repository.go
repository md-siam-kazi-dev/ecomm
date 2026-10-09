package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// pgUniqueViolation is PostgreSQL's unique_violation SQLSTATE.
const pgUniqueViolation = "23505"

// defaultRole is assigned to every new account (PRD §5). Seeded by
// migrations/006_seed_roles.sql.
const defaultRole = "user"

var (
	// errEmailTaken is returned when a user already exists for the given email.
	errEmailTaken = apperr.New(http.StatusConflict, "EMAIL_ALREADY_EXISTS", "An account with this email already exists")
	// errUserNotFound is returned when no account matches the lookup.
	errUserNotFound = apperr.New(http.StatusNotFound, "USER_NOT_FOUND", "User not found")
)

// UserRepository is the persistence boundary for accounts. The service depends
// on this interface so it can be exercised without a database.
type UserRepository interface {
	CreateUser(ctx context.Context, params CreateUserParams) (CreatedUser, error)
	FindUserByEmail(ctx context.Context, email string) (User, error)
	FindUserByID(ctx context.Context, id string) (User, error)
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	MarkEmailVerified(ctx context.Context, userID string) error
	TouchLastLogin(ctx context.Context, userID string) error
}

// CreateUserParams carries the data needed to create a user and profile.
type CreateUserParams struct {
	Email        string
	Phone        string
	PasswordHash string
	FirstName    string
	LastName     string
}

// CreatedUser is the subset of the created user returned to the service.
type CreatedUser struct {
	ID        string
	Email     string
	Status    string
	CreatedAt time.Time
}

// Repository persists auth data in PostgreSQL (Handler → Service → Repository).
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// userSelect resolves the display name from user_profiles and collapses
// user_roles into a text[] of role names. Callers append WHERE/GROUP BY.
const userSelect = `
	SELECT
		u.id,
		u.email,
		u.status::text,
		u.email_verified,
		u.password_hash,
		u.created_at,
		COALESCE(u.phone, '') AS phone,
		COALESCE(p.avatar_url, '') AS avatar_url,
		COALESCE(NULLIF(TRIM(CONCAT_WS(' ', p.first_name, p.last_name)), ''), split_part(u.email, '@', 1)) AS name,
		COALESCE(array_agg(r.name::text) FILTER (WHERE r.name IS NOT NULL), '{}') AS roles
	FROM users u
	LEFT JOIN user_profiles p ON p.user_id = u.id
	LEFT JOIN user_roles ur ON ur.user_id = u.id
	LEFT JOIN roles r ON r.id = ur.role_id
`

const userGroupBy = ` GROUP BY u.id, p.first_name, p.last_name, p.avatar_url`

// CreateUser inserts the user, their profile and their default role in a single
// transaction so a failure never leaves a partially-created account. A
// duplicate email maps to errEmailTaken.
func (r *Repository) CreateUser(ctx context.Context, p CreateUserParams) (CreatedUser, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return CreatedUser{}, apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var phone *string
	if p.Phone != "" {
		phone = &p.Phone
	}

	// `status` is left to the column default ('pending' until email
	// verification is implemented — PRD §13).
	var user CreatedUser
	err = tx.QueryRow(ctx, `
		INSERT INTO users (email, phone, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, status, created_at
	`, p.Email, phone, p.PasswordHash).Scan(
		&user.ID, &user.Email, &user.Status, &user.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return CreatedUser{}, errEmailTaken.WithCause(err)
		}
		return CreatedUser{}, apperr.ErrInternal.WithCause(err)
	}

	var firstName, lastName *string
	if p.FirstName != "" {
		firstName = &p.FirstName
	}
	if p.LastName != "" {
		lastName = &p.LastName
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_profiles (user_id, first_name, last_name)
		VALUES ($1, $2, $3)
	`, user.ID, firstName, lastName); err != nil {
		return CreatedUser{}, apperr.ErrInternal.WithCause(err)
	}

	// Assign the default role. The role row must exist (see
	// migrations/006_seed_roles.sql); if it is missing the insert affects no
	// rows and the whole registration rolls back.
	tag, err := tx.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, id FROM roles WHERE name = $2::user_role_type
	`, user.ID, defaultRole)
	if err != nil {
		return CreatedUser{}, apperr.ErrInternal.WithCause(err)
	}
	if tag.RowsAffected() != 1 {
		return CreatedUser{}, apperr.ErrInternal.WithCause(
			fmt.Errorf("default role %q not found; run migrations/006_seed_roles.sql", defaultRole))
	}

	if err := tx.Commit(ctx); err != nil {
		return CreatedUser{}, apperr.ErrInternal.WithCause(err)
	}

	return user, nil
}

// HasAnyRole reports whether the user currently holds any of the given roles,
// reading from user_roles so authorization never relies on a stale token claim.
// A user may hold several roles at once.
func (r *Repository) HasAnyRole(ctx context.Context, userID string, roles ...string) (bool, error) {
	if len(roles) == 0 {
		return false, nil
	}
	var allowed bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			WHERE ur.user_id = $1 AND r.name::text = ANY($2)
		)
	`, userID, roles).Scan(&allowed)
	if err != nil {
		return false, err
	}
	return allowed, nil
}

// FindUserByEmail loads an account (with password hash and effective role) by
// email. Email is matched case-insensitively.
func (r *Repository) FindUserByEmail(ctx context.Context, email string) (User, error) {
	return r.findOne(ctx, userSelect+` WHERE u.email = $1`+userGroupBy, email)
}

// FindUserByID loads an account by primary key.
func (r *Repository) FindUserByID(ctx context.Context, id string) (User, error) {
	return r.findOne(ctx, userSelect+` WHERE u.id = $1`+userGroupBy, id)
}

func (r *Repository) findOne(ctx context.Context, query string, arg any) (User, error) {
	var (
		user  User
		roles []string
	)
	err := r.db.QueryRow(ctx, query, arg).Scan(
		&user.ID, &user.Email, &user.Status, &user.EmailVerified,
		&user.PasswordHash, &user.CreatedAt, &user.Phone, &user.AvatarURL,
		&user.Name, &roles,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, errUserNotFound
	}
	if err != nil {
		return User{}, apperr.ErrInternal.WithCause(err)
	}
	user.Role = highestRole(roles)
	return user, nil
}

// UpdatePassword replaces the stored bcrypt hash.
func (r *Repository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1
	`, userID, passwordHash)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if tag.RowsAffected() == 0 {
		return errUserNotFound
	}
	return nil
}

// MarkEmailVerified flips email_verified and activates a `pending` account.
func (r *Repository) MarkEmailVerified(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users
		SET email_verified = TRUE,
		    status = CASE WHEN status = 'pending' THEN 'active'::user_status ELSE status END,
		    updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}

// TouchLastLogin records a successful sign-in. Failures are non-fatal for the
// caller (best-effort bookkeeping).
func (r *Repository) TouchLastLogin(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET last_login_at = NOW() WHERE id = $1`, userID)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	return nil
}
