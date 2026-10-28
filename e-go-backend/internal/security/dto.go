package security

import "time"

// EventRecord is one persisted security event. It doubles as the admin read
// projection, so the JSON shape is the API contract.
type EventRecord struct {
	ID           string         `json:"id"`
	EventID      string         `json:"event_id"`
	RequestID    *string        `json:"request_id"`
	EventType    string         `json:"event_type"`
	Severity     string         `json:"severity"`
	UserID       *string        `json:"user_id"`
	UserRole     *string        `json:"user_role"`
	IPAddress    *string        `json:"ip_address"`
	Method       *string        `json:"method"`
	Endpoint     *string        `json:"endpoint"`
	StatusCode   *int           `json:"status_code"`
	DurationMS   *int           `json:"duration_ms"`
	UserAgent    *string        `json:"user_agent"`
	ResourceType *string        `json:"resource_type"`
	ResourceID   *string        `json:"resource_id"`
	Metadata     map[string]any `json:"metadata"`
	CreatedAt    time.Time      `json:"created_at"`
}

// StreamEvent is the SSE payload for a live event. Ordinary traffic that was
// counted but not persisted carries Persisted=false and an empty type.
type StreamEvent struct {
	EventID      string         `json:"event_id"`
	RequestID    string         `json:"request_id,omitempty"`
	EventType    string         `json:"event_type,omitempty"`
	Severity     string         `json:"severity,omitempty"`
	UserID       *string        `json:"user_id"`
	UserRole     *string        `json:"user_role"`
	IPAddress    *string        `json:"ip_address"`
	Method       string         `json:"method,omitempty"`
	Endpoint     string         `json:"endpoint,omitempty"`
	StatusCode   int            `json:"status_code"`
	DurationMS   int            `json:"duration_ms"`
	UserAgent    string         `json:"user_agent,omitempty"`
	ResourceType *string        `json:"resource_type"`
	ResourceID   *string        `json:"resource_id"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	Persisted    bool           `json:"persisted"`
	CreatedAt    time.Time      `json:"created_at"`
}

// ToStreamEvent projects an internal event for the live stream.
func ToStreamEvent(ev SecurityEvent) StreamEvent {
	return StreamEvent{
		EventID:      ev.EventID,
		RequestID:    ev.RequestID,
		EventType:    string(ev.Type),
		Severity:     string(ev.Severity),
		UserID:       optional(ev.UserID),
		UserRole:     optional(ev.UserRole),
		IPAddress:    optional(ev.IP),
		Method:       ev.Method,
		Endpoint:     ev.Endpoint,
		StatusCode:   ev.StatusCode,
		DurationMS:   ev.DurationMS,
		UserAgent:    ev.UserAgent,
		ResourceType: optional(ev.ResourceType),
		ResourceID:   optional(ev.ResourceID),
		Metadata:     sanitizeMetadata(ev.Metadata),
		Persisted:    ev.Persisted,
		CreatedAt:    ev.CreatedAt,
	}
}

// statsResponse is the overview payload: the aggregated snapshot plus the
// health of the collector itself, so lost telemetry is visible in the UI.
type statsResponse struct {
	Snapshot
	DroppedEvents int64 `json:"dropped_events"`
	PersistedOnly bool  `json:"persisted_stream_only"`
}

func optional(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
