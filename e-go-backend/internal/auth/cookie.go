package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/md-siam-kazi/e-go-backend/internal/config"
)

const (
	accessCookieName  = "access_token"
	refreshCookieName = "refresh_token"
	// The refresh cookie is only ever sent to the auth endpoints, limiting its
	// exposure on every other request.
	refreshCookiePath = "/api/v1/auth"
)

// sameSiteMode maps the COOKIE_SAMESITE setting to the http.SameSite value.
// Unknown or empty values fall back to Lax.
func sameSiteMode(v string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

// setAuthCookies writes both auth cookies. They are HttpOnly so JavaScript can
// never read them and Secure in production (COOKIE_SECURE). SameSite defaults
// to Lax (CSRF defence) but is configurable via COOKIE_SAMESITE — set "none"
// when the frontend lives on a different site than the API.
func setAuthCookies(
	w http.ResponseWriter,
	cfg config.AuthConfig,
	accessToken string,
	accessTTL time.Duration,
	refreshToken string,
	refreshTTL time.Duration,
) {
	sameSite := sameSiteMode(cfg.CookieSameSite)
	http.SetCookie(w, &http.Cookie{
		Name:     accessCookieName,
		Value:    accessToken,
		Path:     "/",
		Domain:   cfg.CookieDomain,
		MaxAge:   int(accessTTL.Seconds()),
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: sameSite,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken,
		Path:     refreshCookiePath,
		Domain:   cfg.CookieDomain,
		MaxAge:   int(refreshTTL.Seconds()),
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: sameSite,
	})
}

// clearAuthCookies expires both cookies so the browser drops them.
func clearAuthCookies(w http.ResponseWriter, cfg config.AuthConfig) {
	sameSite := sameSiteMode(cfg.CookieSameSite)
	for _, cookie := range []struct{ name, path string }{
		{accessCookieName, "/"},
		{refreshCookieName, refreshCookiePath},
	} {
		http.SetCookie(w, &http.Cookie{
			Name:     cookie.name,
			Value:    "",
			Path:     cookie.path,
			Domain:   cfg.CookieDomain,
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   cfg.CookieSecure,
			SameSite: sameSite,
		})
	}
}
