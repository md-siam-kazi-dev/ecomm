// Package courier defines the integration boundary for shipment tracking
// providers. No carrier API is wired up yet, so the only implementation is
// Manual: the vendor records the carrier and tracking number themselves. We
// never fabricate a tracking URL or claim a parcel was dispatched.
//
// A real provider (Pathao, DHL, …) plugs in by implementing Provider and
// registering it in NewRegistry — no route or service change required.
package courier

// Provider resolves buyer-facing tracking for a shipment.
type Provider interface {
	// Name is the provider identifier (e.g. "manual").
	Name() string
	// TrackingURL returns a public tracking URL, or "" when the provider cannot
	// build one.
	TrackingURL(carrier, trackingNumber string) string
}

// Manual is the default provider: tracking is whatever the vendor entered.
type Manual struct{}

// Name implements Provider.
func (Manual) Name() string { return "manual" }

// TrackingURL returns "" — there is no carrier integration to build one.
func (Manual) TrackingURL(_, _ string) string { return "" }

// Registry holds the available courier providers.
type Registry struct {
	providers map[string]Provider
}

// NewRegistry builds a registry from the supplied providers.
func NewRegistry(providers ...Provider) *Registry {
	r := &Registry{providers: make(map[string]Provider, len(providers))}
	for _, p := range providers {
		r.providers[p.Name()] = p
	}
	return r
}

// Default returns the manual-only registry used until a carrier is integrated.
func Default() *Registry { return NewRegistry(Manual{}) }

// TrackingURL asks the named provider for a tracking URL, falling back to
// manual (which returns "") when the provider is unknown.
func (r *Registry) TrackingURL(provider, carrier, trackingNumber string) string {
	if p, ok := r.providers[provider]; ok {
		return p.TrackingURL(carrier, trackingNumber)
	}
	if trackingNumber == "" {
		return ""
	}
	return (Manual{}).TrackingURL(carrier, trackingNumber)
}
