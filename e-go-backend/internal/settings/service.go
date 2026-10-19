package settings

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// Service owns the storefront settings business logic
// (Handler → Service → Repository).
type Service struct {
	repo SettingsRepository
}

func NewService(repo SettingsRepository) *Service { return &Service{repo: repo} }

// Get returns the current settings, falling back to the seeded defaults when
// the row is missing so the storefront never fails to render.
func (s *Service) Get(ctx context.Context) (Item, error) {
	record, err := s.repo.Get(ctx)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return toItem(Settings{
				AnnouncementEnabled: true,
				AnnouncementMessage: DefaultAnnouncementMessage,
				UpdatedAt:           time.Now().UTC(),
			}), nil
		}
		return Item{}, err
	}
	return toItem(record), nil
}

// Update validates the payload and persists the new settings.
func (s *Service) Update(ctx context.Context, actorID string, req UpdateRequest) (Item, error) {
	message := strings.TrimSpace(req.AnnouncementMessage)

	if req.AnnouncementEnabled && message == "" {
		return Item{}, errMessageRequired
	}
	if utf8.RuneCountInString(message) > MaxAnnouncementLength {
		return Item{}, errMessageTooLong
	}

	record, err := s.repo.Update(ctx, actorID, req.AnnouncementEnabled, message)
	if err != nil {
		return Item{}, err
	}
	return toItem(record), nil
}
