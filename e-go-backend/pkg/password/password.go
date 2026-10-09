// Package password hashes and verifies user passwords with bcrypt.
package password

import "golang.org/x/crypto/bcrypt"

// Cost is the bcrypt cost factor. 12 is a good balance between resistance to
// brute force and login latency.
const Cost = 12

// Hash returns the bcrypt hash of a plaintext password.
func Hash(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), Cost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// Verify reports whether plain matches the stored bcrypt hash.
func Verify(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
