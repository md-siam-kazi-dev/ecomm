package auth

// RegisterRequest is the payload accepted by POST /api/v1/auth/register (PRD §11).
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
}

// LoginRequest is the payload accepted by POST /api/v1/auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// VerifyEmailRequest is the payload accepted by POST /api/v1/auth/verify-email.
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// ForgotPasswordRequest is the payload accepted by POST /api/v1/auth/forgot-password.
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest is the payload accepted by POST /api/v1/auth/reset-password.
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// RegisterResponse is the account returned after a successful registration. It
// never contains the password or its hash (PRD §64).
type RegisterResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

// UserResponse is the safe representation of an account returned by login,
// refresh and /me. It never includes password_hash, refresh tokens or secrets.
type UserResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Phone         string `json:"phone,omitempty"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
	CreatedAt     string `json:"created_at,omitempty"`
}

// userEnvelope wraps a user so responses read { "user": { ... } }.
type userEnvelope struct {
	User UserResponse `json:"user"`
}

func toUserResponse(u User) UserResponse {
	resp := UserResponse{
		ID:            u.ID,
		Email:         u.Email,
		Name:          u.Name,
		Phone:         u.Phone,
		AvatarURL:     u.AvatarURL,
		Role:          u.Role,
		EmailVerified: u.EmailVerified,
	}
	if !u.CreatedAt.IsZero() {
		resp.CreatedAt = u.CreatedAt.UTC().Format("2006-01-02")
	}
	return resp
}
