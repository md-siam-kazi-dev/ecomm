package settings

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

func TestGet_FallsBackToDefaultsWhenRowMissing(t *testing.T) {
	svc := NewService(&fakeRepo{getErr: apperr.ErrNotFound})

	item, err := svc.Get(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !item.AnnouncementEnabled || item.AnnouncementMessage != DefaultAnnouncementMessage {
		t.Fatalf("unexpected fallback: %+v", item)
	}
}

func TestGet_PropagatesUnexpectedError(t *testing.T) {
	svc := NewService(&fakeRepo{getErr: apperr.ErrInternal})

	if _, err := svc.Get(context.Background()); !errors.Is(err, apperr.ErrInternal) {
		t.Fatalf("got %v, want ErrInternal", err)
	}
}

func TestUpdate_RejectsEmptyMessageWhenEnabled(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	_, err := svc.Update(context.Background(), "u-1", UpdateRequest{
		AnnouncementEnabled: true,
		AnnouncementMessage: "   ",
	})
	if !errors.Is(err, errMessageRequired) {
		t.Fatalf("got %v, want errMessageRequired", err)
	}
	if repo.lastMsg != "" {
		t.Fatalf("repository should not be called on validation failure")
	}
}

func TestUpdate_RejectsOverlongMessage(t *testing.T) {
	svc := NewService(&fakeRepo{})

	_, err := svc.Update(context.Background(), "u-1", UpdateRequest{
		AnnouncementEnabled: true,
		AnnouncementMessage: strings.Repeat("a", MaxAnnouncementLength+1),
	})
	if !errors.Is(err, errMessageTooLong) {
		t.Fatalf("got %v, want errMessageTooLong", err)
	}
}

func TestUpdate_TrimsAndPersists(t *testing.T) {
	repo := &fakeRepo{updated: Settings{AnnouncementEnabled: true, AnnouncementMessage: "Hi"}}
	svc := NewService(repo)

	item, err := svc.Update(context.Background(), "u-1", UpdateRequest{
		AnnouncementEnabled: true,
		AnnouncementMessage: "  Hi  ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastActor != "u-1" || repo.lastMsg != "Hi" || !repo.lastEnable {
		t.Fatalf("unexpected repo args: actor=%q enabled=%v msg=%q", repo.lastActor, repo.lastEnable, repo.lastMsg)
	}
	if item.AnnouncementMessage != "Hi" {
		t.Fatalf("got %q, want trimmed message", item.AnnouncementMessage)
	}
}

func TestUpdate_AllowsEmptyMessageWhenDisabled(t *testing.T) {
	repo := &fakeRepo{updated: Settings{AnnouncementEnabled: false}}
	svc := NewService(repo)

	if _, err := svc.Update(context.Background(), "u-1", UpdateRequest{
		AnnouncementEnabled: false,
		AnnouncementMessage: "",
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
