package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Core Identity Fields
	Email    string `json:"email" gorm:"uniqueIndex;size:255;not null"`
	Username string `json:"username" gorm:"index;size:100"`
	Name     string `json:"name" gorm:"size:255"` // Can be auto-generated from email
	Avatar   string `json:"avatar" gorm:"size:500"`

	// Authentication Fields
	Password string `json:"-" gorm:"size:255"` // bcrypt hash, hidden from JSON
	Provider string `json:"provider" gorm:"size:50;not null;default:'local';index"`
	Status   string `json:"status" gorm:"size:20;not null;default:'active';index"` // active, inactive, banned
	Role     string `json:"role" gorm:"size:20;not null;default:'user';index"`     // user, admin

	// OAuth Provider IDs (nullable for local accounts)
	GoogleID   *string `json:"google_id,omitempty" gorm:"uniqueIndex;size:100"`
	GitHubID   *string `json:"github_id,omitempty" gorm:"uniqueIndex;size:100;column:github_id"`
	TelegramID *string `json:"telegram_id,omitempty" gorm:"uniqueIndex;size:100"`

	// Provider Metadata
	ProviderData string `json:"provider_data,omitempty" gorm:"type:text"`

	// Invite Code Fields
	InviteCodeID   *uint   `json:"invite_code_id,omitempty" gorm:"index"`           // 使用的邀请码ID
	InviteCodeUsed *string `json:"invite_code_used,omitempty" gorm:"size:32;index"` // 使用的邀请码(冗余字段，便于查询)

	// Timestamp Fields (GORM convention order)
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for User model
func (User) TableName() string {
	return "users"
}

// User status constants
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusBanned   = "banned"
)

// User role constants
const (
	UserRoleUser  = "user"
	UserRoleAdmin = "admin"
)

// Provider constants
const (
	ProviderLocal    = "local"
	ProviderGoogle   = "google"
	ProviderGitHub   = "github"
	ProviderTelegram = "telegram"
)

// IsDeleted checks if the user is soft deleted
func (u *User) IsDeleted() bool {
	return u.DeletedAt.Valid
}

// IsActive checks if the user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive && !u.IsDeleted()
}

// IsAdmin checks if the user is an admin
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin && u.IsActive()
}

// IsLocalAccount checks if this is a local (email/password) account
func (u *User) IsLocalAccount() bool {
	return u.Provider == ProviderLocal
}

// IsOAuthAccount checks if this is an OAuth account
func (u *User) IsOAuthAccount() bool {
	return u.Provider != ProviderLocal
}

// GetProviderID returns the provider-specific ID based on the provider
func (u *User) GetProviderID() string {
	switch u.Provider {
	case ProviderGoogle:
		if u.GoogleID != nil {
			return *u.GoogleID
		}
	case ProviderGitHub:
		if u.GitHubID != nil {
			return *u.GitHubID
		}
	case ProviderTelegram:
		if u.TelegramID != nil {
			return *u.TelegramID
		}
	}
	return ""
}

// SoftDelete performs soft delete on the user
func (u *User) SoftDelete(db *gorm.DB) error {
	return db.Delete(u).Error
}

// Restore restores a soft deleted user
func (u *User) Restore(db *gorm.DB) error {
	return db.Unscoped().Model(u).Update("deleted_at", nil).Error
}

// UserResponse represents the user data structure for API responses
// Fields are ordered to match the User model for consistency
type UserResponse struct {
	// Primary Key
	ID uint `json:"id"`

	// Core Identity Fields
	Email    string `json:"email"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`

	// Authentication Fields (excluding password)
	Provider string `json:"provider"`
	Status   string `json:"status"`
	Role     string `json:"role"`

	// OAuth Provider IDs (only show if not empty)
	GoogleID   *string `json:"google_id,omitempty"`
	GitHubID   *string `json:"github_id,omitempty"`
	TelegramID *string `json:"telegram_id,omitempty"`

	// Provider Metadata (only show if not empty)
	ProviderData string `json:"provider_data,omitempty"`

	// Invite Code Fields
	InviteCodeID   *uint   `json:"invite_code_id,omitempty"`
	InviteCodeUsed *string `json:"invite_code_used,omitempty"`

	// Timestamp Fields
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	resp := &UserResponse{
		// Primary Key
		ID: u.ID,

		// Core Identity Fields
		Email:    u.Email,
		Username: u.Username,
		Name:     u.Name,
		Avatar:   u.Avatar,

		// Authentication Fields
		Provider: u.Provider,
		Status:   u.Status,
		Role:     u.Role,

		// OAuth Provider IDs
		GoogleID:   u.GoogleID,
		GitHubID:   u.GitHubID,
		TelegramID: u.TelegramID,

		// Provider Metadata
		ProviderData: u.ProviderData,

		// Invite Code Fields
		InviteCodeID:   u.InviteCodeID,
		InviteCodeUsed: u.InviteCodeUsed,

		// Timestamp Fields
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}

	// Set DeletedAt only if valid
	if u.DeletedAt.Valid {
		resp.DeletedAt = &u.DeletedAt.Time
	}

	return resp
}