package entities

import (
	"time"

	"gorm.io/gorm"
)

// InviteCodeUsage represents the usage record of an invite code
type InviteCodeUsage struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	InviteCodeID uint `json:"invite_code_id" gorm:"not null;index"`
	UsedByID     uint `json:"used_by_id" gorm:"not null;index"`

	// Usage Info
	UsedAt    time.Time `json:"used_at" gorm:"not null;index"`
	IPAddress string    `json:"ip_address" gorm:"size:45"`  // IPv4/IPv6 address
	UserAgent string    `json:"user_agent" gorm:"size:255"` // User agent string

	// Relationships (no foreign key constraints for performance)
	// TODO: Fix cross-domain references in entities
	// InviteCode *InviteCode `json:"invite_code,omitempty" gorm:"-"`
	// UsedBy     *User       `json:"used_by,omitempty" gorm:"-"`

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for InviteCodeUsage model
func (InviteCodeUsage) TableName() string {
	return "invite_code_usages"
}

// ToResponse should be implemented in service layer to avoid import cycles
// Use dto.ToInviteCodeUsageResponse(icu) instead
