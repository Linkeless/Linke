package model

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
	UsedAt time.Time `json:"used_at" gorm:"not null;index"`
	IPAddress string `json:"ip_address" gorm:"size:45"` // IPv4/IPv6 address
	UserAgent string `json:"user_agent" gorm:"size:255"` // User agent string

	// Relationships (no foreign key constraints for performance)
	InviteCode *InviteCode `json:"invite_code,omitempty" gorm:"-"`
	UsedBy     *User       `json:"used_by,omitempty" gorm:"-"`

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for InviteCodeUsage model
func (InviteCodeUsage) TableName() string {
	return "invite_code_usages"
}

// InviteCodeUsageResponse represents the invite code usage data structure for API responses
type InviteCodeUsageResponse struct {
	ID           uint                 `json:"id" example:"1"`                                    // Usage record ID
	InviteCodeID uint                 `json:"invite_code_id" example:"1"`                       // Invite code ID
	UsedByID     uint                 `json:"used_by_id" example:"2"`                           // User ID who used the code
	UsedAt       time.Time            `json:"used_at" example:"2024-01-01T00:00:00Z"`          // When the code was used
	IPAddress    string               `json:"ip_address" example:"192.168.1.100"`              // IP address of the user
	UserAgent    string               `json:"user_agent" example:"Mozilla/5.0..."`             // User agent string
	CreatedAt    time.Time            `json:"created_at" example:"2024-01-01T00:00:00Z"`       // Creation time
	
	// Optional related data
	InviteCode *InviteCodeResponse `json:"invite_code,omitempty"` // Invite code details
	UsedBy     *UserResponse       `json:"used_by,omitempty"`     // User who used the code
}

// ToResponse converts InviteCodeUsage to InviteCodeUsageResponse
func (icu *InviteCodeUsage) ToResponse() *InviteCodeUsageResponse {
	resp := &InviteCodeUsageResponse{
		ID:           icu.ID,
		InviteCodeID: icu.InviteCodeID,
		UsedByID:     icu.UsedByID,
		UsedAt:       icu.UsedAt,
		IPAddress:    icu.IPAddress,
		UserAgent:    icu.UserAgent,
		CreatedAt:    icu.CreatedAt,
	}
	
	// Include related data if loaded
	if icu.InviteCode != nil {
		resp.InviteCode = icu.InviteCode.ToResponse()
	}
	if icu.UsedBy != nil {
		resp.UsedBy = icu.UsedBy.ToResponse()
	}
	
	return resp
}