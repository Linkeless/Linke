package entities

import (
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/referral/constants"
)

// InviteCode represents an invitation code
type InviteCode struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Core Fields
	Code        string `json:"code" gorm:"uniqueIndex;size:32;not null"` // 邀请码
	CreatedByID uint   `json:"created_by_id" gorm:"not null;index"`      // 创建者ID

	// Referral Integration Fields
	ReferralCampaignID     *uint   `json:"referral_campaign_id,omitempty" gorm:"index"`                            // 关联的推广活动ID
	ReferralRewardAmount   float64 `json:"referral_reward_amount" gorm:"type:decimal(10,2);not null;default:0.00"` // 推广奖励金额
	ReferralRewardCurrency string  `json:"referral_reward_currency" gorm:"size:10;not null;default:'CNY'"`         // 推广奖励货币

	// Status and Limits
	Status    string `json:"status" gorm:"size:20;not null;default:'active';index"` // active, used, disabled
	MaxUses   int    `json:"max_uses" gorm:"not null;default:10"`                   // 最大使用次数
	UsedCount int    `json:"used_count" gorm:"not null;default:0"`                  // 已使用次数

	// Metadata
	Description string `json:"description" gorm:"size:255"`         // 描述
	Metadata    string `json:"metadata,omitempty" gorm:"type:text"` // 额外元数据(JSON)

	// Relationships (no foreign key constraints for performance)
	// TODO: Fix cross-domain references in entities
	// CreatedBy *userEntities.User                `json:"created_by,omitempty" gorm:"-"`
	// UsageRecords []*InviteCodeUsage `json:"usage_records,omitempty" gorm:"-"`
	// ReferralCampaign *ReferralCampaign `json:"referral_campaign,omitempty" gorm:"-"`

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for InviteCode model
func (InviteCode) TableName() string {
	return "invite_codes"
}

// IsActive checks if the invite code is active and can be used
func (ic *InviteCode) IsActive() bool {
	if ic.Status != constants.InviteCodeStatusActive {
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

// ToResponse should be implemented in service layer to avoid import cycles
// Use dto.ToInviteCodeResponse(ic) instead

// ToPublicResponse should be implemented in service layer to avoid import cycles
// Use dto.ToInviteCodeResponse(ic) and clean sensitive data instead
