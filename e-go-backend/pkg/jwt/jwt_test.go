package jwt

import (
	"testing"
	"time"
)

func TestSignParseRoundTrip(t *testing.T) {
	s := NewSigner("secret", "iss", "aud", 15*time.Minute)

	token, err := s.Sign("user-1", "admin")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	claims, err := s.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.Subject != "user-1" || claims.Role != "admin" || claims.Type != "access" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestParse_Expired(t *testing.T) {
	s := NewSigner("secret", "iss", "aud", time.Minute)
	s.now = func() time.Time { return time.Now().Add(-time.Hour) }
	token, err := s.Sign("user-1", "user")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	s.now = time.Now

	if _, err := s.Parse(token); err != ErrExpired {
		t.Fatalf("expected ErrExpired, got %v", err)
	}
}

func TestParse_WrongSecret(t *testing.T) {
	other := NewSigner("other-secret", "iss", "aud", time.Minute)
	token, _ := other.Sign("user-1", "user")

	s := NewSigner("secret", "iss", "aud", time.Minute)
	if _, err := s.Parse(token); err != ErrInvalid {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestParse_TamperedPayload(t *testing.T) {
	s := NewSigner("secret", "iss", "aud", time.Minute)
	token, _ := s.Sign("user-1", "user")

	// Swap the signature segment for another validly-signed token's signature.
	other, _ := s.Sign("attacker", "admin")
	tampered := token[:len(token)-len(signature(other))] + signature(other)

	if _, err := s.Parse(tampered); err != ErrInvalid {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestParse_WrongType(t *testing.T) {
	s := NewSigner("secret", "iss", "aud", time.Minute)

	header, _ := encodeSegment(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, _ := encodeSegment(Claims{
		Subject:   "user-1",
		Role:      "user",
		Type:      "refresh",
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Minute).Unix(),
	})
	input := header + "." + payload
	token := input + "." + s.sign(input)

	if _, err := s.Parse(token); err != ErrInvalid {
		t.Fatalf("expected ErrInvalid for wrong token type, got %v", err)
	}
}

func TestParse_MissingSubject(t *testing.T) {
	s := NewSigner("secret", "iss", "aud", time.Minute)

	header, _ := encodeSegment(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, _ := encodeSegment(Claims{Type: "access", ExpiresAt: time.Now().Add(time.Minute).Unix()})
	input := header + "." + payload
	token := input + "." + s.sign(input)

	if _, err := s.Parse(token); err != ErrInvalid {
		t.Fatalf("expected ErrInvalid for missing subject, got %v", err)
	}
}

func TestParse_Malformed(t *testing.T) {
	s := NewSigner("secret", "iss", "aud", time.Minute)
	for _, token := range []string{"", "not-a-jwt", "a.b", "a.b.c.d"} {
		if _, err := s.Parse(token); err != ErrInvalid {
			t.Errorf("Parse(%q) = %v, want ErrInvalid", token, err)
		}
	}
}

func signature(token string) string { return token[len(token)-43:] }
