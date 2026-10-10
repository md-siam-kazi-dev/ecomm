package feedback

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// FeedbackRepository is the persistence boundary for customer reviews and
// product questions.
type FeedbackRepository interface {
	EligibleOrderItem(ctx context.Context, userID, orderItemID string) (OrderItemEligibility, error)
	CreateReview(ctx context.Context, p ReviewParams) (ReviewRecord, error)
	UpdateReview(ctx context.Context, userID, reviewID string, p ReviewParams) (ReviewRecord, error)
	LiveProductIDBySlug(ctx context.Context, slug string) (string, error)
	CreateQuestion(ctx context.Context, userID, productID, question string) (QuestionRecord, error)
}

// Repository persists feedback in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

const reviewColumns = `
	id, product_id, user_id, order_item_id, rating, COALESCE(title, ''), content,
	status::text, created_at, updated_at`

func scanReview(s scanner) (ReviewRecord, error) {
	var r ReviewRecord
	err := s.Scan(&r.ID, &r.ProductID, &r.UserID, &r.OrderItemID, &r.Rating, &r.Title,
		&r.Comment, &r.Status, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

type scanner interface{ Scan(dest ...any) error }

// EligibleOrderItem loads the purchase context for one of the caller's order
// items. A missing row means the order item does not exist or is not theirs.
func (r *Repository) EligibleOrderItem(ctx context.Context, userID, orderItemID string) (OrderItemEligibility, error) {
	var e OrderItemEligibility
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(oi.product_id::text, ''), o.user_id, o.status::text
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE oi.id = $1 AND o.user_id = $2`, orderItemID, userID,
	).Scan(&e.ProductID, &e.UserID, &e.OrderStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return OrderItemEligibility{}, errOrderNotFound
	}
	if err != nil {
		return OrderItemEligibility{}, apperr.ErrInternal.WithCause(err)
	}
	return e, nil
}

// CreateReview publishes a review. The unique order_item_id index makes this
// idempotent-safe: a duplicate returns errReviewExists.
func (r *Repository) CreateReview(ctx context.Context, p ReviewParams) (ReviewRecord, error) {
	rec, err := scanReview(r.db.QueryRow(ctx, `
		INSERT INTO reviews (product_id, user_id, order_item_id, rating, title, content, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'published')
		ON CONFLICT (order_item_id) DO NOTHING
		RETURNING `+reviewColumns,
		p.ProductID, p.UserID, p.OrderItemID, p.Rating, nullable(p.Title), p.Comment))
	if errors.Is(err, pgx.ErrNoRows) {
		return ReviewRecord{}, errReviewExists
	}
	if err != nil {
		return ReviewRecord{}, apperr.ErrInternal.WithCause(err)
	}
	return rec, nil
}

// UpdateReview edits the caller's own review.
func (r *Repository) UpdateReview(ctx context.Context, userID, reviewID string, p ReviewParams) (ReviewRecord, error) {
	rec, err := scanReview(r.db.QueryRow(ctx, `
		UPDATE reviews
		SET rating = $3, title = $4, content = $5, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING `+reviewColumns,
		reviewID, userID, p.Rating, nullable(p.Title), p.Comment))
	if errors.Is(err, pgx.ErrNoRows) {
		return ReviewRecord{}, errReviewNotFound
	}
	if err != nil {
		return ReviewRecord{}, apperr.ErrInternal.WithCause(err)
	}
	return rec, nil
}

// LiveProductIDBySlug resolves a publicly visible product to its id.
func (r *Repository) LiveProductIDBySlug(ctx context.Context, slug string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		SELECT id FROM products
		WHERE (slug = $1 OR id::text = $1) AND status = 'active' AND approval_status = 'approved'`, slug).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", errProductNotFound
	}
	if err != nil {
		return "", apperr.ErrInternal.WithCause(err)
	}
	return id, nil
}

// CreateQuestion inserts a published product question.
func (r *Repository) CreateQuestion(ctx context.Context, userID, productID, question string) (QuestionRecord, error) {
	var q QuestionRecord
	err := r.db.QueryRow(ctx, `
		INSERT INTO product_questions (product_id, user_id, question, is_published)
		VALUES ($1, $2, $3, TRUE)
		RETURNING id, product_id, user_id, question, created_at`,
		productID, userID, question,
	).Scan(&q.ID, &q.ProductID, &q.UserID, &q.Question, &q.CreatedAt)
	if err != nil {
		return QuestionRecord{}, apperr.ErrInternal.WithCause(err)
	}
	return q, nil
}

func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
