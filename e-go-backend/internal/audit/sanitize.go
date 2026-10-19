package audit

import "strings"

// redacted replaces the value of any field whose key looks sensitive.
const redacted = "[REDACTED]"

// maxStringBytes caps individual string values so a large blob (e.g. an
// embedded base64 image) cannot bloat an audit row.
const maxStringBytes = 4096

// maxDepth bounds recursion so a pathological structure cannot cause a stack
// overflow.
const maxDepth = 6

// sensitiveKeys is a denylist of key names that must never be persisted.
// Matching is case-insensitive and substring-based on the normalised key so
// "access_token", "refreshToken" and "user_password" are all caught.
var sensitiveKeys = []string{
	"password",
	"passwd",
	"secret",
	"token",
	"authorization",
	"cookie",
	"api_key",
	"apikey",
	"access_key",
	"private_key",
	"credit_card",
	"card_number",
	"number", // card number is the only "number" we ever pass; be conservative
	"cvv",
	"cvc",
	"ssn",
	"otp",
	"pin",
	"webhook",
	"signature",
	"client_secret",
}

// Sanitize returns a copy of m with sensitive values redacted and oversized
// values truncated. It never mutates the input. A nil input yields nil so the
// column is stored as SQL NULL rather than an empty object.
func Sanitize(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	return sanitizeMap(m, 0)
}

func sanitizeMap(m map[string]any, depth int) map[string]any {
	if depth > maxDepth {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		if isSensitiveKey(k) {
			out[k] = redacted
			continue
		}
		out[k] = sanitizeValue(v, depth+1)
	}
	return out
}

func sanitizeValue(v any, depth int) any {
	switch val := v.(type) {
	case map[string]any:
		return sanitizeMap(val, depth)
	case []any:
		if depth > maxDepth {
			return nil
		}
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = sanitizeValue(item, depth+1)
		}
		return out
	case string:
		if len(val) > maxStringBytes {
			return val[:maxStringBytes] + "…[truncated]"
		}
		return val
	default:
		return v
	}
}

// isSensitiveKey normalises the key (lowercase, strip separators) and reports
// whether it contains any sensitive marker. It deliberately errs toward
// redaction: a benign field like "token_count" is still redacted.
func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(key)
	normalized = strings.NewReplacer("_", "", "-", "", " ", "").Replace(normalized)
	for _, marker := range sensitiveKeys {
		if strings.Contains(normalized, strings.ReplaceAll(marker, "_", "")) {
			return true
		}
	}
	return false
}
