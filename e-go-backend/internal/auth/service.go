package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/jwt"
	"github.com/md-siam-kazi/e-go-backend/pkg/password"
	"github.com/md-siam-kazi/e-go-backend/pkg/securetoken"
)

const (
	// verifyTokenTTL is how long an email-verification link stays valid.
	verifyTokenTTL = 24 * time.Hour
	// resetTokenTTL is how long a password-reset link stays valid.
	resetTokenTTL = time.Hour
)

var (
	errInvalidCredentials = apperr.New(http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
	errAccountInactive    = apperr.New(http.StatusForbidden, "ACCOUNT_INACTIVE", "This account is not active")
)

// dummyHash is compared against when an email is unknown so login latency does
// not reveal whether an account exists.
var dummyHash = func() string {
	hash, _ := password.Hash("timing-equalization-placeholder")
	return hash
}()

// ClientMeta carries request metadata stored with a session.
type ClientMeta struct {
	IPAddress string
	UserAgent string
}

// LoginResult is what login and refresh hand back: the safe user plus the two
// raw tokens to place in cookies. The tokens are never serialized to JSON.
type LoginResult struct {
	User         User
	AccessToken  string
	RefreshToken string
}

// ServiceDeps wires the auth service's collaborators (dependency injection).
type ServiceDeps struct {
	Users      UserRepository
	Sessions   SessionRepository
	Tokens     TokenRepository
	Signer     *jwt.Signer
	Mailer     Mailer
	Log        *slog.Logger
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// Service owns the authentication business logic (Handler → Service →
// Repository, PRD §4).
type Service struct {
	users      UserRepository
	sessions   SessionRepository
	tokens     TokenRepository
	signer     *jwt.Signer
	mailer     Mailer
	log        *slog.Logger
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewService(d ServiceDeps) *Service {
	return &Service{
		users:      d.Users,
		sessions:   d.Sessions,
		tokens:     d.Tokens,
		signer:     d.Signer,
		mailer:     d.Mailer,
		log:        d.Log,
		accessTTL:  d.AccessTTL,
		refreshTTL: d.RefreshTTL,
	}
}

// Register creates an account (unchanged behaviour) and additionally issues an
// email-verification token. Mail delivery is best-effort and never fails the
// registration itself.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (RegisterResponse, error) {
	name := strings.TrimSpace(req.Name)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	phone := strings.TrimSpace(req.Phone)

	if err := validateRegistration(name, email, req.Password); err != nil {
		return RegisterResponse{}, err
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		return RegisterResponse{}, apperr.ErrInternal.WithCause(err)
	}

	firstName, lastName := splitName(name)

	created, err := s.users.CreateUser(ctx, CreateUserParams{
		Email:        email,
		Phone:        phone,
		PasswordHash: hash,
		FirstName:    firstName,
		LastName:     lastName,
	})
	if err != nil {
		return RegisterResponse{}, err
	}

	s.sendVerificationEmail(ctx, created.ID, created.Email)

	return RegisterResponse{
		ID:     created.ID,
		Name:   name,
		Email:  created.Email,
		Status: created.Status,
	}, nil
}

// Login verifies credentials, opens a session and issues cookies-ready tokens.
// Any credential failure returns the same generic error so account existence is
// not revealed.
func (s *Service) Login(ctx context.Context, req LoginRequest, meta ClientMeta) (LoginResult, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		return LoginResult{}, errInvalidCredentials
	}

	user, err := s.users.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, errUserNotFound) {
			// Spend comparable time so unknown emails are not distinguishable.
			_ = password.Verify(dummyHash, req.Password)
			return LoginResult{}, errInvalidCredentials
		}
		return LoginResult{}, err
	}

	if !password.Verify(user.PasswordHash, req.Password) {
		return LoginResult{}, errInvalidCredentials
	}
	if user.Status == "suspended" || user.Status == "blocked" {
		return LoginResult{}, errAccountInactive
	}

	return s.issueSession(ctx, user, meta)
}

// Refresh validates the presented refresh token, rotates it, and issues a new
// token pair. Presenting an already-revoked token is treated as compromise and
// revokes every session for that user.
func (s *Service) Refresh(ctx context.Context, rawRefresh string, meta ClientMeta) (LoginResult, error) {
	if rawRefresh == "" {
		return LoginResult{}, apperr.ErrUnauthorized
	}

	session, err := s.sessions.FindByRefreshHash(ctx, securetoken.Hash(rawRefresh))
	if err != nil {
		if errors.Is(err, errSessionNotFound) {
			return LoginResult{}, apperr.ErrUnauthorized
		}
		return LoginResult{}, err
	}

	if session.RevokedAt != nil {
		_ = s.sessions.RevokeAllForUser(ctx, session.UserID)
		return LoginResult{}, apperr.ErrUnauthorized
	}
	if time.Now().After(session.ExpiresAt) {
		return LoginResult{}, apperr.ErrUnauthorized
	}

	user, err := s.users.FindUserByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, errUserNotFound) {
			return LoginResult{}, apperr.ErrUnauthorized
		}
		return LoginResult{}, err
	}
	if user.Status == "suspended" || user.Status == "blocked" {
		return LoginResult{}, errAccountInactive
	}

	raw, newHash, err := securetoken.Generate()
	if err != nil {
		return LoginResult{}, apperr.ErrInternal.WithCause(err)
	}
	if _, err := s.sessions.Rotate(ctx, session.ID, SessionParams{
		UserID:      user.ID,
		RefreshHash: newHash,
		UserAgent:   meta.UserAgent,
		IPAddress:   meta.IPAddress,
		ExpiresAt:   time.Now().Add(s.refreshTTL),
	}); err != nil {
		return LoginResult{}, err
	}

	access, err := s.signer.Sign(user.ID, user.Role)
	if err != nil {
		return LoginResult{}, apperr.ErrInternal.WithCause(err)
	}
	return LoginResult{User: user, AccessToken: access, RefreshToken: raw}, nil
}

// Logout revokes the session behind the presented refresh token. It is
// idempotent: an unknown or absent token is not an error.
func (s *Service) Logout(ctx context.Context, rawRefresh string) error {
	if rawRefresh == "" {
		return nil
	}
	return s.sessions.RevokeByHash(ctx, securetoken.Hash(rawRefresh))
}

// LogoutAll revokes every active session for the user.
func (s *Service) LogoutAll(ctx context.Context, userID string) error {
	return s.sessions.RevokeAllForUser(ctx, userID)
}

// Me returns the current account, re-read from the database so role/status are
// always authoritative.
func (s *Service) Me(ctx context.Context, userID string) (UserResponse, error) {
	user, err := s.users.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, errUserNotFound) {
			return UserResponse{}, apperr.ErrUnauthorized
		}
		return UserResponse{}, err
	}
	return toUserResponse(user), nil
}

// VerifyEmail consumes a verification token and marks the email verified.
func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	if token == "" {
		return errInvalidToken
	}
	userID, err := s.tokens.Consume(ctx, securetoken.Hash(token), purposeEmailVerification)
	if err != nil {
		return err
	}
	return s.users.MarkEmailVerified(ctx, userID)
}

// ForgotPassword always succeeds for the caller. When the email matches an
// account it invalidates prior tokens, stores a fresh one and sends the link.
func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil
	}

	user, err := s.users.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, errUserNotFound) {
			return nil
		}
		return err
	}

	if err := s.tokens.InvalidateUnused(ctx, user.ID, purposePasswordReset); err != nil {
		return err
	}
	raw, hash, err := securetoken.Generate()
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if err := s.tokens.Create(ctx, TokenParams{
		UserID:    user.ID,
		Purpose:   purposePasswordReset,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(resetTokenTTL),
	}); err != nil {
		return err
	}

	if err := s.mailer.SendPasswordResetEmail(ctx, user.Email, raw); err != nil {
		s.log.ErrorContext(ctx, "failed to send password reset email", slog.String("error", err.Error()))
	}
	return nil
}

// ResetPassword consumes a reset token, stores the new password hash and
// revokes every existing session.
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	if token == "" {
		return errInvalidToken
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	userID, err := s.tokens.Consume(ctx, securetoken.Hash(token), purposePasswordReset)
	if err != nil {
		return err
	}
	hash, err := password.Hash(newPassword)
	if err != nil {
		return apperr.ErrInternal.WithCause(err)
	}
	if err := s.users.UpdatePassword(ctx, userID, hash); err != nil {
		return err
	}
	return s.sessions.RevokeAllForUser(ctx, userID)
}

// issueSession creates a session row and signs an access token.
func (s *Service) issueSession(ctx context.Context, user User, meta ClientMeta) (LoginResult, error) {
	raw, hash, err := securetoken.Generate()
	if err != nil {
		return LoginResult{}, apperr.ErrInternal.WithCause(err)
	}
	if _, err := s.sessions.Create(ctx, SessionParams{
		UserID:      user.ID,
		RefreshHash: hash,
		UserAgent:   meta.UserAgent,
		IPAddress:   meta.IPAddress,
		ExpiresAt:   time.Now().Add(s.refreshTTL),
	}); err != nil {
		return LoginResult{}, err
	}

	access, err := s.signer.Sign(user.ID, user.Role)
	if err != nil {
		return LoginResult{}, apperr.ErrInternal.WithCause(err)
	}
	_ = s.users.TouchLastLogin(ctx, user.ID)

	return LoginResult{User: user, AccessToken: access, RefreshToken: raw}, nil
}

// sendVerificationEmail creates and mails a verification token (best-effort).
func (s *Service) sendVerificationEmail(ctx context.Context, userID, email string) {
	raw, hash, err := securetoken.Generate()
	if err != nil {
		s.log.ErrorContext(ctx, "failed to generate verification token", slog.String("error", err.Error()))
		return
	}
	if err := s.tokens.Create(ctx, TokenParams{
		UserID:    userID,
		Purpose:   purposeEmailVerification,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(verifyTokenTTL),
	}); err != nil {
		s.log.ErrorContext(ctx, "failed to store verification token", slog.String("error", err.Error()))
		return
	}
	if err := s.mailer.SendVerificationEmail(ctx, email, raw); err != nil {
		s.log.ErrorContext(ctx, "failed to send verification email", slog.String("error", err.Error()))
	}
}
