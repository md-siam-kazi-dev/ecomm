package account

import "time"

// AddressRequest is the create/update payload for an address.
type AddressRequest struct {
	FullName      string `json:"full_name"`
	Phone         string `json:"phone"`
	Division      string `json:"division"`
	District      string `json:"district"`
	Upazila       string `json:"upazila"`
	VillageTown   string `json:"village_town"`
	Clarification string `json:"clarification"`
	IsDefault     bool   `json:"is_default"`
}

// AddressResponse is the safe representation of a saved address.
type AddressResponse struct {
	ID            string `json:"id"`
	FullName      string `json:"full_name"`
	Phone         string `json:"phone"`
	Division      string `json:"division"`
	District      string `json:"district"`
	Upazila       string `json:"upazila"`
	VillageTown   string `json:"village_town"`
	Clarification string `json:"clarification"`
	IsDefault     bool   `json:"is_default"`
	CreatedAt     string `json:"created_at"`
}

// ProfileRequest is the payload for updating name + phone.
type ProfileRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

// ProfileResponse is the user's profile after a read/write.
type ProfileResponse struct {
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	AvatarURL string `json:"avatar_url"`
}

func toProfileResponse(p Profile) ProfileResponse {
	return ProfileResponse{
		FullName:  p.FullName,
		Email:     p.Email,
		Phone:     p.Phone,
		AvatarURL: p.AvatarURL,
	}
}

func toResponse(a Address) AddressResponse {
	return AddressResponse{
		ID:            a.ID,
		FullName:      a.FullName,
		Phone:         a.Phone,
		Division:      a.Division,
		District:      a.District,
		Upazila:       a.Upazila,
		VillageTown:   a.VillageTown,
		Clarification: a.Clarification,
		IsDefault:     a.IsDefault,
		CreatedAt:     a.CreatedAt.UTC().Format(time.RFC3339),
	}
}
