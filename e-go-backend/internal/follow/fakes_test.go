package follow

import "context"

// fakeRepo is a configurable in-memory FollowRepository.
type fakeRepo struct {
	storeID  string
	storeErr error

	following     bool
	followerCount int64
	followErr     error
	unfollowErr   error

	lastFollowed   string
	lastUnfollowed string
}

func (f *fakeRepo) ActiveStoreID(context.Context, string) (string, error) {
	if f.storeErr != nil {
		return "", f.storeErr
	}
	return f.storeID, nil
}

func (f *fakeRepo) Follow(_ context.Context, _, userID string) error {
	f.lastFollowed = userID
	if f.followErr != nil {
		return f.followErr
	}
	f.following = true
	f.followerCount++
	return nil
}

func (f *fakeRepo) Unfollow(_ context.Context, _, userID string) error {
	f.lastUnfollowed = userID
	if f.unfollowErr != nil {
		return f.unfollowErr
	}
	f.following = false
	if f.followerCount > 0 {
		f.followerCount--
	}
	return nil
}

func (f *fakeRepo) Following(context.Context, string, string) (bool, error) {
	return f.following, nil
}

func (f *fakeRepo) Count(context.Context, string) (int64, error) {
	return f.followerCount, nil
}
