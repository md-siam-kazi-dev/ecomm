package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/password"
)

func TestSplitName(t *testing.T) {
	cases := []struct {
		in          string
		first, last string
	}{
		{"John Doe", "John", "Doe"},
		{"John", "John", ""},
		{"John Ronald Reuel Tolkien", "John", "Ronald Reuel Tolkien"},
		{"  Amelia   Hart  ", "Amelia", "Hart"},
		{"", "", ""},
	}

	for _, c := range cases {
		first, last := splitName(c.in)
		if first != c.first || last != c.last {
			t.Errorf("splitName(%q) = %q/%q, want %q/%q", c.in, first, last, c.first, c.last)
		}
	}
}

func TestRegister_HashesAndNormalizes(t *testing.T) {
	env := newTestEnv()

	if _, err := env.svc.Register(context.Background(), RegisterRequest{
		Name: "John Doe", Email: "John@Example.COM", Password: "super-secret-123",
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !password.Verify(env.users.lastCreate.PasswordHash, "super-secret-123") {
		t.Error("stored hash is not a valid bcrypt hash of the password")
	}
	if env.users.lastCreate.Email != "john@example.com" {
		t.Errorf("email not normalized: %q", env.users.lastCreate.Email)
	}
}

func TestRegister_PropagatesRepositoryError(t *testing.T) {
	env := newTestEnv()
	env.users.createErr = errEmailTaken

	_, err := env.svc.Register(context.Background(), RegisterRequest{
		Name: "John Doe", Email: "john@example.com", Password: "super-secret-123",
	})
	if !errors.Is(err, errEmailTaken) {
		t.Fatalf("expected errEmailTaken, got %v", err)
	}
}

func TestRegister_ValidationErrors(t *testing.T) {
	cases := map[string]RegisterRequest{
		"empty name":        {Name: "", Email: "john@example.com", Password: "super-secret-123"},
		"invalid email":     {Name: "John Doe", Email: "not-an-email", Password: "super-secret-123"},
		"short password":    {Name: "John Doe", Email: "john@example.com", Password: "short"},
		"password too long": {Name: "John Doe", Email: "john@example.com", Password: string(make([]byte, 73))},
	}

	for name, req := range cases {
		t.Run(name, func(t *testing.T) {
			env := newTestEnv()
			_, err := env.svc.Register(context.Background(), req)

			appErr, ok := apperr.As(err)
			if !ok || appErr.Code != apperr.ErrValidation.Code {
				t.Fatalf("expected validation error, got %v", err)
			}
			if env.users.createCalls != 0 {
				t.Error("repository must not be called when validation fails")
			}
		})
	}
}

func TestHighestRole(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{[]string{"user"}, "user"},
		{[]string{"user", "vendor"}, "vendor"},
		{[]string{"user", "vendor", "admin"}, "admin"},
		{[]string{}, "user"},
	}
	for _, c := range cases {
		if got := highestRole(c.in); got != c.want {
			t.Errorf("highestRole(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}
