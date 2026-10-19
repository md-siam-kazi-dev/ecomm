package settings

import "time"

// UpdateRequest is the admin write payload for the storefront settings.
type UpdateRequest struct {
	AnnouncementEnabled bool   `json:"announcement_enabled"`
	AnnouncementMessage string `json:"announcement_message"`
}

// Item is the settings projection returned to both the storefront and admins.
type Item struct {
	AnnouncementEnabled bool   `json:"announcement_enabled"`
	AnnouncementMessage string `json:"announcement_message"`
	UpdatedAt           string `json:"updated_at"`
}

func toItem(s Settings) Item {
	return Item{
		AnnouncementEnabled: s.AnnouncementEnabled,
		AnnouncementMessage: s.AnnouncementMessage,
		UpdatedAt:           s.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
