package feedback

import (
	"context"
	"strings"
)

const (
	minCommentLen  = 3
	maxCommentLen  = 2000
	maxTitleLen    = 255
	minQuestionLen = 5
	maxQuestionLen = 1000
)

// Service owns the customer feedback business logic
// (Handler → Service → Repository).
type Service struct {
	repo FeedbackRepository
}

func NewService(repo FeedbackRepository) *Service { return &Service{repo: repo} }

// CreateReview validates the payload, verifies the caller's delivered purchase
// and publishes the review. The product is derived from the order item.
func (s *Service) CreateReview(ctx context.Context, userID string, req CreateReviewRequest) (ReviewResponse, error) {
	orderItemID := strings.TrimSpace(req.OrderItemID)
	if orderItemID == "" {
		return ReviewResponse{}, errInvalidReview
	}
	rating, title, comment, err := validateReviewBody(req.Rating, req.Title, req.Comment)
	if err != nil {
		return ReviewResponse{}, err
	}

	elig, err := s.repo.EligibleOrderItem(ctx, userID, orderItemID)
	if err != nil {
		return ReviewResponse{}, err
	}
	if elig.ProductID == "" {
		return ReviewResponse{}, errProductUnavailable
	}
	if elig.OrderStatus != orderStatusDelivered {
		return ReviewResponse{}, errOrderNotDelivered
	}

	rec, err := s.repo.CreateReview(ctx, ReviewParams{
		ProductID:   elig.ProductID,
		UserID:      userID,
		OrderItemID: orderItemID,
		Rating:      rating,
		Title:       title,
		Comment:     comment,
	})
	if err != nil {
		return ReviewResponse{}, err
	}
	return toReviewResponse(rec), nil
}

// UpdateReview edits the caller's own review.
func (s *Service) UpdateReview(ctx context.Context, userID, reviewID string, req UpdateReviewRequest) (ReviewResponse, error) {
	rating, title, comment, err := validateReviewBody(req.Rating, req.Title, req.Comment)
	if err != nil {
		return ReviewResponse{}, err
	}

	rec, err := s.repo.UpdateReview(ctx, userID, reviewID, ReviewParams{
		Rating:  rating,
		Title:   title,
		Comment: comment,
	})
	if err != nil {
		return ReviewResponse{}, err
	}
	return toReviewResponse(rec), nil
}

// CreateQuestion validates and stores a question against a live product. No
// purchase is required.
func (s *Service) CreateQuestion(ctx context.Context, userID, slug string, req CreateQuestionRequest) (QuestionResponse, error) {
	question := strings.TrimSpace(req.Question)
	if len(question) < minQuestionLen || len(question) > maxQuestionLen {
		return QuestionResponse{}, errInvalidQuestion
	}

	productID, err := s.repo.LiveProductIDBySlug(ctx, strings.TrimSpace(slug))
	if err != nil {
		return QuestionResponse{}, err
	}
	rec, err := s.repo.CreateQuestion(ctx, userID, productID, question)
	if err != nil {
		return QuestionResponse{}, err
	}
	return toQuestionResponse(rec), nil
}

func validateReviewBody(rating int, title, comment string) (int, string, string, error) {
	if rating < 1 || rating > 5 {
		return 0, "", "", errInvalidReview
	}
	comment = strings.TrimSpace(comment)
	if len(comment) < minCommentLen || len(comment) > maxCommentLen {
		return 0, "", "", errInvalidReview
	}
	title = strings.TrimSpace(title)
	if len(title) > maxTitleLen {
		title = title[:maxTitleLen]
	}
	return rating, title, comment, nil
}
