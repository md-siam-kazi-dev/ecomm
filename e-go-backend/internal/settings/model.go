// Package settings owns the admin-editable storefront settings. The values are
// changed by admins at runtime and read by the public storefront (currently the
// announcement bar above the header). It is the single source of truth for this
// configuration.
package settings

import (
	"net/http"
	"time"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// SingletonRowID is the fixed id of the one settings row. The table carries a
// CHECK constraint enforcing this value.
const SingletonRowID = 1

// MaxAnnouncementLength bounds the announcement message so an admin cannot
// store an unbounded string that would break the storefront layout.
const MaxAnnouncementLength = 280

// DefaultAnnouncementMessage is the fallback used when no row exists yet. It
// mirrors the seed in migrations/021_site_settings.sql.
const DefaultAnnouncementMessage = "Free shipping on orders over $75 · 30-day easy returns"

var (
	errInvalidBody = apperr.New(http.StatusBadRequest, "BAD_REQUEST", "Request body must be valid JSON")

	errMessageRequired = apperr.New(
		http.StatusUnprocessableEntity,
		"VALIDATION_ERROR",
		"Announcement message is required when the bar is enabled",
	)
	errMessageTooLong = apperr.New(
		http.StatusUnprocessableEntity,
		"VALIDATION_ERROR",
		"Announcement message must be 280 characters or fewer",
	)
)

// Settings is the storefront configuration editable from the admin panel.
type Settings struct {
	AnnouncementEnabled bool
	AnnouncementMessage string
	UpdatedAt           time.Time
}
