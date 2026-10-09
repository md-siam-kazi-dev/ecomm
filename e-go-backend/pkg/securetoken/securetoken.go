// Package securetoken generates cryptographically random opaque tokens (refresh
// tokens, email-verification tokens, password-reset tokens) and hashes them for
// storage, so the raw value never touches the database.
package securetoken

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// tokenBytes is the entropy of a generated token (256 bits).
const tokenBytes = 32

// Generate returns a URL-safe random token and its SHA-256 hash (lowercase hex).
// Store only the hash; hand the raw token to the user.
func Generate() (raw string, hash string, err error) {
	buf := make([]byte, tokenBytes)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(buf)
	return raw, Hash(raw), nil
}

// Hash returns the lowercase hex SHA-256 of a raw token. Hashing is
// deterministic, so a presented token can be looked up by its hash.
func Hash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
