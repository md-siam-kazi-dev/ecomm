package vendor

import "time"

// AnswerRequest is the payload for answering a question.
type AnswerRequest struct {
	Answer string `json:"answer"`
}

// VendorQuestionResponse is a question on the vendor's products.
type VendorQuestionResponse struct {
	ID              string `json:"id"`
	ProductID       string `json:"product_id"`
	ProductName     string `json:"product_name"`
	ProductSlug     string `json:"product_slug"`
	ProductImage    string `json:"product_image,omitempty"`
	Author          string `json:"author"`
	Question        string `json:"question"`
	CreatedAt       string `json:"created_at"`
	Answered        bool   `json:"answered"`
	Answer          string `json:"answer,omitempty"`
	AnswerCreatedAt string `json:"answer_created_at,omitempty"`
}

func toVendorQuestionResponse(q VendorQuestionRecord) VendorQuestionResponse {
	resp := VendorQuestionResponse{
		ID:           q.ID,
		ProductID:    q.ProductID,
		ProductName:  q.ProductName,
		ProductSlug:  q.ProductSlug,
		ProductImage: q.ProductImage,
		Author:       q.Author,
		Question:     q.Question,
		CreatedAt:    q.CreatedAt.UTC().Format(time.RFC3339),
		Answered:     q.Answered,
		Answer:       q.Answer,
	}
	if q.AnswerCreatedAt != nil {
		resp.AnswerCreatedAt = q.AnswerCreatedAt.UTC().Format(time.RFC3339)
	}
	return resp
}
