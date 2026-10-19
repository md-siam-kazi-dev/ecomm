package catalog

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

func TestRating_ComputesAverageAndDistribution(t *testing.T) {
	repo := &fakeRepo{rating: RatingSummaryRecord{Total: 3, Distribution: [5]int64{0, 0, 1, 1, 1}}}

	got, err := NewService(repo).Rating(context.Background(), "demo-blender")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// (3 + 4 + 5) / 3 = 4.0
	if got.Average != 4.0 || got.Total != 3 {
		t.Fatalf("unexpected summary: %+v", got)
	}
	if got.Distribution["5"] != 1 || got.Distribution["4"] != 1 || got.Distribution["1"] != 0 {
		t.Fatalf("unexpected distribution: %+v", got.Distribution)
	}
}

func TestReviews_PaginatedAndIncludesVendorReply(t *testing.T) {
	repliedAt := time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepo{
		reviews: []ReviewRecord{{
			ID: "r-1", ProductID: "p-1", Author: "Siam", Rating: 5, Title: "Great",
			Comment: "Loved it", VerifiedPurchase: true, VendorReply: "Thanks!",
			VendorRepliedAt: &repliedAt, CreatedAt: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		}},
		reviewsTotal: 25,
	}

	items, pagination, err := NewService(repo).Reviews(context.Background(), "demo-blender", 2, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pagination.Total != 25 || pagination.Page != 2 {
		t.Fatalf("unexpected pagination: %+v", pagination)
	}
	if len(items) != 1 || items[0].VendorReply != "Thanks!" || items[0].VendorRepliedAt == "" {
		t.Fatalf("unexpected review mapping: %+v", items)
	}
}

func TestQuestions_MapsAnswers(t *testing.T) {
	repo := &fakeRepo{
		questions: []QuestionRecord{{
			ID: "q-1", Question: "Dishwasher safe?", Author: "Siam",
			CreatedAt: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			Answers: []AnswerRecord{{
				ID: "a-1", Answer: "Yes, it is.", IsOfficialVendorAnswer: true, Author: "Vendor",
				CreatedAt: time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
			}},
		}},
		questionsTotal: 1,
	}

	items, pagination, err := NewService(repo).Questions(context.Background(), "demo-blender", 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pagination.Total != 1 || len(items) != 1 {
		t.Fatalf("unexpected result: %+v / %+v", items, pagination)
	}
	if len(items[0].Answers) != 1 || !items[0].Answers[0].IsOfficialVendorAnswer {
		t.Fatalf("answers not mapped: %+v", items[0].Answers)
	}
}

func TestListReviews_RoutePaginationEnvelope(t *testing.T) {
	repo := &fakeRepo{
		reviews:      []ReviewRecord{{ID: "r-1", Rating: 5, Author: "Siam", CreatedAt: time.Now()}},
		reviewsTotal: 25,
	}
	rec := get(newRouter(repo), "/products/demo-blender/reviews?page=2&limit=10")
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}

	var body struct {
		Data       []ReviewDTO `json:"data"`
		Pagination struct {
			Page       int  `json:"page"`
			TotalPages int  `json:"total_pages"`
			HasNext    bool `json:"has_next"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body.Pagination.Page != 2 || body.Pagination.TotalPages != 3 || !body.Pagination.HasNext {
		t.Fatalf("unexpected pagination: %+v", body.Pagination)
	}
}

func TestRating_Route(t *testing.T) {
	repo := &fakeRepo{rating: RatingSummaryRecord{Total: 2, Distribution: [5]int64{0, 0, 0, 1, 1}}}
	rec := get(newRouter(repo), "/products/demo-blender/rating")
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}

	var body struct {
		Data RatingSummaryDTO `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body.Data.Total != 2 || body.Data.Average != 4.5 {
		t.Fatalf("unexpected summary: %+v", body.Data)
	}
}

func TestQuestions_Route(t *testing.T) {
	repo := &fakeRepo{questions: []QuestionRecord{{ID: "q-1", Question: "Safe?", Author: "Siam", CreatedAt: time.Now()}}, questionsTotal: 1}
	rec := get(newRouter(repo), "/products/demo-blender/questions")
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}

	var body struct {
		Data       []QuestionDTO `json:"data"`
		Pagination struct {
			Total int `json:"total"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body.Pagination.Total != 1 || len(body.Data) != 1 {
		t.Fatalf("unexpected result: %+v", body)
	}
}

func TestStore_MapsPositiveFeedback(t *testing.T) {
	repo := &fakeRepo{store: StoreRecord{
		ID: "s-1", Name: "FreshMix", Slug: "freshmix", Description: "Blenders",
		LogoURL: "https://i.ibb.co/logo.png", BannerURL: "https://i.ibb.co/banner.png",
		ProductCount: 4, FollowerCount: 12, ReviewCount: 5, PositiveCount: 4,
	}}

	got, err := NewService(repo).Store(context.Background(), "freshmix")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "FreshMix" || got.FollowerCount != 12 || got.ProductCount != 4 {
		t.Fatalf("unexpected store: %+v", got)
	}
	// 4 of 5 positive => 80%.
	if got.PositiveFeedback != 80 {
		t.Fatalf("positive feedback = %v, want 80", got.PositiveFeedback)
	}
}

func TestStore_NotFound(t *testing.T) {
	_, err := NewService(&fakeRepo{storeErr: errStoreNotFound}).Store(context.Background(), "missing")
	if appErr, ok := apperr.As(err); !ok || appErr.Code != "STORE_NOT_FOUND" {
		t.Fatalf("expected STORE_NOT_FOUND, got %v", err)
	}
}

func TestStoreProducts_RouteEnvelope(t *testing.T) {
	repo := &fakeRepo{storeProducts: []ProductRecord{sampleProduct()}, storeProductsTotal: 25}
	rec := get(newRouter(repo), "/stores/freshmix/products?page=2&limit=10")
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}

	var body struct {
		Data       []ProductDTO `json:"data"`
		Pagination struct {
			Page       int `json:"page"`
			TotalPages int `json:"total_pages"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(body.Data) != 1 || body.Pagination.Page != 2 || body.Pagination.TotalPages != 3 {
		t.Fatalf("unexpected result: %+v", body)
	}
	if repo.lastStoreQuery.Page != 2 || repo.lastStoreQuery.Limit != 10 {
		t.Fatalf("store query not forwarded: %+v", repo.lastStoreQuery)
	}
}
