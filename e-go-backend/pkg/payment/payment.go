// Package payment defines the payment-provider abstraction used by the
// application-fee workflow. No real provider is integrated yet: the platform
// ships with the NotConfigured placeholder, and a future provider plugs in by
// implementing Provider and registering it — no service or route changes.
package payment

import (
	"context"
	"errors"
	"strings"
)

// ErrNotConfigured is returned when the platform has no real payment provider
// wired up. Callers surface this as an explicit "not configured" state rather
// than fabricating a successful payment.
var ErrNotConfigured = errors.New("payment: no payment provider is configured")

// ErrUnknownProvider is returned when a configured provider name has no
// implementation registered.
var ErrUnknownProvider = errors.New("payment: unknown provider")

// InitRequest starts a payment. Amount is a decimal string (never a float).
// SuccessURL/CancelURL optionally override the provider defaults (used to send
// the customer back to a specific order).
type InitRequest struct {
	Reference  string
	Amount     string
	Currency   string
	Metadata   map[string]string
	SuccessURL string
	CancelURL  string
}

// InitResult is what a provider returns when a payment session is created.
type InitResult struct {
	Provider      string
	RedirectURL   string
	ClientPayload map[string]string
}

// VerifyRequest carries the data needed to verify a payment server-side (from a
// signed webhook or a provider API call). RawBody is the untouched request body.
type VerifyRequest struct {
	Reference             string
	ProviderTransactionID string
	Signature             string
	RawBody               []byte
}

// VerifyResult is the provider-verified outcome. Only Paid==true, returned by
// the provider, may move a payment to `paid`.
type VerifyResult struct {
	Reference             string
	ProviderTransactionID string
	Amount                string
	Currency              string
	Paid                  bool
}

// Provider is implemented by payment gateways.
type Provider interface {
	Name() string
	Initiate(ctx context.Context, req InitRequest) (InitResult, error)
	Verify(ctx context.Context, req VerifyRequest) (VerifyResult, error)
}

// NotConfigured is the placeholder provider used when none is configured.
type NotConfigured struct{}

func (NotConfigured) Name() string { return "" }

func (NotConfigured) Initiate(context.Context, InitRequest) (InitResult, error) {
	return InitResult{}, ErrNotConfigured
}

func (NotConfigured) Verify(context.Context, VerifyRequest) (VerifyResult, error) {
	return VerifyResult{}, ErrNotConfigured
}

// Registry resolves providers by name.
type Registry struct {
	byName map[string]Provider
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{byName: make(map[string]Provider)}
}

// Register adds a provider. Providers with an empty name are ignored.
func (r *Registry) Register(p Provider) {
	if p == nil || p.Name() == "" {
		return
	}
	r.byName[p.Name()] = p
}

// Get returns the provider for name. An empty name yields NotConfigured; an
// unknown non-empty name yields ErrUnknownProvider.
func (r *Registry) Get(name string) (Provider, error) {
	if name == "" {
		return NotConfigured{}, nil
	}
	if p, ok := r.byName[name]; ok {
		return p, nil
	}
	return nil, ErrUnknownProvider
}

// Config carries the payment-provider wiring from application configuration.
type Config struct {
	Provider      string
	SecretKey     string
	WebhookSecret string
	Currency      string
	SuccessURL    string
	CancelURL     string
}

// NewRegistryFromConfig builds the registry from configuration. Real providers
// are registered here; an unconfigured name resolves to ErrUnknownProvider.
func NewRegistryFromConfig(cfg Config) *Registry {
	r := NewRegistry()
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "stripe":
		if strings.TrimSpace(cfg.SecretKey) != "" {
			r.Register(NewStripe(cfg.SecretKey, cfg.WebhookSecret, cfg.Currency, cfg.SuccessURL, cfg.CancelURL))
		}
	}
	return r
}
