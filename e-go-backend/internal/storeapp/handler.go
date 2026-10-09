package storeapp

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/payment"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

// maxWebhookBodyBytes bounds a provider callback body.
const maxWebhookBodyBytes = 1 << 20

// isAllowedImage sniffs the bytes so only real images are uploaded.
func isAllowedImage(data []byte) bool {
	switch http.DetectContentType(data) {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}

// Handler translates applicant HTTP requests into service calls.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) identity(r *http.Request) (auth.Identity, bool) {
	return auth.IdentityFromContext(r.Context())
}

// Submit handles POST /api/v1/store-applications.
func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	id, ok := h.identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	var req SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	detail, err := h.svc.Submit(r.Context(), id.UserID, req)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, detail, "Your store application was submitted")
}

// MyApplications handles GET /api/v1/store-applications/me.
func (h *Handler) MyApplications(w http.ResponseWriter, r *http.Request) {
	id, ok := h.identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	items, err := h.svc.MyApplications(r.Context(), id.UserID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, items, "")
}

// Detail handles GET /api/v1/store-applications/{id}.
func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	id, ok := h.identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	isAdmin := id.Role == "admin"
	detail, err := h.svc.Detail(r.Context(), id.UserID, chi.URLParam(r, "id"), isAdmin)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, detail, "")
}

// ApprovalLetter handles GET /api/v1/store-applications/{id}/approval-letter.
func (h *Handler) ApprovalLetter(w http.ResponseWriter, r *http.Request) {
	id, ok := h.identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	isAdmin := id.Role == "admin"
	doc, err := h.svc.ApprovalLetter(r.Context(), id.UserID, chi.URLParam(r, "id"), isAdmin)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(doc)
}

// InitiatePayment handles POST /api/v1/store-applications/{id}/payments.
func (h *Handler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	id, ok := h.identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	out, err := h.svc.InitiatePayment(r.Context(), id.UserID, chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, out, "")
}

// Payments handles GET /api/v1/store-applications/{id}/payments.
func (h *Handler) Payments(w http.ResponseWriter, r *http.Request) {
	id, ok := h.identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	items, err := h.svc.Payments(r.Context(), id.UserID, chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, items, "")
}

// MyStores handles GET /api/v1/my-stores.
func (h *Handler) MyStores(w http.ResponseWriter, r *http.Request) {
	id, ok := h.identity(r)
	if !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	items, err := h.svc.MyStores(r.Context(), id.UserID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, items, "")
}

// UploadImage handles POST /api/v1/store-applications/images (multipart, "image").
func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.identity(r); !ok {
		response.Error(w, r, apperr.ErrUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxApplicationImageBytes+1024)
	if err := r.ParseMultipartForm(maxApplicationImageBytes); err != nil {
		response.Error(w, r, errInvalidImage)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		response.Error(w, r, errInvalidImage)
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, maxApplicationImageBytes+1))
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
	response.Created(w, map[string]string{"url": url}, "Image uploaded")
}

// Webhook handles POST /api/v1/payments/webhook/{provider}. It is public but
// only a payload the provider verifies can mark a payment paid.
func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxWebhookBodyBytes)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, r, errInvalidBody.WithCause(err))
		return
	}

	verify := payment.VerifyRequest{
		Signature: r.Header.Get("X-Payment-Signature"),
		RawBody:   raw,
	}
	// Reference/provider transaction id may arrive as a header or in the body;
	// the provider implementation decides what it trusts.
	verify.Reference = r.Header.Get("X-Payment-Reference")
	verify.ProviderTransactionID = r.Header.Get("X-Payment-Transaction-Id")

	item, err := h.svc.HandleWebhook(r.Context(), chi.URLParam(r, "provider"), verify)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, item, "Payment verified")
}
