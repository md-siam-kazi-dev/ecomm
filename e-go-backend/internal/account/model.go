package account

import "time"

// Address is a saved shipping address for a user (Bangladesh administrative
// structure). A user can have many addresses; exactly one may be the default.
type Address struct {
	ID            string
	FullName      string
	Phone         string
	Division      string
	District      string
	Upazila       string
	VillageTown   string
	Clarification string
	IsDefault     bool
	CreatedAt     time.Time
}

// Profile is a user's editable profile (name, phone, avatar).
type Profile struct {
	FullName  string
	Email     string
	Phone     string
	AvatarURL string
}

// AddressParams is the validated input for creating or updating an address.
type AddressParams struct {
	FullName      string
	Phone         string
	Division      string
	District      string
	Upazila       string
	VillageTown   string
	Clarification string
	IsDefault     bool
}
