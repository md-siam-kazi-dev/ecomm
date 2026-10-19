package feedback

import "time"

// CreateReviewRequest is the review submission payload. The product and user
// are derived server-side from the order item — never sent by the client.
type CreateReviewRequest struct {
	OrderItemID string   `json:"order_item_id"`
	Rating      int      `json:"rating"`
	Title       string   `json:"title"`
	Comment     string   `json:"comment"`
	Images      []string `json:"images"`
}

// UpdateReviewRequest edits an existing review (author only).
type UpdateReviewRequest struct {
	Rating  int      `json:"rating"`
	Title   string   `json:"title"`
	Comment string   `json:"comment"`
	Images  []string `json:"images"`
}

// ReviewResponse is a persisted review owned by the caller.
type ReviewResponse struct {
	ID          string   `json:"id"`
	ProductID   string   `json:"product_id"`
	OrderItemID string   `json:"order_item_id"`
	Rating      int      `json:"rating"`
	Title       string   `json:"title"`
	Comment     string   `json:"comment"`
	Images      []string `json:"images"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// ImageUploadResponse carries the hosted URL of an uploaded image.
type ImageUploadResponse struct {
	URL string `json:"url"`
}

// CreateQuestionRequest is a product question submission.
type CreateQuestionRequest struct {
	Question string `json:"question"`
}

// QuestionResponse is a persisted product question owned by the caller.
type QuestionResponse struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	Question  string `json:"question"`
	CreatedAt string `json:"created_at"`
}

func toReviewResponse(r ReviewRecord) ReviewResponse {
	images := r.Images
	if images == nil {
		images = []string{}
	}
	return ReviewResponse{
		ID:          r.ID,
		ProductID:   r.ProductID,
		OrderItemID: r.OrderItemID,
		Rating:      r.Rating,
		Title:       r.Title,
		Comment:     r.Comment,
		Images:      images,
		CreatedAt:   r.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   r.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toQuestionResponse(q QuestionRecord) QuestionResponse {
	return QuestionResponse{
		ID:        q.ID,
		ProductID: q.ProductID,
		Question:  q.Question,
		CreatedAt: q.CreatedAt.UTC().Format(time.RFC3339),
	}
}
