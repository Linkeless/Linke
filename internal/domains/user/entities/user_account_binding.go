package entities

import (
	"time"

	"linke/internal/domains/user/constants"

	"gorm.io/gorm"
)

// UserAccountBinding represents a user's third-party account binding
type UserAccountBinding struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Key (no constraint, managed by application)
	UserID uint `json:"user_id" gorm:"not null;index"`

	// Provider Information
	Provider       string `json:"provider" gorm:"size:20;not null;index"` // google, github, telegram
	ProviderUserID string `json:"provider_user_id" gorm:"size:100;not null;index"`

	// Provider User Data
	ProviderEmail    *string `json:"provider_email,omitempty" gorm:"size:255;index"`
	ProviderUsername *string `json:"provider_username,omitempty" gorm:"size:100"`
	ProviderName     *string `json:"provider_name,omitempty" gorm:"size:255"`
	ProviderAvatar   *string `json:"provider_avatar,omitempty" gorm:"size:500"`
	ProviderData     *string `json:"provider_data,omitempty" gorm:"type:json"`

	// Binding Status
	IsPrimary  bool       `json:"is_primary" gorm:"default:false"`
	BoundAt    time.Time  `json:"bound_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty" gorm:"index"`

	// Timestamp Fields (GORM convention order)
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for UserAccountBinding model
func (UserAccountBinding) TableName() string {
	return "user_account_bindings"
}

// IsDeleted checks if the binding is soft deleted
func (uab *UserAccountBinding) IsDeleted() bool {
	return uab.DeletedAt.Valid
}

// IsValidProvider checks if a provider is valid for this binding
func (uab *UserAccountBinding) IsValidProvider() bool {
	validProviders := constants.ValidBindingProviders()
	for _, validProvider := range validProviders {
		if uab.Provider == validProvider {
			return true
		}
	}
	return false
}

// UpdateLastUsed updates the last used timestamp
func (uab *UserAccountBinding) UpdateLastUsed(db *gorm.DB) error {
	now := time.Now()
	uab.LastUsedAt = &now
	return db.Model(uab).Update("last_used_at", now).Error
}

// SetAsPrimary sets this binding as the primary one for the user
// Note: This should be used within a transaction to ensure consistency
func (uab *UserAccountBinding) SetAsPrimary(db *gorm.DB) error {
	// First, unset all other bindings for this user
	if err := db.Model(&UserAccountBinding{}).
		Where("user_id = ? AND id != ?", uab.UserID, uab.ID).
		Update("is_primary", false).Error; err != nil {
		return err
	}

	// Then set this one as primary
	uab.IsPrimary = true
	return db.Model(uab).Update("is_primary", true).Error
}

// SoftDelete performs soft delete on the binding
func (uab *UserAccountBinding) SoftDelete(db *gorm.DB) error {
	return db.Delete(uab).Error
}

// Restore restores a soft deleted binding
func (uab *UserAccountBinding) Restore(db *gorm.DB) error {
	return db.Unscoped().Model(uab).Update("deleted_at", nil).Error
}

// Business logic methods

// CanBeSetAsPrimary checks if this binding can be set as primary
func (uab *UserAccountBinding) CanBeSetAsPrimary() bool {
	return !uab.IsDeleted() && uab.IsValidProvider()
}

// GetDisplayName returns a user-friendly display name for the binding
func (uab *UserAccountBinding) GetDisplayName() string {
	// Prefer provider name, fallback to username, then email, finally provider user ID
	if uab.ProviderName != nil && *uab.ProviderName != "" {
		return *uab.ProviderName
	}
	if uab.ProviderUsername != nil && *uab.ProviderUsername != "" {
		return *uab.ProviderUsername
	}
	if uab.ProviderEmail != nil && *uab.ProviderEmail != "" {
		return *uab.ProviderEmail
	}
	return uab.ProviderUserID
}

// IsActive checks if the binding is active (not deleted and recently used)
func (uab *UserAccountBinding) IsActive() bool {
	if uab.IsDeleted() {
		return false
	}

	// Consider binding active if it was used within the last 90 days or is primary
	if uab.IsPrimary {
		return true
	}

	if uab.LastUsedAt == nil {
		// If never used but bound recently (within 7 days), consider active
		return time.Since(uab.BoundAt) <= 7*24*time.Hour
	}

	// Check if used within last 90 days
	return time.Since(*uab.LastUsedAt) <= 90*24*time.Hour
}

// GetBindingAge returns the age of the binding in days
func (uab *UserAccountBinding) GetBindingAge() int {
	return int(time.Since(uab.BoundAt).Hours() / 24)
}

// GetLastUsedDaysAgo returns days since last use, -1 if never used
func (uab *UserAccountBinding) GetLastUsedDaysAgo() int {
	if uab.LastUsedAt == nil {
		return -1
	}
	return int(time.Since(*uab.LastUsedAt).Hours() / 24)
}
