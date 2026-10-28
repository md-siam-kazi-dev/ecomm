package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
)

func mustPrefixes(t *testing.T, values ...string) []netip.Prefix {
	t.Helper()
	out := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			t.Fatalf("parse prefix %q: %v", value, err)
		}
		out = append(out, prefix)
	}
	return out
}

func TestResolveClientIP(t *testing.T) {
	tests := []struct {
		name    string
		trusted []string
		remote  string
		xff     string
		realIP  string
		want    string
	}{
		{
			name:   "untrusted peer ignores spoofed X-Forwarded-For",
			remote: "198.51.100.7:5555",
			xff:    "1.2.3.4",
			want:   "198.51.100.7",
		},
		{
			name:   "untrusted peer ignores X-Real-IP",
			remote: "198.51.100.7:5555",
			realIP: "1.2.3.4",
			want:   "198.51.100.7",
		},
		{
			name:    "trusted peer honours X-Forwarded-For",
			trusted: []string{"10.0.0.0/8"},
			remote:  "10.0.0.5:443",
			xff:     "203.0.113.9",
			want:    "203.0.113.9",
		},
		{
			name:    "trusted peer walks X-Forwarded-For right to left",
			trusted: []string{"10.0.0.0/8"},
			remote:  "10.0.0.5:443",
			xff:     "1.2.3.4, 10.0.0.9",
			want:    "1.2.3.4",
		},
		{
			name:    "forged left-most hop is ignored",
			trusted: []string{"10.0.0.0/8"},
			remote:  "10.0.0.5:443",
			xff:     "6.6.6.6, 1.2.3.4, 10.0.0.9",
			want:    "1.2.3.4",
		},
		{
			name:    "all hops trusted falls back to X-Real-IP",
			trusted: []string{"10.0.0.0/8"},
			remote:  "10.0.0.5:443",
			xff:     "10.0.0.9, 10.0.0.10",
			realIP:  "192.0.2.44",
			want:    "192.0.2.44",
		},
		{
			name:    "all hops trusted and no X-Real-IP returns the peer",
			trusted: []string{"10.0.0.0/8"},
			remote:  "10.0.0.5:443",
			xff:     "10.0.0.9",
			want:    "10.0.0.5",
		},
		{
			name:    "trusted peer with no headers returns the peer",
			trusted: []string{"10.0.0.0/8"},
			remote:  "10.0.0.5:443",
			want:    "10.0.0.5",
		},
		{
			name:    "ipv6 peer and ipv6 trusted prefix",
			trusted: []string{"2001:db8::/32"},
			remote:  "[2001:db8::1]:443",
			xff:     "2001:db9:aaaa::5",
			want:    "2001:db9:aaaa::5",
		},
		{
			name:    "ipv6 hop inside the trusted prefix is skipped",
			trusted: []string{"2001:db8::/32"},
			remote:  "[2001:db8::1]:443",
			xff:     "2001:db8:aaaa::5",
			want:    "2001:db8::1",
		},
		{
			name:   "ipv6 peer untrusted",
			remote: "[2001:db8::1]:443",
			xff:    "2001:db8:aaaa::5",
			want:   "2001:db8::1",
		},
		{
			name:    "ipv4-mapped peer matches an ipv4 trusted prefix",
			trusted: []string{"10.0.0.0/8"},
			remote:  "[::ffff:10.0.0.5]:443",
			xff:     "203.0.113.9",
			want:    "203.0.113.9",
		},
		{
			name:   "garbage remote address resolves to empty",
			remote: "not-an-address",
			xff:    "1.2.3.4",
			want:   "",
		},
		{
			name:    "garbage forwarded hop is skipped",
			trusted: []string{"10.0.0.0/8"},
			remote:  "10.0.0.5:443",
			xff:     "203.0.113.9, not-an-address",
			want:    "203.0.113.9",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
			req.RemoteAddr = tc.remote
			if tc.xff != "" {
				req.Header.Set("X-Forwarded-For", tc.xff)
			}
			if tc.realIP != "" {
				req.Header.Set("X-Real-IP", tc.realIP)
			}

			got := resolveClientIP(mustPrefixes(t, tc.trusted...), req)
			if got != tc.want {
				t.Fatalf("resolveClientIP() = %q, want %q", got, tc.want)
			}

			// Outside the middleware chain there is no trusted-proxy list, so a
			// direct ClientIP call must refuse forwarded headers and report the
			// socket address. Only assert that where no prefixes were given.
			if len(tc.trusted) == 0 {
				if direct := ClientIP(req); direct != tc.want {
					t.Fatalf("ClientIP() = %q, want %q", direct, tc.want)
				}
			}
		})
	}
}

func TestResolveClientIPStoresContext(t *testing.T) {
	var fromContext string
	var clientIPInHandler string

	handler := ResolveClientIP(mustPrefixes(t, "10.0.0.0/8"))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fromContext, _ = ClientIPFromContext(r.Context())
			clientIPInHandler = ClientIP(r)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	req.RemoteAddr = "10.0.0.5:443"
	req.Header.Set("X-Forwarded-For", "203.0.113.9")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if fromContext != "203.0.113.9" {
		t.Fatalf("context IP = %q, want 203.0.113.9", fromContext)
	}
	if clientIPInHandler != "203.0.113.9" {
		t.Fatalf("ClientIP in handler = %q, want 203.0.113.9", clientIPInHandler)
	}
}

func TestRateKeyUsesResolvedIP(t *testing.T) {
	var key string

	handler := ResolveClientIP(mustPrefixes(t, "10.0.0.0/8"))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key = RateKey("auth_login")(r)
		}),
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "10.0.0.5:443"
	req.Header.Set("X-Forwarded-For", "203.0.113.9")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if key != "auth_login:203.0.113.9" {
		t.Fatalf("RateKey = %q, want auth_login:203.0.113.9", key)
	}
}
