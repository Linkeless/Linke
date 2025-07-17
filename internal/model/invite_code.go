package model

import (
	"time"

	"gorm.io/gorm"
)

// InviteCode represents an invitation code
type InviteCode struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Core Fields
	Code        string `json:"code" gorm:"uniqueIndex;size:32;not null"`        // 邀请码
	CreatedByID uint   `json:"created_by_id" gorm:"not null;index"`             // 创建者ID
	
	// Status and Limits
	Status      string `json:"status" gorm:"size:20;not null;default:'active';index"` // active, used, disabled
	MaxUses     int    `json:"max_uses" gorm:"not null;default:10"`                    // 最大使用次数
	UsedCount   int    `json:"used_count" gorm:"not null;default:0"`                   // 已使用次数
	
	// Metadata
	Description string `json:"description" gorm:"size:255"` // 描述
	Metadata    string `json:"metadata,omitempty" gorm:"type:text"` // 额外元数据(JSON)

	// Relationships (no foreign key constraints for performance)
	CreatedBy *User                `json:"created_by,omitempty" gorm:"-"`
	UsageRecords []*InviteCodeUsage `json:"usage_records,omitempty" gorm:"-"`

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for InviteCode model
func (InviteCode) TableName() string {
	return "invite_codes"
}

// Status constants
const (
	InviteCodeStatusActive   = "active"
	InviteCodeStatusUsed     = "used"
	InviteCodeStatusDisabled = "disabled"
)

// IsActive checks if the invite code is active and can be used
func (ic *InviteCode) IsActive() bool {
	if ic.Status != InviteCodeStatusActive {
		return false
	}
	
	// Check if max uses reached
	if ic.UsedCount >= ic.MaxUses {
		return false
	}
	
	return true
}


// IsExhausted checks if the invite code has reached its maximum uses
func (ic *InviteCode) IsExhausted() bool {
	return ic.UsedCount >= ic.MaxUses
}

// CanBeUsed checks if the invite code can be used
func (ic *InviteCode) CanBeUsed() bool {
	return ic.IsActive() && !ic.IsDeleted()
}

// IsDeleted checks if the invite code is soft deleted
func (ic *InviteCode) IsDeleted() bool {
	return ic.DeletedAt.Valid
}

// InviteCodeResponse represents the invite code data structure for API responses
type InviteCodeResponse struct {
	ID          uint      `json:"id" example:"1"`                                        // Invite code ID
	Code        string    `json:"code" example:"a1b2c3d4e5f6789012345678901234567890abcd"` // Invite code string
	CreatedByID uint      `json:"created_by_id" example:"1"`                             // Creator user ID
	Status      string    `json:"status" example:"active" enums:"active,used,disabled"`   // Invite code status
	MaxUses     int       `json:"max_uses" example:"10"`                                 // Maximum number of uses
	UsedCount   int       `json:"used_count" example:"0"`                                // Current usage count
	Description string    `json:"description" example:"Friend invitation code"`          // Description
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`            // Creation time
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`            // Last update time
	
	// Optional related data
	CreatedBy    *UserResponse               `json:"created_by,omitempty"`    // Creator user info
	UsageRecords []*InviteCodeUsageResponse  `json:"usage_records,omitempty"` // Usage records
}

// ToResponse converts InviteCode to InviteCodeResponse
func (ic *InviteCode) ToResponse() *InviteCodeResponse {
	resp := &InviteCodeResponse{
		ID:          ic.ID,
		Code:        ic.Code,
		CreatedByID: ic.CreatedByID,
		Status:      ic.Status,
		MaxUses:     ic.MaxUses,
		UsedCount:   ic.UsedCount,
		Description: ic.Description,
		CreatedAt:   ic.CreatedAt,
		UpdatedAt:   ic.UpdatedAt,
	}
	
	// Include related data if loaded
	if ic.CreatedBy != nil {
		resp.CreatedBy = ic.CreatedBy.ToResponse()
	}
	if ic.UsageRecords != nil {
		for _, usage := range ic.UsageRecords {
			resp.UsageRecords = append(resp.UsageRecords, usage.ToResponse())
		}
	}
	
	return resp
}

// ToPublicResponse converts InviteCode to a public response (hides sensitive info)
func (ic *InviteCode) ToPublicResponse() *InviteCodeResponse {
	return &InviteCodeResponse{
		Code:        ic.Code,
		Status:      ic.Status,
		MaxUses:     ic.MaxUses,
		UsedCount:   ic.UsedCount,
		Description: ic.Description,
	}
}