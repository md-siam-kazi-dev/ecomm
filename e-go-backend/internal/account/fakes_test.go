package account

import (
	"context"
	"time"
)

type fakeRepo struct {
	addresses []Address
	lastID    string
	lastUser  string
	lastParam AddressParams
	deleted   bool
	defaulted bool

	profileFullName string
	profilePhone    string
	avatarURL       string
	profileErr      error
}

func (f *fakeRepo) ListAddresses(_ context.Context, userID string) ([]Address, error) {
	f.lastUser = userID
	return f.addresses, nil
}

func (f *fakeRepo) CreateAddress(_ context.Context, userID string, p AddressParams) (Address, error) {
	f.lastUser, f.lastParam = userID, p
	return Address{
		ID: "addr-1", FullName: p.FullName, Phone: p.Phone, Division: p.Division,
		District: p.District, Upazila: p.Upazila, VillageTown: p.VillageTown,
		Clarification: p.Clarification, IsDefault: len(f.addresses) == 0 || p.IsDefault,
		CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}, nil
}

func (f *fakeRepo) UpdateAddress(_ context.Context, userID, id string, p AddressParams) (Address, error) {
	f.lastUser, f.lastID, f.lastParam = userID, id, p
	return Address{ID: id, FullName: p.FullName, Phone: p.Phone, Division: p.Division,
		District: p.District, Upazila: p.Upazila, VillageTown: p.VillageTown,
		Clarification: p.Clarification, IsDefault: p.IsDefault,
		CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}, nil
}

func (f *fakeRepo) DeleteAddress(_ context.Context, userID, id string) error {
	f.lastUser, f.lastID, f.deleted = userID, id, true
	return nil
}

func (f *fakeRepo) SetDefault(_ context.Context, userID, id string) error {
	f.lastUser, f.lastID, f.defaulted = userID, id, true
	return nil
}

func (f *fakeRepo) GetProfile(_ context.Context, userID string) (Profile, error) {
	f.lastUser = userID
	if f.profileErr != nil {
		return Profile{}, f.profileErr
	}
	fullName := f.profileFullName
	if fullName == "" {
		fullName = "Siam Kazi"
	}
	return Profile{FullName: fullName, Email: "u@example.com", Phone: f.profilePhone, AvatarURL: f.avatarURL}, nil
}

func (f *fakeRepo) UpdateProfile(_ context.Context, userID, fullName, phone string) error {
	f.lastUser, f.profileFullName, f.profilePhone = userID, fullName, phone
	return nil
}

func (f *fakeRepo) UpdateAvatar(_ context.Context, userID, avatarURL string) error {
	f.lastUser, f.avatarURL = userID, avatarURL
	return nil
}

type fakeUploader struct {
	url    string
	err    error
	called bool
}

func (f *fakeUploader) Upload(_ context.Context, _ string, _ []byte) (string, error) {
	f.called = true
	if f.err != nil {
		return "", f.err
	}
	if f.url == "" {
		return "https://i.ibb.co/hosted/avatar.png", nil
	}
	return f.url, nil
}

func validRequest() AddressRequest {
	return AddressRequest{
		FullName:    "Siam Kazi",
		Phone:       "+8801700000000",
		Division:    "Dhaka",
		District:    "Gazipur",
		Upazila:     "Kaliakair",
		VillageTown: "Mouchak",
	}
}
