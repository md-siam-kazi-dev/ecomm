package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
)

// Mailer delivers verification and password-reset links. A real provider
// (SES, Postmark, Resend, ...) is plugged in for production; the development
// implementation logs the link instead of sending mail.
type Mailer interface {
	SendVerificationEmail(ctx context.Context, to, token string) error
	SendPasswordResetEmail(ctx context.Context, to, token string) error
}

// LogMailer is the development mailer: it writes the action link to the log so
// the flow can be exercised without an email provider.
type LogMailer struct {
	log         *slog.Logger
	frontendURL string
	production  bool
}

func NewLogMailer(log *slog.Logger, frontendURL string, production bool) *LogMailer {
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	return &LogMailer{log: log, frontendURL: frontendURL, production: production}
}

func (m *LogMailer) SendVerificationEmail(ctx context.Context, to, token string) error {
	m.logLink(ctx, "verify_email", to, "/auth/verify-email", token)
	return nil
}

func (m *LogMailer) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	m.logLink(ctx, "reset_password", to, "/auth/reset-password", token)
	return nil
}

func (m *LogMailer) logLink(ctx context.Context, kind, to, path, token string) {
	if m.production {
		// Never write action links to production logs.
		m.log.WarnContext(ctx, "email delivery not configured",
			slog.String("kind", kind), slog.String("to", to))
		return
	}
	link := fmt.Sprintf("%s%s?token=%s", m.frontendURL, path, url.QueryEscape(token))
	m.log.InfoContext(ctx, "auth_email_link",
		slog.String("kind", kind),
		slog.String("to", to),
		slog.String("link", link),
	)
}
