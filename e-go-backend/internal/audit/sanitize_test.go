package audit

import "testing"

func TestSanitizeRedactsSensitiveKeys(t *testing.T) {
	in := map[string]any{
		"email":         "a@b.com",
		"password":      "hunter2",
		"access_token":  "abc",
		"refreshToken":  "xyz",
		"apiKey":        "k",
		"card_number":   "4111111111111111",
		"cvv":           "123",
		"authorization": "Bearer x",
		"nested": map[string]any{
			"client_secret": "shh",
			"name":          "ok",
		},
	}

	got := Sanitize(in)

	for _, key := range []string{"password", "access_token", "refreshToken", "apiKey", "card_number", "cvv", "authorization"} {
		if got[key] != redacted {
			t.Errorf("%s not redacted: %v", key, got[key])
		}
	}
	if got["email"] != "a@b.com" {
		t.Errorf("email should be preserved, got %v", got["email"])
	}
	nested := got["nested"].(map[string]any)
	if nested["client_secret"] != redacted {
		t.Errorf("nested client_secret not redacted: %v", nested["client_secret"])
	}
	if nested["name"] != "ok" {
		t.Errorf("nested name should be preserved, got %v", nested["name"])
	}

	// The input must never be mutated.
	if in["password"] != "hunter2" {
		t.Errorf("Sanitize mutated the input: %v", in["password"])
	}
}

func TestSanitizeNilAndTruncation(t *testing.T) {
	if Sanitize(nil) != nil {
		t.Errorf("nil input should yield nil")
	}

	long := make([]byte, maxStringBytes+10)
	for i := range long {
		long[i] = 'x'
	}
	got := Sanitize(map[string]any{"blob": string(long)})
	s := got["blob"].(string)
	if len(s) <= maxStringBytes {
		t.Errorf("long string not truncated: len=%d", len(s))
	}
}

func TestSanitizeSlices(t *testing.T) {
	got := Sanitize(map[string]any{
		"items": []any{
			map[string]any{"secret": "a", "keep": "b"},
			"plain",
		},
	})
	items := got["items"].([]any)
	first := items[0].(map[string]any)
	if first["secret"] != redacted || first["keep"] != "b" {
		t.Errorf("slice element not sanitized: %v", first)
	}
}
