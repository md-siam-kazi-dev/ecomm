package account

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/imgbb"
)

const (
	maxFieldLen         = 150
	maxClarificationLen = 500
	// maxAvatarBytes caps an uploaded profile picture (5 MB).
	maxAvatarBytes = 5 << 20
)

var (
	// errInvalidImage is returned for an empty or unsupported avatar upload.
	errInvalidImage = apperr.New(http.StatusBadRequest, "INVALID_IMAGE", "Attach a PNG, JPG, WebP or GIF image (max 5 MB)")
	// errImageUnavailable is returned when image hosting is not configured.
	errImageUnavailable = apperr.New(http.StatusServiceUnavailable, "IMAGE_HOSTING_UNAVAILABLE", "Image uploads are not configured")
)

// Store is the persistence boundary for the account module.
type Store interface {
	AddressRepository
	ProfileRepository
}

// ImageUploader hosts an image and returns its public URL.
type ImageUploader interface {
	Upload(ctx context.Context, filename string, data []byte) (string, error)
}

// Service owns account business logic (Handler → Service → Repository).
type Service struct {
	repo     Store
	uploader ImageUploader
}

func NewService(repo Store, uploader ImageUploader) *Service {
	return &Service{repo: repo, uploader: uploader}
}

// ---- addresses ------------------------------------------------------------

// List returns the user's saved addresses.
func (s *Service) List(ctx context.Context, userID string) ([]AddressResponse, error) {
	addresses, err := s.repo.ListAddresses(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]AddressResponse, 0, len(addresses))
	for _, a := range addresses {
		out = append(out, toResponse(a))
	}
	return out, nil
}

// Create adds an address for the user.
func (s *Service) Create(ctx context.Context, userID string, req AddressRequest) (AddressResponse, error) {
	params, err := validateAddress(req)
	if err != nil {
		return AddressResponse{}, err
	}
	addr, err := s.repo.CreateAddress(ctx, userID, params)
	if err != nil {
		return AddressResponse{}, err
	}
	return toResponse(addr), nil
}

// Update edits one of the user's addresses.
func (s *Service) Update(ctx context.Context, userID, id string, req AddressRequest) (AddressResponse, error) {
	if id == "" {
		return AddressResponse{}, errAddressNotFound
	}
	params, err := validateAddress(req)
	if err != nil {
		return AddressResponse{}, err
	}
	addr, err := s.repo.UpdateAddress(ctx, userID, id, params)
	if err != nil {
		return AddressResponse{}, err
	}
	return toResponse(addr), nil
}

// Delete removes one of the user's addresses.
func (s *Service) Delete(ctx context.Context, userID, id string) error {
	if id == "" {
		return errAddressNotFound
	}
	return s.repo.DeleteAddress(ctx, userID, id)
}

// SetDefault marks one address as the default.
func (s *Service) SetDefault(ctx context.Context, userID, id string) error {
	if id == "" {
		return errAddressNotFound
	}
	return s.repo.SetDefault(ctx, userID, id)
}

// ---- profile --------------------------------------------------------------

// GetProfile returns the user's editable profile.
func (s *Service) GetProfile(ctx context.Context, userID string) (ProfileResponse, error) {
	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return ProfileResponse{}, err
	}
	return toProfileResponse(profile), nil
}

// UpdateProfile changes the display name and phone number.
func (s *Service) UpdateProfile(ctx context.Context, userID string, req ProfileRequest) (ProfileResponse, error) {
	fullName := strings.TrimSpace(req.FullName)
	phone := strings.TrimSpace(req.Phone)

	if len(fullName) < 2 || len(fullName) > 200 {
		return ProfileResponse{}, apperr.ErrValidation
	}
	if phone != "" && (len(phone) < 6 || len(phone) > 30) {
		return ProfileResponse{}, apperr.ErrValidation
	}

	if err := s.repo.UpdateProfile(ctx, userID, fullName, phone); err != nil {
		return ProfileResponse{}, err
	}
	return s.GetProfile(ctx, userID)
}

// UploadAvatar hosts the image and stores its URL on the user's profile.
func (s *Service) UploadAvatar(ctx context.Context, userID, filename string, data []byte) (ProfileResponse, error) {
	if len(data) == 0 || len(data) > maxAvatarBytes {
		return ProfileResponse{}, errInvalidImage
	}

	url, err := s.uploader.Upload(ctx, filename, data)
	if err != nil {
		if errors.Is(err, imgbb.ErrNotConfigured) {
			return ProfileResponse{}, errImageUnavailable
		}
		return ProfileResponse{}, apperr.ErrInternal.WithCause(err)
	}

	if err := s.repo.UpdateAvatar(ctx, userID, url); err != nil {
		return ProfileResponse{}, err
	}
	return s.GetProfile(ctx, userID)
}

// ---- validation -----------------------------------------------------------

func validateAddress(req AddressRequest) (AddressParams, error) {
	params := AddressParams{
		FullName:      strings.TrimSpace(req.FullName),
		Phone:         strings.TrimSpace(req.Phone),
		Division:      strings.TrimSpace(req.Division),
		District:      strings.TrimSpace(req.District),
		Upazila:       strings.TrimSpace(req.Upazila),
		VillageTown:   strings.TrimSpace(req.VillageTown),
		Clarification: strings.TrimSpace(req.Clarification),
		IsDefault:     req.IsDefault,
	}

	if len(params.FullName) < 2 || len(params.Phone) < 6 {
		return AddressParams{}, apperr.ErrValidation
	}
	if params.Division == "" || params.District == "" || params.Upazila == "" || params.VillageTown == "" {
		return AddressParams{}, apperr.ErrValidation
	}
	if len(params.Division) > maxFieldLen || len(params.District) > maxFieldLen ||
		len(params.Upazila) > maxFieldLen || len(params.VillageTown) > maxFieldLen ||
		len(params.Phone) > 30 || len(params.FullName) > 200 {
		return AddressParams{}, apperr.ErrValidation
	}
	if len(params.Clarification) > maxClarificationLen {
		return AddressParams{}, apperr.ErrValidation
	}

	return params, nil
}

// splitName maps a display name onto first_name / last_name.
func splitName(name string) (first, last string) {
	parts := strings.Fields(name)
	switch len(parts) {
	case 0:
		return "", ""
	case 1:
		return parts[0], ""
	default:
		return parts[0], strings.Join(parts[1:], " ")
	}
}
