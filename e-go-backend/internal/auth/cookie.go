package auth

import (
	"net/http"
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

// setAuthCookies writes both auth cookies. They are HttpOnly so JavaScript can
// never read them, SameSite=Lax to blunt cross-site request forgery, and
// Secure in production (COOKIE_SECURE).
func setAuthCookies(
	w http.ResponseWriter,
	cfg config.AuthConfig,
	accessToken string,
	accessTTL time.Duration,
	refreshToken string,
	refreshTTL time.Duration,
) {
	http.SetCookie(w, &http.Cookie{
		Name:     accessCookieName,
		Value:    accessToken,
		Path:     "/",
		Domain:   cfg.CookieDomain,
		MaxAge:   int(accessTTL.Seconds()),
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken,
		Path:     refreshCookiePath,
		Domain:   cfg.CookieDomain,
		MaxAge:   int(refreshTTL.Seconds()),
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

// clearAuthCookies expires both cookies so the browser drops them.
func clearAuthCookies(w http.ResponseWriter, cfg config.AuthConfig) {
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
			SameSite: http.SameSiteLaxMode,
		})
	}
}
