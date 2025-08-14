package dto

import (
	"time"

	userDTO "linke/internal/domains/user/dto"
)

// AuthResponse represents authentication response
type AuthResponse struct {
	User  *UserResponse  `json:"user"`
	Token *TokenResponse `json:"token"`
}

// UserResponse represents user data in auth responses (simplified for Swagger)
type UserResponse struct {
	ID        uint      `json:"id" example:"1"`
	Email     string    `json:"email" example:"user@example.com"`
	Username  string    `json:"username" example:"johndoe"`
	Name      string    `json:"name" example:"John Doe"`
	Avatar    string    `json:"avatar" example:"https://example.com/avatar.jpg"`
	Provider  string    `json:"provider" example:"local"`
	Status    string    `json:"status" example:"active"`
	Role      string    `json:"role" example:"user"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// ConvertUserResponse converts userDTO.UserResponse to auth DTO UserResponse
func ConvertUserResponse(userResp *userDTO.UserResponse) *UserResponse {
	if userResp == nil {
		return nil
	}
	return &UserResponse{
		ID:        userResp.ID,
		Email:     userResp.Email,
		Username:  userResp.Username,
		Name:      userResp.Name,
		Avatar:    userResp.Avatar,
		Provider:  userResp.Provider,
		Status:    userResp.Status,
		Role:      userResp.Role,
		CreatedAt: userResp.CreatedAt,
		UpdatedAt: userResp.UpdatedAt,
	}
}

// TokenResponse represents token information
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

// Claims represents JWT token claims (moved to claims.go)

// UserInfo represents OAuth user information from providers
type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Provider string `json:"provider"`
}
