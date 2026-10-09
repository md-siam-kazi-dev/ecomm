package auth

import (
	"net/mail"
	"strings"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

const (
	minPasswordLength = 8
	// bcrypt only considers the first 72 bytes of a password.
	maxPasswordLength = 72
)

// validateRegistration validates the sign-up payload.
func validateRegistration(name, email, plainPassword string) error {
	if name == "" || email == "" || plainPassword == "" {
		return apperr.ErrValidation
	}
	if err := validateEmail(email); err != nil {
		return err
	}
	return validatePassword(plainPassword)
}

// validateEmail performs a light syntactic check and length bound.
func validateEmail(email string) error {
	if email == "" || len(email) > 255 {
		return apperr.ErrValidation
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return apperr.ErrValidation
	}
	return nil
}

// validatePassword enforces the length policy (PRD §18).
func validatePassword(plainPassword string) error {
	if len(plainPassword) < minPasswordLength || len(plainPassword) > maxPasswordLength {
		return apperr.ErrValidation
	}
	return nil
}

// splitName maps a display name onto the profile's first_name / last_name.
func splitName(name string) (first, last string) {
	parts := strings.Fields(name)
	switch len(parts) {
	case 0:
		return "", ""
	case 1:
		return parts[0], ""
	default:
		return parts[0], strings.Join(parts[1:], " ")
	}
}
