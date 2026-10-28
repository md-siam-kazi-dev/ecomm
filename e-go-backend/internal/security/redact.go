package security

import (
	"github.com/md-siam-kazi/e-go-backend/internal/audit"
)

// sanitizeMetadata redacts sensitive keys (password, token, secret, cookie,
// card, cvv, ...) and truncates oversized values.
//
// It delegates to the audit trail's proven sanitizer so the codebase keeps a
// single denylist rather than two that can drift apart. Bodies, headers and
// tokens are never captured in the first place; this is the second line of
// defence for anything a caller puts in Metadata.
func sanitizeMetadata(metadata map[string]any) map[string]any {
	return audit.Sanitize(metadata)
}
