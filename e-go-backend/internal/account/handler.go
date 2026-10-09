package account

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

var errInvalidBody = apperr.New(http.StatusBadRequest, "BAD_REQUEST", "Request body must be valid JSON")

// isAllowedImage sniffs the bytes so only real images are uploaded.
func isAllowedImage(data []byte) bool {
	switch http.DetectContentType(data) {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}

// Handler translates address HTTP requests into service calls. Callers are
// authenticated (auth.Authenticate) so the user id comes from the context.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// GetProfile handles GET /api/v1/account/profile.
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	profile, err := h.svc.GetProfile(r.Context(), userID.UserID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, profile, "")
}

// UpdateProfile handles PUT /api/v1/account/profile.
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req ProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	profile, err := h.svc.UpdateProfile(r.Context(), userID.UserID, req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, profile, "Profile updated")
}

// UploadAvatar handles POST /api/v1/account/avatar (multipart, field "image").
// The image is hosted and only its URL is stored.
func (h *Handler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarBytes+1024)
	if err := r.ParseMultipartForm(maxAvatarBytes); err != nil {
		response.Error(w, r, errInvalidImage)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		response.Error(w, r, errInvalidImage)
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, maxAvatarBytes+1))
	if err != nil {
		response.Error(w, r, apperr.ErrInternal.WithCause(err))
		return
	}
	if !isAllowedImage(data) {
		response.Error(w, r, errInvalidImage)
		return
	}

	profile, err := h.svc.UploadAvatar(r.Context(), userID.UserID, header.Filename, data)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, profile, "Profile picture updated")
}

// List handles GET /api/v1/account/addresses.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	addresses, err := h.svc.List(r.Context(), userID.UserID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, addresses, "")
}

// Create handles POST /api/v1/account/addresses.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req AddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	address, err := h.svc.Create(r.Context(), userID.UserID, req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, address, "Address saved")
}

// Update handles PUT /api/v1/account/addresses/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req AddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	address, err := h.svc.Update(r.Context(), userID.UserID, chi.URLParam(r, "id"), req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, address, "Address updated")
}

// Delete handles DELETE /api/v1/account/addresses/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	if err := h.svc.Delete(r.Context(), userID.UserID, chi.URLParam(r, "id")); err != nil {
		response.Error(w, r, err)
		return
	}
	response.NoContent(w)
}

// SetDefault handles POST /api/v1/account/addresses/{id}/default.
func (h *Handler) SetDefault(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	if err := h.svc.SetDefault(r.Context(), userID.UserID, chi.URLParam(r, "id")); err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, nil, "Default address updated")
}
