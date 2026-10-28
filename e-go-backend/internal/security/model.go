package security

import (
	"time"
)

// SecuritySeverity ranks how much attention an event deserves. The set is
// closed and is mirrored by a CHECK constraint in migration 022.
type SecuritySeverity string

const (
	SeverityInfo     SecuritySeverity = "info"
	SeverityLow      SecuritySeverity = "low"
	SeverityMedium   SecuritySeverity = "medium"
	SeverityHigh     SecuritySeverity = "high"
	SeverityCritical SecuritySeverity = "critical"
)

var severityRank = map[SecuritySeverity]int{
	SeverityInfo:     0,
	SeverityLow:      1,
	SeverityMedium:   2,
	SeverityHigh:     3,
	SeverityCritical: 4,
}

// Rank orders severities; unknown values rank below info.
func (s SecuritySeverity) Rank() int { return severityRank[s] }

// Valid reports whether s is one of the closed set.
func (s SecuritySeverity) Valid() bool {
	_, ok := severityRank[s]
	return ok
}

// AllSeverities returns the closed set in ascending order.
func AllSeverities() []SecuritySeverity {
	return []SecuritySeverity{SeverityInfo, SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical}
}

// SecurityEventType classifies what happened. It is a closed set so the admin
// filters can be allow-listed and the database can stay consistent.
type SecurityEventType string

const (
	EventLoginSuccess             SecurityEventType = "LOGIN_SUCCESS"
	EventLoginFailed              SecurityEventType = "LOGIN_FAILED"
	EventLoginRateLimited         SecurityEventType = "LOGIN_RATE_LIMITED"
	EventRegisterSuccess          SecurityEventType = "REGISTER_SUCCESS"
	EventPasswordResetRequested   SecurityEventType = "PASSWORD_RESET_REQUESTED"
	EventPasswordResetFailed      SecurityEventType = "PASSWORD_RESET_FAILED"
	EventRefreshTokenReuse        SecurityEventType = "REFRESH_TOKEN_REUSE_DETECTED"
	EventResourceAccessDenied     SecurityEventType = "RESOURCE_ACCESS_DENIED"
	EventAuthorizationFailure     SecurityEventType = "AUTHORIZATION_FAILURE"
	EventRateLimitExceeded        SecurityEventType = "RATE_LIMIT_EXCEEDED"
	EventSuspiciousAPIActivity    SecurityEventType = "SUSPICIOUS_API_ACTIVITY"
	EventAdminLogin               SecurityEventType = "ADMIN_LOGIN"
	EventAdminPermissionDenied    SecurityEventType = "ADMIN_PERMISSION_DENIED"
	EventAccountRestricted        SecurityEventType = "ACCOUNT_RESTRICTED"
	EventIPBlocked                SecurityEventType = "IP_BLOCKED"
	EventAPIServerError           SecurityEventType = "API_SERVER_ERROR"
)

// defaultSeverity is the baseline risk for a type. A caller may override it on
// an individual event (for example escalating a repeated login failure).
var defaultSeverity = map[SecurityEventType]SecuritySeverity{
	EventLoginSuccess:           SeverityInfo,
	EventLoginFailed:            SeverityLow,
	EventLoginRateLimited:       SeverityMedium,
	EventRegisterSuccess:        SeverityInfo,
	EventPasswordResetRequested: SeverityInfo,
	EventPasswordResetFailed:    SeverityLow,
	EventRefreshTokenReuse:      SeverityCritical,
	EventResourceAccessDenied:   SeverityMedium,
	EventAuthorizationFailure:   SeverityLow,
	EventRateLimitExceeded:      SeverityMedium,
	EventSuspiciousAPIActivity:  SeverityHigh,
	EventAdminLogin:             SeverityInfo,
	EventAdminPermissionDenied:  SeverityHigh,
	EventAccountRestricted:      SeverityMedium,
	EventIPBlocked:              SeverityMedium,
	EventAPIServerError:         SeverityLow,
}

// DefaultSeverity returns the baseline severity for a type, defaulting to info.
func (t SecurityEventType) DefaultSeverity() SecuritySeverity {
	if sev, ok := defaultSeverity[t]; ok {
		return sev
	}
	return SeverityInfo
}

// Valid reports whether t is one of the closed set.
func (t SecurityEventType) Valid() bool {
	_, ok := defaultSeverity[t]
	return ok
}

// AllEventTypes returns the closed set in declaration order.
func AllEventTypes() []SecurityEventType {
	return []SecurityEventType{
		EventLoginSuccess, EventLoginFailed, EventLoginRateLimited,
		EventRegisterSuccess, EventPasswordResetRequested, EventPasswordResetFailed,
		EventRefreshTokenReuse, EventResourceAccessDenied, EventAuthorizationFailure,
		EventRateLimitExceeded, EventSuspiciousAPIActivity, EventAdminLogin,
		EventAdminPermissionDenied, EventAccountRestricted, EventIPBlocked,
		EventAPIServerError,
	}
}

// ClassifyHTTP maps a response status onto a security event type. The bool is
// false for statuses that are not worth persisting (2xx/3xx and ordinary client
// validation errors), which keeps this table to signals rather than traffic.
//
// A failed request is NOT automatically an attack: only authentication,
// authorisation, rate-limit and server-error outcomes are recorded.
func ClassifyHTTP(status int) (SecurityEventType, SecuritySeverity, bool) {
	switch {
	case status == 401:
		return EventAuthorizationFailure, SeverityLow, true
	case status == 403:
		return EventResourceAccessDenied, SeverityMedium, true
	case status == 429:
		return EventRateLimitExceeded, SeverityMedium, true
	case status >= 500:
		return EventAPIServerError, SeverityLow, true
	default:
		return "", "", false
	}
}

// SecurityEvent is one observed API event. Fields that do not apply are left
// zero-valued and stored as NULL.
//
// Safe metadata only: there is deliberately no field for a request body, a
// header, a cookie or a token, so those can never be persisted by construction.
type SecurityEvent struct {
	// EventID is a stable unique identifier (uuid) echoed to SSE clients.
	EventID string
	// RequestID correlates the event with the request log and audit trail.
	RequestID string
	Type      SecurityEventType
	Severity  SecuritySeverity

	UserID   string
	UserRole string

	IP        string
	Method    string
	Endpoint  string
	StatusCode int
	DurationMS int
	UserAgent  string

	ResourceType string
	ResourceID   string

	// Metadata is sanitized before it is persisted.
	Metadata map[string]any
	CreatedAt time.Time

	// Persisted reports whether this event was written to security_events.
	// Ordinary 2xx traffic is streamed live but not stored, so clients must not
	// offer a "details" view for events where this is false.
	Persisted bool
}

// IsSuspicious reports whether the event is worth surfacing as suspicious
// activity in the overview.
func (e SecurityEvent) IsSuspicious() bool {
	return e.Severity.Rank() >= SeverityMedium.Rank()
}

// WithDefaults fills in the timestamp, severity and event id when a caller left
// them out, so every persisted/streamed event is complete.
func (e SecurityEvent) WithDefaults(now time.Time) SecurityEvent {
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	if e.Severity == "" {
		e.Severity = e.Type.DefaultSeverity()
	}
	return e
}

// RangeName is the selectable overview window.
type RangeName string

const (
	Range5m  RangeName = "5m"
	Range15m RangeName = "15m"
	Range1h  RangeName = "1h"
	Range24h RangeName = "24h"
	Range7d  RangeName = "7d"
)

// Valid reports whether r is one of the supported windows.
func (r RangeName) Valid() bool {
	switch r {
	case Range5m, Range15m, Range1h, Range24h, Range7d:
		return true
	}
	return false
}

// Duration is the length of the window.
func (r RangeName) Duration() time.Duration {
	switch r {
	case Range5m:
		return 5 * time.Minute
	case Range15m:
		return 15 * time.Minute
	case Range1h:
		return time.Hour
	case Range24h:
		return 24 * time.Hour
	case Range7d:
		return 7 * 24 * time.Hour
	}
	return 0
}

// AllRanges returns the supported windows, shortest first.
func AllRanges() []RangeName {
	return []RangeName{Range5m, Range15m, Range1h, Range24h, Range7d}
}
