package feedback

import (
	"context"
)

// fakeRepo is a configurable in-memory FeedbackRepository.
type fakeRepo struct {
	elig    OrderItemEligibility
	eligErr error

	created    ReviewRecord
	createErr  error
	lastCreate ReviewParams

	updated      ReviewRecord
	updateErr    error
	lastUpdateID string
	lastUpdate   ReviewParams

	productID  string
	productErr error

	question     QuestionRecord
	questionErr  error
	lastQuestion string
}

func (f *fakeRepo) EligibleOrderItem(context.Context, string, string) (OrderItemEligibility, error) {
	if f.eligErr != nil {
		return OrderItemEligibility{}, f.eligErr
	}
	return f.elig, nil
}

func (f *fakeRepo) CreateReview(_ context.Context, p ReviewParams) (ReviewRecord, error) {
	f.lastCreate = p
	if f.createErr != nil {
		return ReviewRecord{}, f.createErr
	}
	if f.created.ID == "" {
		return ReviewRecord{ID: "r-1", ProductID: p.ProductID, UserID: p.UserID,
			OrderItemID: p.OrderItemID, Rating: p.Rating, Title: p.Title, Comment: p.Comment,
			Images: p.Images, Status: "published"}, nil
	}
	return f.created, nil
}

func (f *fakeRepo) UpdateReview(_ context.Context, _, reviewID string, p ReviewParams) (ReviewRecord, error) {
	f.lastUpdateID, f.lastUpdate = reviewID, p
	if f.updateErr != nil {
		return ReviewRecord{}, f.updateErr
	}
	return f.updated, nil
}

func (f *fakeRepo) LiveProductIDBySlug(context.Context, string) (string, error) {
	if f.productErr != nil {
		return "", f.productErr
	}
	return f.productID, nil
}

func (f *fakeRepo) CreateQuestion(_ context.Context, userID, productID, question string) (QuestionRecord, error) {
	f.lastQuestion = question
	if f.questionErr != nil {
		return QuestionRecord{}, f.questionErr
	}
	return QuestionRecord{ID: "q-1", ProductID: productID, UserID: userID, Question: question}, nil
}

// fakeUploader stands in for the imgbb host.
type fakeUploader struct {
	url string
	err error
}

func (f *fakeUploader) Upload(_ context.Context, _ string, _ []byte) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	if f.url == "" {
		return "https://i.ibb.co/hosted/review.png", nil
	}
	return f.url, nil
}
