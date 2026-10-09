package account

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/imgbb"
)

func TestCreate_Validation(t *testing.T) {
	cases := map[string]func(*AddressRequest){
		"short name":         func(r *AddressRequest) { r.FullName = "X" },
		"short phone":        func(r *AddressRequest) { r.Phone = "123" },
		"missing division":   func(r *AddressRequest) { r.Division = "  " },
		"missing district":   func(r *AddressRequest) { r.District = "" },
		"missing upazila":    func(r *AddressRequest) { r.Upazila = "" },
		"missing village":    func(r *AddressRequest) { r.VillageTown = "" },
		"long clarification": func(r *AddressRequest) { r.Clarification = strings.Repeat("x", maxClarificationLen+1) },
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			req := validRequest()
			mutate(&req)

			_, err := NewService(&fakeRepo{}, &fakeUploader{}).Create(context.Background(), "u-1", req)
			appErr, ok := apperr.As(err)
			if !ok || appErr.Code != apperr.ErrValidation.Code {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestCreate_TrimsAndPassesParams(t *testing.T) {
	repo := &fakeRepo{}
	req := validRequest()
	req.FullName = "  Siam Kazi  "
	req.VillageTown = "  Mouchak  "

	resp, err := NewService(repo, &fakeUploader{}).Create(context.Background(), "u-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastParam.FullName != "Siam Kazi" || repo.lastParam.VillageTown != "Mouchak" {
		t.Errorf("params not trimmed: %+v", repo.lastParam)
	}
	if resp.Division != "Dhaka" || resp.District != "Gazipur" {
		t.Errorf("unexpected response: %+v", resp)
	}
	if !resp.IsDefault {
		t.Error("first address should be the default")
	}
}

func TestList_MapsAddresses(t *testing.T) {
	repo := &fakeRepo{addresses: []Address{
		{ID: "a-1", FullName: "Siam", Division: "Dhaka", District: "Gazipur", Upazila: "Kaliakair", VillageTown: "Mouchak"},
	}}

	got, err := NewService(repo, &fakeUploader{}).List(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].VillageTown != "Mouchak" {
		t.Fatalf("unexpected list: %+v", got)
	}
}

func TestDeleteAndSetDefault_PassIdentity(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, &fakeUploader{})

	if err := svc.Delete(context.Background(), "u-9", "a-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.deleted || repo.lastUser != "u-9" || repo.lastID != "a-1" {
		t.Errorf("delete not forwarded: %+v", repo)
	}

	if err := svc.SetDefault(context.Background(), "u-9", "a-2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.defaulted || repo.lastID != "a-2" {
		t.Errorf("set default not forwarded: %+v", repo)
	}
}

// ---- profile --------------------------------------------------------------

func TestUpdateProfile_Validation(t *testing.T) {
	cases := map[string]ProfileRequest{
		"short name": {FullName: "X", Phone: "+8801700000000"},
		"long phone": {FullName: "Siam Kazi", Phone: strings.Repeat("1", 31)},
	}
	for name, req := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := NewService(&fakeRepo{}, &fakeUploader{}).UpdateProfile(context.Background(), "u-1", req)
			appErr, ok := apperr.As(err)
			if !ok || appErr.Code != apperr.ErrValidation.Code {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestUpdateProfile_TrimsAndPersists(t *testing.T) {
	repo := &fakeRepo{}

	resp, err := NewService(repo, &fakeUploader{}).UpdateProfile(context.Background(), "u-1", ProfileRequest{
		FullName: "  Siam Kazi  ", Phone: "  +8801700000000  ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.profileFullName != "Siam Kazi" || repo.profilePhone != "+8801700000000" {
		t.Errorf("profile not trimmed/persisted: %q %q", repo.profileFullName, repo.profilePhone)
	}
	if resp.FullName != "Siam Kazi" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestUploadAvatar_StoresHostedURL(t *testing.T) {
	repo := &fakeRepo{}
	uploader := &fakeUploader{url: "https://i.ibb.co/abc/pic.png"}

	resp, err := NewService(repo, uploader).UploadAvatar(context.Background(), "u-1", "pic.png", []byte("bytes"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !uploader.called {
		t.Error("uploader should have been called")
	}
	if repo.avatarURL != "https://i.ibb.co/abc/pic.png" {
		t.Errorf("avatar url not stored: %q", repo.avatarURL)
	}
	if resp.AvatarURL != "https://i.ibb.co/abc/pic.png" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestUploadAvatar_Errors(t *testing.T) {
	t.Run("empty data", func(t *testing.T) {
		_, err := NewService(&fakeRepo{}, &fakeUploader{}).UploadAvatar(context.Background(), "u-1", "pic.png", nil)
		appErr, ok := apperr.As(err)
		if !ok || appErr.Code != "INVALID_IMAGE" {
			t.Fatalf("expected INVALID_IMAGE, got %v", err)
		}
	})

	t.Run("hosting not configured", func(t *testing.T) {
		uploader := &fakeUploader{err: imgbb.ErrNotConfigured}
		_, err := NewService(&fakeRepo{}, uploader).UploadAvatar(context.Background(), "u-1", "pic.png", []byte("bytes"))
		appErr, ok := apperr.As(err)
		if !ok || appErr.Code != "IMAGE_HOSTING_UNAVAILABLE" {
			t.Fatalf("expected IMAGE_HOSTING_UNAVAILABLE, got %v", err)
		}
	})

	t.Run("uploader failure", func(t *testing.T) {
		uploader := &fakeUploader{err: errors.New("boom")}
		_, err := NewService(&fakeRepo{}, uploader).UploadAvatar(context.Background(), "u-1", "pic.png", []byte("bytes"))
		if _, ok := apperr.As(err); !ok {
			t.Fatalf("expected app error, got %v", err)
		}
	})
}
