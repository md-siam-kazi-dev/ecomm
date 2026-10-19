package follow

import (
	"context"
	"testing"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

func TestStatus(t *testing.T) {
	repo := &fakeRepo{storeID: "s-1", following: true, followerCount: 7}
	got, err := NewService(repo).Status(context.Background(), "demo-store", "u-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Following || got.FollowerCount != 7 {
		t.Fatalf("unexpected status: %+v", got)
	}
}

func TestStatus_StoreNotFound(t *testing.T) {
	_, err := NewService(&fakeRepo{storeErr: errStoreNotFound}).Status(context.Background(), "missing", "u-1")
	appErr, ok := apperr.As(err)
	if !ok || appErr.Code != errStoreNotFound.Code {
		t.Fatalf("expected store-not-found, got %v", err)
	}
}

func TestFollowAndUnfollow(t *testing.T) {
	repo := &fakeRepo{storeID: "s-1"}

	followed, err := NewService(repo).Follow(context.Background(), "demo-store", "u-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !followed.Following || followed.FollowerCount != 1 || repo.lastFollowed != "u-1" {
		t.Fatalf("unexpected follow result: %+v", followed)
	}

	unfollowed, err := NewService(repo).Unfollow(context.Background(), "demo-store", "u-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if unfollowed.Following || unfollowed.FollowerCount != 0 || repo.lastUnfollowed != "u-1" {
		t.Fatalf("unexpected unfollow result: %+v", unfollowed)
	}
}
