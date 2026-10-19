package audit

import (
	"strings"
	"time"
)

// Actor is the person (or system) that performed an action.
type Actor struct {
	ID    *string `json:"id"`
	Name  string  `json:"name"`
	Email *string `json:"email"`
}

// EventItem is the list projection of an audit event.
type EventItem struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Action      string    `json:"action"`
	ActionLabel string    `json:"action_label"`
	Category    string    `json:"category"`
	Actor       Actor     `json:"actor"`
	ActorRole   *string   `json:"actor_role"`
	EntityType  *string   `json:"entity_type"`
	EntityID    *string   `json:"entity_id"`
	Result      *string   `json:"result"`
	Reason      *string   `json:"reason"`
}

// EventDetail is the full projection including payloads and request metadata.
type EventDetail struct {
	EventItem
	OldValues map[string]any `json:"old_values"`
	NewValues map[string]any `json:"new_values"`
	Metadata  map[string]any `json:"metadata"`
	IPAddress *string        `json:"ip_address"`
	UserAgent *string        `json:"user_agent"`
	RequestID *string        `json:"request_id"`
}

// Summary is the aggregate shown above the audit list.
type Summary struct {
	Total          int64           `json:"total"`
	AdminActions   int64           `json:"admin_actions"`
	VendorActions  int64           `json:"vendor_actions"`
	SecurityEvents int64           `json:"security_events"`
	From           *string         `json:"from,omitempty"`
	To             *string         `json:"to,omitempty"`
	Actions        []LabeledAction `json:"actions"`
}

func toEventItem(rec EventRecord) EventItem {
	return EventItem{
		ID:          rec.ID,
		CreatedAt:   rec.CreatedAt,
		Action:      rec.Action,
		ActionLabel: Label(rec.Action),
		Category:    category(rec.Action, rec.Result),
		Actor: Actor{
			ID:    rec.ActorID,
			Name:  rec.ActorName,
			Email: rec.ActorEmail,
		},
		ActorRole:  rec.ActorRole,
		EntityType: rec.EntityType,
		EntityID:   rec.EntityID,
		Result:     rec.Result,
		Reason:     rec.Reason,
	}
}

// category classifies an event for UI grouping/filtering.
func category(action string, result *string) string {
	if strings.HasPrefix(action, "auth.") || (result != nil && *result == ResultFailure) {
		return "security"
	}
	if strings.HasPrefix(action, "admin.") {
		return "admin"
	}
	if strings.HasPrefix(action, "vendor.") {
		return "vendor"
	}
	return "system"
}
