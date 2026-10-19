package settings

import (
	"context"
)

// fakeRepo is a configurable in-memory SettingsRepository.
type fakeRepo struct {
	get    Settings
	getErr error

	updated    Settings
	updateErr  error
	lastActor  string
	lastEnable bool
	lastMsg    string
}

func (f *fakeRepo) Get(context.Context) (Settings, error) {
	if f.getErr != nil {
		return Settings{}, f.getErr
	}
	return f.get, nil
}

func (f *fakeRepo) Update(_ context.Context, actorID string, enabled bool, message string) (Settings, error) {
	f.lastActor, f.lastEnable, f.lastMsg = actorID, enabled, message
	if f.updateErr != nil {
		return Settings{}, f.updateErr
	}
	return f.updated, nil
}
