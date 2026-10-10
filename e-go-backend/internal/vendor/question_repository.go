package vendor

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

const vendorQuestionFrom = `
	FROM product_questions q
	JOIN products p ON p.id = q.product_id AND p.vendor_id = $1
	JOIN users u ON u.id = q.user_id
	LEFT JOIN user_profiles up ON up.user_id = q.user_id
	LEFT JOIN LATERAL (
		SELECT url FROM product_images pi
		WHERE pi.product_id = p.id
		ORDER BY pi.is_primary DESC, pi.sort_order
		LIMIT 1
	) img ON TRUE
	LEFT JOIN LATERAL (
		SELECT a.answer, a.created_at
		FROM product_question_answers a
		WHERE a.question_id = q.id AND a.is_official_vendor_answer = TRUE
		ORDER BY a.created_at DESC
		LIMIT 1
	) a ON TRUE`

const vendorQuestionColumns = `
	q.id, q.product_id, p.name, p.slug, COALESCE(img.url, ''),
	COALESCE(NULLIF(TRIM(CONCAT_WS(' ', up.first_name, up.last_name)), ''), split_part(u.email, '@', 1)),
	q.question, q.created_at, a.answer, a.created_at`

func scanVendorQuestion(s rowScanner) (VendorQuestionRecord, error) {
	var (
		rec             VendorQuestionRecord
		answer          *string
		answerCreatedAt *time.Time
	)
	if err := s.Scan(&rec.ID, &rec.ProductID, &rec.ProductName, &rec.ProductSlug, &rec.ProductImage,
		&rec.Author, &rec.Question, &rec.CreatedAt, &answer, &answerCreatedAt); err != nil {
		return rec, err
	}
	if answer != nil {
		rec.Answered = true
		rec.Answer = *answer
		rec.AnswerCreatedAt = answerCreatedAt
	}
	return rec, nil
}

// ListQuestions returns one page of questions on the vendor's products plus the
// total matching count.
func (r *Repository) ListQuestions(ctx context.Context, vendorID string, q ListQuery) ([]VendorQuestionRecord, int64, error) {
	args := []any{vendorID}
	where := ""
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		place := "$" + strconv.Itoa(len(args))
		where += " AND (p.name ILIKE " + place + " OR q.question ILIKE " + place + ")"
	}
	switch q.Status {
	case "unanswered":
		where += " AND a.answer IS NULL"
	case "answered":
		where += " AND a.answer IS NOT NULL"
	}

	var total int64
	if err := r.db.QueryRow(ctx, "SELECT count(*)"+vendorQuestionFrom+where, args...).Scan(&total); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	if total == 0 {
		return []VendorQuestionRecord{}, 0, nil
	}

	listArgs := append(append([]any{}, args...), q.Limit, (q.Page-1)*q.Limit)
	rows, err := r.db.Query(ctx, "SELECT "+vendorQuestionColumns+vendorQuestionFrom+where+`
		ORDER BY q.created_at DESC, q.id
		LIMIT $`+strconv.Itoa(len(args)+1)+` OFFSET $`+strconv.Itoa(len(args)+2), listArgs...)
	if err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	defer rows.Close()

	records := make([]VendorQuestionRecord, 0, q.Limit)
	for rows.Next() {
		rec, err := scanVendorQuestion(rows)
		if err != nil {
			return nil, 0, apperr.ErrInternal.WithCause(err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperr.ErrInternal.WithCause(err)
	}
	return records, total, nil
}

// AnswerQuestion upserts the vendor's single official answer for one of their
// questions and returns the refreshed question.
func (r *Repository) AnswerQuestion(ctx context.Context, vendorID, questionID, userID, answer string) (VendorQuestionRecord, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return VendorQuestionRecord{}, apperr.ErrInternal.WithCause(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var exists int
	err = tx.QueryRow(ctx, `
		SELECT 1 FROM product_questions q
		JOIN products p ON p.id = q.product_id
		WHERE q.id = $1 AND p.vendor_id = $2`, questionID, vendorID).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return VendorQuestionRecord{}, errQuestionNotFound
	}
	if err != nil {
		return VendorQuestionRecord{}, apperr.ErrInternal.WithCause(err)
	}

	tag, err := tx.Exec(ctx, `
		UPDATE product_question_answers
		SET answer = $2, updated_at = NOW()
		WHERE question_id = $1 AND is_official_vendor_answer = TRUE`, questionID, answer)
	if err != nil {
		return VendorQuestionRecord{}, apperr.ErrInternal.WithCause(err)
	}
	if tag.RowsAffected() == 0 {
		if _, err := tx.Exec(ctx, `
			INSERT INTO product_question_answers (question_id, user_id, answer, is_official_vendor_answer)
			VALUES ($1, $2, $3, TRUE)`, questionID, userID, answer); err != nil {
			return VendorQuestionRecord{}, apperr.ErrInternal.WithCause(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return VendorQuestionRecord{}, apperr.ErrInternal.WithCause(err)
	}
	return r.questionByID(ctx, vendorID, questionID)
}

// questionByID loads one question scoped to the vendor's products.
func (r *Repository) questionByID(ctx context.Context, vendorID, questionID string) (VendorQuestionRecord, error) {
	rec, err := scanVendorQuestion(r.db.QueryRow(ctx,
		"SELECT "+vendorQuestionColumns+vendorQuestionFrom+" WHERE q.id = $2", vendorID, questionID))
	if errors.Is(err, pgx.ErrNoRows) {
		return VendorQuestionRecord{}, errQuestionNotFound
	}
	if err != nil {
		return VendorQuestionRecord{}, apperr.ErrInternal.WithCause(err)
	}
	return rec, nil
}
