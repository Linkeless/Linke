package entities

import (
	"time"

	"linke/internal/domains/user/constants"

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

	// Provider Metadata (MySQL JSON type)
	ProviderData *string `json:"provider_data,omitempty" gorm:"type:json"`

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

// IsDeleted checks if the user is soft deleted
func (u *User) IsDeleted() bool {
	return u.DeletedAt.Valid
}

// IsActive checks if the user is active
func (u *User) IsActive() bool {
	return u.Status == constants.UserStatusActive && !u.IsDeleted()
}

// GetID returns the user's ID (required for middleware interface)
func (u *User) GetID() uint {
	return u.ID
}

// IsAdmin checks if the user is an admin
func (u *User) IsAdmin() bool {
	return u.Role == constants.UserRoleAdmin && u.IsActive()
}

// IsLocalAccount checks if this is a local (email/password) account
func (u *User) IsLocalAccount() bool {
	return u.Provider == constants.ProviderLocal
}

// IsOAuthAccount checks if this is an OAuth account
func (u *User) IsOAuthAccount() bool {
	return u.Provider != constants.ProviderLocal
}

// GetProviderID returns the provider-specific ID based on the provider
func (u *User) GetProviderID() string {
	switch u.Provider {
	case constants.ProviderGoogle:
		if u.GoogleID != nil {
			return *u.GoogleID
		}
	case constants.ProviderGitHub:
		if u.GitHubID != nil {
			return *u.GitHubID
		}
	case constants.ProviderTelegram:
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
