package feedback

import (
	"context"
	"errors"
	"strings"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/imgbb"
)

const (
	minCommentLen  = 3
	maxCommentLen  = 2000
	maxTitleLen    = 255
	minQuestionLen = 5
	maxQuestionLen = 1000
	// maxReviewImages caps how many photos a review may attach.
	maxReviewImages = 4
	// maxImageURLLen bounds a stored image URL.
	maxImageURLLen = 1000
	// maxReviewImageBytes caps an uploaded review photo (5 MB).
	maxReviewImageBytes = 5 << 20
)

// ImageUploader hosts an image and returns its public URL.
type ImageUploader interface {
	Upload(ctx context.Context, filename string, data []byte) (string, error)
}

// Service owns the customer feedback business logic
// (Handler → Service → Repository).
type Service struct {
	repo     FeedbackRepository
	uploader ImageUploader
}

func NewService(repo FeedbackRepository, uploader ImageUploader) *Service {
	return &Service{repo: repo, uploader: uploader}
}

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
	images, err := validateImages(req.Images)
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
		Images:      images,
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
	images, err := validateImages(req.Images)
	if err != nil {
		return ReviewResponse{}, err
	}

	rec, err := s.repo.UpdateReview(ctx, userID, reviewID, ReviewParams{
		Rating:  rating,
		Title:   title,
		Comment: comment,
		Images:  images,
	})
	if err != nil {
		return ReviewResponse{}, err
	}
	return toReviewResponse(rec), nil
}

// UploadImage hosts a review photo and returns its URL.
func (s *Service) UploadImage(ctx context.Context, filename string, data []byte) (string, error) {
	if len(data) == 0 || len(data) > maxReviewImageBytes {
		return "", errInvalidImage
	}
	url, err := s.uploader.Upload(ctx, filename, data)
	if err != nil {
		if errors.Is(err, imgbb.ErrNotConfigured) {
			return "", errImageUnavailable
		}
		return "", apperr.ErrInternal.WithCause(err)
	}
	return url, nil
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

// validateImages bounds and sanitizes the optional review photo URLs. Photos
// are optional; a client that skips them sends none.
func validateImages(images []string) ([]string, error) {
	if len(images) > maxReviewImages {
		return nil, errInvalidImages
	}
	out := make([]string, 0, len(images))
	for _, raw := range images {
		url := strings.TrimSpace(raw)
		if url == "" {
			continue
		}
		if len(url) > maxImageURLLen || !strings.HasPrefix(url, "http") {
			return nil, errInvalidImages
		}
		out = append(out, url)
	}
	return out, nil
}
