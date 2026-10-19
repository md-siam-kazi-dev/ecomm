package vendor

import (
	"context"
	"testing"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

func TestReplyToReview_Validation(t *testing.T) {
	_, err := NewService(newFakeRepo(), &fakeUploader{}).ReplyToReview(
		context.Background(), "u-1", "", "r-1", ReplyRequest{Reply: "   "})

	appErr, ok := apperr.As(err)
	if !ok || appErr.Code != errInvalidReply.Code {
		t.Fatalf("expected invalid reply error, got %v", err)
	}
}

func TestReplyToReview_TrimsAndMaps(t *testing.T) {
	repo := newFakeRepo()
	repo.replyRecord = VendorReviewRecord{
		ID: "r-1", ProductID: "p-1", ProductName: "Blender", Rating: 5,
		VendorReply: "Thanks for the feedback!",
	}

	got, err := NewService(repo, &fakeUploader{}).ReplyToReview(
		context.Background(), "u-1", "", "r-1", ReplyRequest{Reply: "  Thanks for the feedback!  "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastReply != "Thanks for the feedback!" {
		t.Fatalf("reply not trimmed: %q", repo.lastReply)
	}
	if !got.Replied || got.VendorReply != "Thanks for the feedback!" {
		t.Fatalf("unexpected response: %+v", got)
	}
}

func TestReplyToReview_ForeignReview(t *testing.T) {
	repo := newFakeRepo()
	repo.replyErr = errReviewNotFound

	_, err := NewService(repo, &fakeUploader{}).ReplyToReview(
		context.Background(), "u-1", "", "r-x", ReplyRequest{Reply: "hello"})

	appErr, ok := apperr.As(err)
	if !ok || appErr.Code != errReviewNotFound.Code {
		t.Fatalf("expected review-not-found, got %v", err)
	}
}

func TestReviewSummary_MapsDistributionAndAverage(t *testing.T) {
	repo := newFakeRepo()
	repo.reviewSummary = ReviewSummaryRecord{Total: 4, Distribution: [5]int64{1, 0, 1, 1, 1}, Unreplied: 2}

	got, err := NewService(repo, &fakeUploader{}).ReviewSummary(context.Background(), "u-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Total != 4 || got.Unreplied != 2 {
		t.Fatalf("unexpected totals: %+v", got)
	}
	if got.Distribution["5"] != 1 || got.Distribution["2"] != 0 {
		t.Fatalf("unexpected distribution: %+v", got.Distribution)
	}
	if _, ok := got.Distribution["1"]; !ok {
		t.Fatalf("distribution missing key 1: %+v", got.Distribution)
	}
	// (1*1 + 2*0 + 3*1 + 4*1 + 5*1) / 4 = 13/4 = 3.25
	if got.Average != 3.25 {
		t.Fatalf("average = %v, want 3.25", got.Average)
	}
}

func TestListReviews_NormalizesFilter(t *testing.T) {
	repo := newFakeRepo()

	if _, _, err := NewService(repo, &fakeUploader{}).ListReviews(
		context.Background(), "u-1", "", ListQuery{Status: "unreplied", Search: "  blend  "}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastReviewQuery.Status != "unreplied" || repo.lastReviewQuery.Search != "blend" {
		t.Fatalf("query not normalized: %+v", repo.lastReviewQuery)
	}

	if _, _, _ = NewService(repo, &fakeUploader{}).ListReviews(
		context.Background(), "u-1", "", ListQuery{Status: "bogus"}); repo.lastReviewQuery.Status != "" {
		t.Fatalf("invalid filter should clear, got %q", repo.lastReviewQuery.Status)
	}
}

func TestAnswerQuestion_ValidationAndMapping(t *testing.T) {
	repo := newFakeRepo()
	repo.answered = VendorQuestionRecord{ID: "q-1", ProductName: "Blender", Answered: true, Answer: "Yes, it is."}

	_, err := NewService(repo, &fakeUploader{}).AnswerQuestion(
		context.Background(), "u-1", "", "q-1", AnswerRequest{Answer: " "})
	if appErr, ok := apperr.As(err); !ok || appErr.Code != errInvalidAnswer.Code {
		t.Fatalf("expected invalid answer error, got %v", err)
	}

	got, err := NewService(repo, &fakeUploader{}).AnswerQuestion(
		context.Background(), "u-1", "", "q-1", AnswerRequest{Answer: "Yes, it is."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Answered || repo.lastAnswer != "Yes, it is." {
		t.Fatalf("unexpected answer: %+v", got)
	}
}

func TestAnswerQuestion_ForeignQuestion(t *testing.T) {
	repo := newFakeRepo()
	repo.answerErr = errQuestionNotFound

	_, err := NewService(repo, &fakeUploader{}).AnswerQuestion(
		context.Background(), "u-1", "", "q-x", AnswerRequest{Answer: "hello"})
	if appErr, ok := apperr.As(err); !ok || appErr.Code != errQuestionNotFound.Code {
		t.Fatalf("expected question-not-found, got %v", err)
	}
}

func TestListQuestions_NormalizesFilter(t *testing.T) {
	repo := newFakeRepo()
	if _, _, err := NewService(repo, &fakeUploader{}).ListQuestions(
		context.Background(), "u-1", "", ListQuery{Status: "unanswered"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastQuestionQuery.Status != "unanswered" {
		t.Fatalf("filter not normalized: %+v", repo.lastQuestionQuery)
	}
}
