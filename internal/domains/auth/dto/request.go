package dto

// RegisterRequest represents registration request data
type RegisterRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	InviteCode string `json:"invite_code"` // Optional invite code
}

// LoginRequest represents login request data
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest represents the password change request structure
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldPassword123"`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"newPassword123"`
}

// TokenExchangeRequest represents token exchange request
type TokenExchangeRequest struct {
	Provider string `json:"provider" binding:"required"`
	Code     string `json:"code" binding:"required"`
	State    string `json:"state,omitempty"`
}

// AuthorizeURLRequest represents authorization URL request
type AuthorizeURLRequest struct {
	Provider    string `json:"provider" binding:"required"`
	State       string `json:"state,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

// AuthorizeURLResponse represents authorization URL response
type AuthorizeURLResponse struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}