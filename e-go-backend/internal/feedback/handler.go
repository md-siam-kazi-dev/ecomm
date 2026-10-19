package feedback

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

// Handler translates authenticated feedback HTTP requests into service calls.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func identity(r *http.Request) (auth.Identity, bool) {
	return auth.IdentityFromContext(r.Context())
}

// isAllowedImage sniffs the bytes so only real images are uploaded.
func isAllowedImage(data []byte) bool {
	switch http.DetectContentType(data) {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}

// CreateReview handles POST /api/v1/reviews.
func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	review, err := h.svc.CreateReview(r.Context(), id.UserID, req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, review, "Review submitted")
}

// UpdateReview handles PUT /api/v1/reviews/{id}.
func (h *Handler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req UpdateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	review, err := h.svc.UpdateReview(r.Context(), id.UserID, chi.URLParam(r, "id"), req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, review, "Review updated")
}

// CreateQuestion handles POST /api/v1/products/{slug}/questions.
func (h *Handler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	id, ok := identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req CreateQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	question, err := h.svc.CreateQuestion(r.Context(), id.UserID, chi.URLParam(r, "slug"), req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, question, "Question submitted")
}

// UploadImage handles POST /api/v1/reviews/images (multipart, field "image").
func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if _, ok := identity(r); !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxReviewImageBytes+1024)
	if err := r.ParseMultipartForm(maxReviewImageBytes); err != nil {
		response.Error(w, r, errInvalidImage)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		response.Error(w, r, errInvalidImage)
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, maxReviewImageBytes+1))
	if err != nil {
		response.Error(w, r, apperr.ErrInternal.WithCause(err))
		return
	}
	if !isAllowedImage(data) {
		response.Error(w, r, errInvalidImage)
		return
	}

	url, err := h.svc.UploadImage(r.Context(), header.Filename, data)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, ImageUploadResponse{URL: url}, "Image uploaded")
}
