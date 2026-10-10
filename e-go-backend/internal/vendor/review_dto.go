package vendor

import (
	"strconv"
	"time"
)

// ReplyRequest is the payload for replying to a review.
type ReplyRequest struct {
	Reply string `json:"reply"`
}

// VendorReviewResponse is a review on the vendor's products.
type VendorReviewResponse struct {
	ID              string `json:"id"`
	ProductID       string `json:"product_id"`
	ProductName     string `json:"product_name"`
	ProductSlug     string `json:"product_slug"`
	ProductImage    string `json:"product_image,omitempty"`
	Author          string `json:"author"`
	OrderNumber     string `json:"order_number,omitempty"`
	Rating          int    `json:"rating"`
	Title           string `json:"title"`
	Comment         string `json:"comment"`
	CreatedAt       string `json:"created_at"`
	Replied         bool   `json:"replied"`
	VendorReply     string `json:"vendor_reply,omitempty"`
	VendorRepliedAt string `json:"vendor_replied_at,omitempty"`
}

// ReviewSummaryResponse is the vendor's aggregate star feedback.
type ReviewSummaryResponse struct {
	Average      float64          `json:"average"`
	Total        int64            `json:"total"`
	Distribution map[string]int64 `json:"distribution"`
	Unreplied    int64            `json:"unreplied"`
}

func toVendorReviewResponse(r VendorReviewRecord) VendorReviewResponse {
	resp := VendorReviewResponse{
		ID:           r.ID,
		ProductID:    r.ProductID,
		ProductName:  r.ProductName,
		ProductSlug:  r.ProductSlug,
		ProductImage: r.ProductImage,
		Author:       r.Author,
		OrderNumber:  r.OrderNumber,
		Rating:       r.Rating,
		Title:        r.Title,
		Comment:      r.Comment,
		CreatedAt:    r.CreatedAt.UTC().Format(time.RFC3339),
		VendorReply:  r.VendorReply,
	}
	if r.VendorReply != "" {
		resp.Replied = true
	}
	if r.VendorRepliedAt != nil {
		resp.VendorRepliedAt = r.VendorRepliedAt.UTC().Format(time.RFC3339)
	}
	return resp
}

func toReviewSummaryResponse(s ReviewSummaryRecord) ReviewSummaryResponse {
	distribution := map[string]int64{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0}
	var weighted int64
	for i, count := range s.Distribution {
		stars := i + 1
		distribution[strconv.Itoa(stars)] = count
		weighted += int64(stars) * count
	}

	var average float64
	if s.Total > 0 {
		average = float64(weighted) / float64(s.Total)
	}
	return ReviewSummaryResponse{
		Average:      average,
		Total:        s.Total,
		Distribution: distribution,
		Unreplied:    s.Unreplied,
	}
}
