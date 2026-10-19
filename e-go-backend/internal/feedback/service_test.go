package feedback

import (
	"context"
	"testing"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/imgbb"
)

func wantCode(t *testing.T, err error, want *apperr.Error) {
	t.Helper()
	appErr, ok := apperr.As(err)
	if !ok || appErr.Code != want.Code {
		t.Fatalf("got %v, want code %s", err, want.Code)
	}
}

func TestCreateReview_HappyPath(t *testing.T) {
	repo := &fakeRepo{elig: OrderItemEligibility{ProductID: "p-1", UserID: "u-1", OrderStatus: "delivered"}}

	got, err := NewService(repo, &fakeUploader{}).CreateReview(context.Background(), "u-1", CreateReviewRequest{
		OrderItemID: "oi-1", Rating: 5, Title: " Great ", Comment: "  Loved it  ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Rating != 5 || got.ProductID != "p-1" {
		t.Fatalf("unexpected review: %+v", got)
	}
	// Product must be derived from the order item, not the client.
	if repo.lastCreate.ProductID != "p-1" || repo.lastCreate.UserID != "u-1" || repo.lastCreate.OrderItemID != "oi-1" {
		t.Fatalf("unexpected params: %+v", repo.lastCreate)
	}
	if repo.lastCreate.Title != "Great" || repo.lastCreate.Comment != "Loved it" {
		t.Fatalf("not trimmed: %+v", repo.lastCreate)
	}
}

func TestCreateReview_RejectsUndelivered(t *testing.T) {
	repo := &fakeRepo{elig: OrderItemEligibility{ProductID: "p-1", UserID: "u-1", OrderStatus: "shipped"}}
	_, err := NewService(repo, &fakeUploader{}).CreateReview(context.Background(), "u-1", CreateReviewRequest{OrderItemID: "oi-1", Rating: 4, Comment: "nice"})
	wantCode(t, err, errOrderNotDelivered)
}

func TestCreateReview_RejectsUnavailableProduct(t *testing.T) {
	repo := &fakeRepo{elig: OrderItemEligibility{ProductID: "", UserID: "u-1", OrderStatus: "delivered"}}
	_, err := NewService(repo, &fakeUploader{}).CreateReview(context.Background(), "u-1", CreateReviewRequest{OrderItemID: "oi-1", Rating: 4, Comment: "nice"})
	wantCode(t, err, errProductUnavailable)
}

func TestCreateReview_PropagatesNotFound(t *testing.T) {
	repo := &fakeRepo{eligErr: errOrderNotFound}
	_, err := NewService(repo, &fakeUploader{}).CreateReview(context.Background(), "u-1", CreateReviewRequest{OrderItemID: "oi-x", Rating: 4, Comment: "nice"})
	wantCode(t, err, errOrderNotFound)
}

func TestCreateReview_Duplicate(t *testing.T) {
	repo := &fakeRepo{
		elig:      OrderItemEligibility{ProductID: "p-1", UserID: "u-1", OrderStatus: "delivered"},
		createErr: errReviewExists,
	}
	_, err := NewService(repo, &fakeUploader{}).CreateReview(context.Background(), "u-1", CreateReviewRequest{OrderItemID: "oi-1", Rating: 4, Comment: "nice"})
	wantCode(t, err, errReviewExists)
}

func TestCreateReview_Validation(t *testing.T) {
	cases := map[string]CreateReviewRequest{
		"missing order item": {OrderItemID: "", Rating: 5, Comment: "hello"},
		"rating too low":     {OrderItemID: "oi-1", Rating: 0, Comment: "hello"},
		"rating too high":    {OrderItemID: "oi-1", Rating: 6, Comment: "hello"},
		"comment too short":  {OrderItemID: "oi-1", Rating: 4, Comment: "hi"},
		"comment empty":      {OrderItemID: "oi-1", Rating: 4, Comment: "   "},
	}
	for name, req := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &fakeRepo{elig: OrderItemEligibility{ProductID: "p-1", OrderStatus: "delivered"}}
			_, err := NewService(repo, &fakeUploader{}).CreateReview(context.Background(), "u-1", req)
			wantCode(t, err, errInvalidReview)
		})
	}
}

func TestUpdateReview_HappyPathAndValidation(t *testing.T) {
	repo := &fakeRepo{updated: ReviewRecord{ID: "r-1", Rating: 3, Comment: "ok, updated"}}
	got, err := NewService(repo, &fakeUploader{}).UpdateReview(context.Background(), "u-1", "r-1", UpdateReviewRequest{Rating: 3, Comment: "ok, updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "r-1" || repo.lastUpdateID != "r-1" {
		t.Fatalf("unexpected update: %+v id=%q", got, repo.lastUpdateID)
	}

	_, err = NewService(&fakeRepo{}, &fakeUploader{}).UpdateReview(context.Background(), "u-1", "r-1", UpdateReviewRequest{Rating: 9, Comment: "ok"})
	wantCode(t, err, errInvalidReview)
}

func TestCreateQuestion_HappyPath(t *testing.T) {
	repo := &fakeRepo{productID: "p-1"}
	got, err := NewService(repo, &fakeUploader{}).CreateQuestion(context.Background(), "u-2", "demo-blender", CreateQuestionRequest{Question: "  Is it dishwasher safe?  "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ProductID != "p-1" || repo.lastQuestion != "Is it dishwasher safe?" {
		t.Fatalf("unexpected question: %+v / %q", got, repo.lastQuestion)
	}
}

func TestCreateQuestion_Validation(t *testing.T) {
	repo := &fakeRepo{productID: "p-1"}
	_, err := NewService(repo, &fakeUploader{}).CreateQuestion(context.Background(), "u-2", "demo", CreateQuestionRequest{Question: "hi"})
	wantCode(t, err, errInvalidQuestion)
}

func TestCreateQuestion_ProductNotFound(t *testing.T) {
	repo := &fakeRepo{productErr: errProductNotFound}
	_, err := NewService(repo, &fakeUploader{}).CreateQuestion(context.Background(), "u-2", "missing", CreateQuestionRequest{Question: "Is it dishwasher safe?"})
	wantCode(t, err, errProductNotFound)
}

func TestCreateReview_PersistsImages(t *testing.T) {
	repo := &fakeRepo{elig: OrderItemEligibility{ProductID: "p-1", OrderStatus: "delivered"}}
	got, err := NewService(repo, &fakeUploader{}).CreateReview(context.Background(), "u-1", CreateReviewRequest{
		OrderItemID: "oi-1", Rating: 5, Comment: "great", Images: []string{" https://i.ibb.co/x.png "},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Images) != 1 || got.Images[0] != "https://i.ibb.co/x.png" {
		t.Fatalf("images not stored/normalized: %+v", got.Images)
	}
	if len(repo.lastCreate.Images) != 1 {
		t.Fatalf("images not passed to repo: %+v", repo.lastCreate.Images)
	}
}

func TestCreateReview_RejectsTooManyImages(t *testing.T) {
	repo := &fakeRepo{elig: OrderItemEligibility{ProductID: "p-1", OrderStatus: "delivered"}}
	_, err := NewService(repo, &fakeUploader{}).CreateReview(context.Background(), "u-1", CreateReviewRequest{
		OrderItemID: "oi-1", Rating: 5, Comment: "great",
		Images: []string{"http://a", "http://b", "http://c", "http://d", "http://e"},
	})
	wantCode(t, err, errInvalidImages)
}

func TestCreateReview_RejectsInvalidImageURL(t *testing.T) {
	repo := &fakeRepo{elig: OrderItemEligibility{ProductID: "p-1", OrderStatus: "delivered"}}
	_, err := NewService(repo, &fakeUploader{}).CreateReview(context.Background(), "u-1", CreateReviewRequest{
		OrderItemID: "oi-1", Rating: 5, Comment: "great", Images: []string{"not-a-url"},
	})
	wantCode(t, err, errInvalidImages)
}

func TestUploadImage(t *testing.T) {
	url, err := NewService(&fakeRepo{}, &fakeUploader{url: "https://i.ibb.co/y.png"}).
		UploadImage(context.Background(), "photo.png", []byte("bytes"))
	if err != nil || url != "https://i.ibb.co/y.png" {
		t.Fatalf("unexpected upload: %q / %v", url, err)
	}

	_, err = NewService(&fakeRepo{}, &fakeUploader{}).UploadImage(context.Background(), "photo.png", nil)
	wantCode(t, err, errInvalidImage)
}

func TestUploadImage_NotConfigured(t *testing.T) {
	_, err := NewService(&fakeRepo{}, &fakeUploader{err: imgbb.ErrNotConfigured}).
		UploadImage(context.Background(), "photo.png", []byte("bytes"))
	wantCode(t, err, errImageUnavailable)
}
