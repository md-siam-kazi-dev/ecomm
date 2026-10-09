// Package jwt implements a minimal, dependency-free HS256 JWT signer and
// verifier for short-lived access tokens.
//
// It intentionally supports only HS256 (HMAC-SHA256): tokens are opaque to the
// client, so no algorithm negotiation is needed and the "none"/RS256 confusion
// attacks are structurally impossible.
package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	// ErrInvalid is returned for a malformed, tampered, or wrong-type token.
	ErrInvalid = errors.New("jwt: invalid token")
	// ErrExpired is returned for a well-formed token past its expiry.
	ErrExpired = errors.New("jwt: token expired")
)

// accessTokenType is the only token type this signer issues. Refresh tokens are
// opaque random strings, not JWTs.
const accessTokenType = "access"

// Claims is the minimal access-token payload. Never put sensitive data here —
// the payload is only base64-encoded, not encrypted.
type Claims struct {
	Subject   string `json:"sub"`
	Role      string `json:"role"`
	Type      string `json:"type"`
	Issuer    string `json:"iss,omitempty"`
	Audience  string `json:"aud,omitempty"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

// Signer issues and verifies HS256 access tokens.
type Signer struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
	now      func() time.Time
}

// NewSigner builds a signer. `ttl` is the access-token lifetime.
func NewSigner(secret, issuer, audience string, ttl time.Duration) *Signer {
	return &Signer{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
		ttl:      ttl,
		now:      time.Now,
	}
}

// Sign returns a signed access token for the given user and role.
func (s *Signer) Sign(userID, role string) (string, error) {
	now := s.now()
	claims := Claims{
		Subject:   userID,
		Role:      role,
		Type:      accessTokenType,
		Issuer:    s.issuer,
		Audience:  s.audience,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(s.ttl).Unix(),
	}

	header, err := encodeSegment(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := encodeSegment(claims)
	if err != nil {
		return "", err
	}

	signingInput := header + "." + payload
	return signingInput + "." + s.sign(signingInput), nil
}

// Parse verifies the signature and the required claims and returns the payload.
func (s *Signer) Parse(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalid
	}

	header, err := decodeSegment(parts[0])
	if err != nil {
		return Claims{}, ErrInvalid
	}
	var h struct {
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(header, &h); err != nil || h.Alg != "HS256" {
		return Claims{}, ErrInvalid
	}

	// Constant-time signature check before trusting any payload.
	expected := s.sign(parts[0] + "." + parts[1])
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return Claims{}, ErrInvalid
	}

	body, err := decodeSegment(parts[1])
	if err != nil {
		return Claims{}, ErrInvalid
	}
	var claims Claims
	if err := json.Unmarshal(body, &claims); err != nil {
		return Claims{}, ErrInvalid
	}

	if claims.Subject == "" || claims.Type != accessTokenType || claims.ExpiresAt == 0 {
		return Claims{}, ErrInvalid
	}
	if s.issuer != "" && claims.Issuer != s.issuer {
		return Claims{}, ErrInvalid
	}
	if s.audience != "" && claims.Audience != s.audience {
		return Claims{}, ErrInvalid
	}
	if s.now().Unix() >= claims.ExpiresAt {
		return Claims{}, ErrExpired
	}

	return claims, nil
}

func (s *Signer) sign(signingInput string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(signingInput))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func encodeSegment(v any) (string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func decodeSegment(seg string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(seg)
}
