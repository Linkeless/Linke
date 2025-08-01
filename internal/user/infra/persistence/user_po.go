package persistence

import (
	"time"

	"gorm.io/gorm"
)

// UserPO represents the user persistence object for database storage
type UserPO struct {
	// Primary Key
	ID uint `gorm:"primaryKey" json:"id"`

	// Core Identity Fields
	Email    string `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Username string `gorm:"index;size:100" json:"username"`
	Name     string `gorm:"size:255" json:"name"`
	Avatar   string `gorm:"size:500" json:"avatar"`

	// Authentication Fields
	Password string `gorm:"size:255" json:"-"` // Hidden from JSON
	Provider string `gorm:"size:50;not null;default:'local';index" json:"provider"`
	Status   string `gorm:"size:20;not null;default:'active';index" json:"status"`
	Role     string `gorm:"size:20;not null;default:'user';index" json:"role"`

	// OAuth Provider IDs (nullable for local accounts)
	GoogleID   *string `gorm:"uniqueIndex;size:100" json:"google_id,omitempty"`
	GitHubID   *string `gorm:"uniqueIndex;size:100;column:github_id" json:"github_id,omitempty"`
	TelegramID *string `gorm:"uniqueIndex;size:100" json:"telegram_id,omitempty"`

	// Provider Metadata (MySQL JSON type)
	ProviderData *string `gorm:"type:json" json:"provider_data,omitempty"`

	// Invite Code Fields
	InviteCodeID   *uint   `gorm:"index" json:"invite_code_id,omitempty"`
	InviteCodeUsed *string `gorm:"size:32;index" json:"invite_code_used,omitempty"`

	// Timestamp Fields (GORM convention order)
	CreatedAt time.Time      `gorm:"not null;index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName returns the table name for UserPO
func (UserPO) TableName() string {
	return "users"
}

// IsDeleted checks if the user is soft deleted
func (u *UserPO) IsDeleted() bool {
	return u.DeletedAt.Valid
}

// IsActive checks if the user is active
func (u *UserPO) IsActive() bool {
	return u.Status == "active" && !u.IsDeleted()
}

// IsAdmin checks if the user is an admin
func (u *UserPO) IsAdmin() bool {
	return u.Role == "admin" && u.IsActive()
}

// IsLocalAccount checks if this is a local (email/password) account
func (u *UserPO) IsLocalAccount() bool {
	return u.Provider == "local"
}

// IsOAuthAccount checks if this is an OAuth account
func (u *UserPO) IsOAuthAccount() bool {
	return u.Provider != "local"
}

// GetProviderID returns the provider-specific ID based on the provider
func (u *UserPO) GetProviderID() string {
	switch u.Provider {
	case "google":
		if u.GoogleID != nil {
			return *u.GoogleID
		}
	case "github":
		if u.GitHubID != nil {
			return *u.GitHubID
		}
	case "telegram":
		if u.TelegramID != nil {
			return *u.TelegramID
		}
	}
	return ""
}

// SoftDelete performs soft delete on the user
func (u *UserPO) SoftDelete(db *gorm.DB) error {
	return db.Delete(u).Error
}

// Restore restores a soft deleted user
func (u *UserPO) Restore(db *gorm.DB) error {
	return db.Unscoped().Model(u).Update("deleted_at", nil).Error
}