// Package cart owns the server-side shopping cart. A cart is either bound to a
// user (after sign-in) or to an opaque guest session cookie, and holds lines
// referencing concrete product variants. Prices and stock are always resolved
// from the database — never trusted from the client.
package cart

import (
	"net/http"
	"time"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

const (
	// CookieName is the HttpOnly cookie holding a guest cart's opaque session id.
	CookieName = "cart_session"
	// SessionTTL is how long a guest cart cookie lives. The cart row itself is
	// not time-limited, so an active guest keeps their cart across visits.
	SessionTTL = 30 * 24 * time.Hour
	// MaxLineQuantity caps a single line, mirroring the storefront stepper.
	MaxLineQuantity = 99
)

// cart_status values (see migrations/002_enums.sql).
const (
	StatusActive    = "active"
	StatusConverted = "converted"
)

var (
	errInvalidBody = apperr.New(http.StatusBadRequest, "BAD_REQUEST", "Request body must be valid JSON")

	errInvalidQuantity = apperr.New(
		http.StatusUnprocessableEntity,
		"VALIDATION_ERROR",
		"Quantity must be a whole number between 1 and 99",
	)
	errProductUnavailable = apperr.New(
		http.StatusConflict,
		"PRODUCT_UNAVAILABLE",
		"A product in your cart is no longer available",
	)
	errInsufficientStock = apperr.New(
		http.StatusConflict,
		"INSUFFICIENT_STOCK",
		"There is not enough stock for this item",
	)
	errItemNotFound = apperr.New(http.StatusNotFound, "CART_ITEM_NOT_FOUND", "Cart item not found")
)

// Cart is a resolved cart: either a user cart or a guest (session) cart.
type Cart struct {
	ID        string
	UserID    string
	SessionID string
	Currency  string
}

// VariantResolution is a concrete variant resolved from a product (+ optional
// variant id), with the authoritative price and available stock.
type VariantResolution struct {
	ProductID string
	VariantID string
	Name      string
	UnitPrice float64
	Available int64
}

// ItemRecord is one persisted cart line joined with its display data.
type ItemRecord struct {
	ID           string
	ProductID    string
	VariantID    string
	Name         string
	Slug         string
	ImageURL     string
	ImageAlt     string
	VariantLabel string
	UnitPrice    float64
	Quantity     int
	Available    int64
}

// Resolution tells the handler how to reconcile the guest session cookie after
// a request: issue a new one, clear a merged one, or leave it untouched.
type Resolution struct {
	// NewSession, when non-empty, is a freshly minted guest session id that must
	// be written to the cart cookie.
	NewSession string
	// ClearSession, when true, means the guest cookie should be expired because
	// the cart it referenced was merged into the user's cart.
	ClearSession bool
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
