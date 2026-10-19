package follow

import "context"

// Service owns the store-follow business logic (Handler → Service → Repository).
type Service struct {
	repo FollowRepository
}

func NewService(repo FollowRepository) *Service { return &Service{repo: repo} }

// Status returns the caller's follow state and the store's follower count. An
// empty userID (guest) reports following=false but still returns the count.
func (s *Service) Status(ctx context.Context, slug, userID string) (StatusResponse, error) {
	storeID, err := s.repo.ActiveStoreID(ctx, slug)
	if err != nil {
		return StatusResponse{}, err
	}
	return s.status(ctx, storeID, userID)
}

// Follow subscribes the caller to the store (idempotent).
func (s *Service) Follow(ctx context.Context, slug, userID string) (StatusResponse, error) {
	storeID, err := s.repo.ActiveStoreID(ctx, slug)
	if err != nil {
		return StatusResponse{}, err
	}
	if err := s.repo.Follow(ctx, storeID, userID); err != nil {
		return StatusResponse{}, err
	}
	return s.status(ctx, storeID, userID)
}

// Unfollow removes the caller's subscription (idempotent).
func (s *Service) Unfollow(ctx context.Context, slug, userID string) (StatusResponse, error) {
	storeID, err := s.repo.ActiveStoreID(ctx, slug)
	if err != nil {
		return StatusResponse{}, err
	}
	if err := s.repo.Unfollow(ctx, storeID, userID); err != nil {
		return StatusResponse{}, err
	}
	return s.status(ctx, storeID, userID)
}

func (s *Service) status(ctx context.Context, storeID, userID string) (StatusResponse, error) {
	var following bool
	if userID != "" {
		value, err := s.repo.Following(ctx, storeID, userID)
		if err != nil {
			return StatusResponse{}, err
		}
		following = value
	}
	count, err := s.repo.Count(ctx, storeID)
	if err != nil {
		return StatusResponse{}, err
	}
	return StatusResponse{Following: following, FollowerCount: count}, nil
}
