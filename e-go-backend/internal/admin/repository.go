package admin

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// UserRecord is the raw row returned by the admin user query.
type UserRecord struct {
	ID            string
	Email         string
	Name          string
	AvatarURL     string
	Status        string
	Roles         []string
	EmailVerified bool
	CreatedAt     time.Time
	LastLoginAt   *time.Time
}

// UserRepository is the persistence boundary for admin user management.
type UserRepository interface {
	ListUsers(ctx context.Context, q ListUsersQuery) ([]UserRecord, int64, error)
}

// Repository persists admin data in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// ListUsers returns one page of users plus the total matching count, applying
// search, role filtering, ordering and pagination in the database.
func (r *Repository) ListUsers(ctx context.Context, q ListUsersQuery) ([]UserRecord, int64, error) {
	where, args := buildUserFilter(q)

	var total int64
	countQuery := `SELECT count(*) FROM users u LEFT JOIN user_profiles p ON p.user_id = u.id WHERE ` + where
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []UserRecord{}, 0, nil
	}

	limitPlaceholder := "$" + strconv.Itoa(len(args)+1)
	offsetPlaceholder := "$" + strconv.Itoa(len(args)+2)
	listArgs := append(append([]any{}, args...), q.Limit, (q.Page-1)*q.Limit)

	query := `
		SELECT
			u.id,
			u.email,
			u.status::text,
			u.email_verified,
			u.created_at,
			u.last_login_at,
			COALESCE(NULLIF(TRIM(CONCAT_WS(' ', p.first_name, p.last_name)), ''), split_part(u.email, '@', 1)) AS name,
			COALESCE(p.avatar_url, '') AS avatar_url,
			COALESCE(array_agg(r.name::text) FILTER (WHERE r.name IS NOT NULL), '{}') AS roles
		FROM users u
		LEFT JOIN user_profiles p ON p.user_id = u.id
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		LEFT JOIN roles r ON r.id = ur.role_id
		WHERE ` + where + `
		GROUP BY u.id, p.first_name, p.last_name, p.avatar_url
		ORDER BY u.created_at DESC, u.id
		LIMIT ` + limitPlaceholder + ` OFFSET ` + offsetPlaceholder

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	records := make([]UserRecord, 0, q.Limit)
	for rows.Next() {
		var rec UserRecord
		if err := rows.Scan(
			&rec.ID, &rec.Email, &rec.Status, &rec.EmailVerified,
			&rec.CreatedAt, &rec.LastLoginAt, &rec.Name, &rec.AvatarURL, &rec.Roles,
		); err != nil {
			return nil, 0, apperr.ErrInternal.WithCause(err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}

	return records, total, nil
}

// buildUserFilter returns the shared WHERE clause (with placeholders) and its
// positional arguments for both the count and the page query.
func buildUserFilter(q ListUsersQuery) (string, []any) {
	conditions := []string{"TRUE"}
	args := []any{}

	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		place := "$" + strconv.Itoa(len(args))
		conditions = append(conditions, fmt.Sprintf(
			"(u.email ILIKE %[1]s OR u.id::text ILIKE %[1]s OR p.first_name ILIKE %[1]s OR p.last_name ILIKE %[1]s)",
			place,
		))
	}

	if q.Role != "" {
		args = append(args, q.Role)
		place := "$" + strconv.Itoa(len(args))
		conditions = append(conditions, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM user_roles ur2 JOIN roles r2 ON r2.id = ur2.role_id WHERE ur2.user_id = u.id AND r2.name = %s::user_role_type)",
			place,
		))
	}

	return strings.Join(conditions, " AND "), args
}
