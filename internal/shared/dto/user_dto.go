package dto

import "time"

// UserBasicDTO represents basic user information for cross-domain references
type UserBasicDTO struct {
	ID       uint   `json:"id" example:"1"`
	Email    string `json:"email" example:"user@example.com"`
	Username string `json:"username" example:"johndoe"`
	Name     string `json:"name" example:"John Doe"`
	Avatar   string `json:"avatar" example:"https://example.com/avatar.jpg"`
	Provider string `json:"provider" example:"google"`
	Status   string `json:"status" example:"active"`
	Role     string `json:"role" example:"user"`
}

// UserSummaryDTO represents minimal user information for listings
type UserSummaryDTO struct {
	ID     uint   `json:"id" example:"1"`
	Name   string `json:"name" example:"John Doe"`
	Email  string `json:"email" example:"user@example.com"`
	Avatar string `json:"avatar" example:"https://example.com/avatar.jpg"`
	Status string `json:"status" example:"active"`
}


// UserProfileDTO represents detailed user profile information
type UserProfileDTO struct {
	UserBasicDTO
	GoogleID       *string    `json:"google_id,omitempty" example:"123456789"`
	GitHubID       *string    `json:"github_id,omitempty" example:"123456789"`
	TelegramID     *string    `json:"telegram_id,omitempty" example:"123456789"`
	ProviderData   *string    `json:"provider_data,omitempty"`
	InviteCodeID   *uint      `json:"invite_code_id,omitempty" example:"1"`
	InviteCodeUsed *string    `json:"invite_code_used,omitempty" example:"ABC123"`
	CreatedAt      time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt      time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}
