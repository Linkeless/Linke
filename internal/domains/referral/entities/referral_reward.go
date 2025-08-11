package entities

import (
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/referral/constants"
)

// ReferralReward represents rewards earned from referrals
type ReferralReward struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	ReferralID uint  `json:"referral_id" gorm:"not null;index"`
	UserID     uint  `json:"user_id" gorm:"not null;index"`      // User who earned the reward
	CampaignID *uint `json:"campaign_id,omitempty" gorm:"index"` // Associated campaign

	// Reward Details
	RewardType        string  `json:"reward_type" gorm:"size:50;not null;index"`             // Type of reward (cash, credit, discount, etc.)
	RewardAmount      float64 `json:"reward_amount" gorm:"type:decimal(10,2);not null"`      // Reward amount
	RewardCurrency    string  `json:"reward_currency" gorm:"size:10;not null;default:'USD'"` // Currency
	RewardDescription string  `json:"reward_description" gorm:"size:255"`                    // Description of the reward

	// Reward Status
	Status    string     `json:"status" gorm:"size:20;not null;default:'pending';index"` // pending, earned, paid, cancelled, expired
	EarnedAt  *time.Time `json:"earned_at,omitempty" gorm:"index"`                       // When reward was earned
	PaidAt    *time.Time `json:"paid_at,omitempty" gorm:"index"`                         // When reward was paid
	ExpiresAt *time.Time `json:"expires_at,omitempty" gorm:"index"`                      // When reward expires

	// Payment Information
	PaymentMethod    string `json:"payment_method" gorm:"size:50;index"`     // Payment method used
	PaymentReference string `json:"payment_reference" gorm:"size:100;index"` // Payment reference/transaction ID
	PaymentData      string `json:"payment_data,omitempty" gorm:"type:text"` // Payment details (JSON)

	// Conversion Details
	ConversionValue float64 `json:"conversion_value" gorm:"type:decimal(10,2);default:0"` // Value of conversion that triggered reward
	ConversionType  string  `json:"conversion_type" gorm:"size:50;index"`                 // Type of conversion
	ConversionID    *uint   `json:"conversion_id,omitempty" gorm:"index"`                 // ID of conversion record (e.g., subscription order)

	// Payout Information
	PayoutBatchID *uint   `json:"payout_batch_id,omitempty" gorm:"index"`         // Associated payout batch
	PayoutFee     float64 `json:"payout_fee" gorm:"type:decimal(10,2);default:0"` // Fee charged for payout
	NetAmount     float64 `json:"net_amount" gorm:"type:decimal(10,2);default:0"` // Net amount after fees

	// Approval Workflow
	RequiresApproval bool       `json:"requires_approval" gorm:"not null;default:false"` // Whether reward requires approval
	ApprovedAt       *time.Time `json:"approved_at,omitempty" gorm:"index"`              // When reward was approved
	ApprovedByID     *uint      `json:"approved_by_id,omitempty" gorm:"index"`           // Admin who approved
	RejectedAt       *time.Time `json:"rejected_at,omitempty" gorm:"index"`              // When reward was rejected
	RejectedByID     *uint      `json:"rejected_by_id,omitempty" gorm:"index"`           // Admin who rejected
	RejectionReason  string     `json:"rejection_reason" gorm:"size:255"`                // Reason for rejection

	// Metadata
	Metadata string `json:"metadata,omitempty" gorm:"type:text"` // Additional reward metadata (JSON)
	Notes    string `json:"notes,omitempty" gorm:"type:text"`    // Admin notes

	// Note: Relationships removed to avoid cross-domain dependencies
	// Related data should be fetched and assembled at the application layer

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for ReferralReward model
func (ReferralReward) TableName() string {
	return "referral_rewards"
}


// IsActive checks if the reward is in an active state
func (rr *ReferralReward) IsActive() bool {
	activeStatuses := map[string]bool{
		constants.RewardStatusPending: true,
		constants.RewardStatusEarned:  true,
	}
	return activeStatuses[rr.Status]
}

// IsExpired checks if the reward has expired
func (rr *ReferralReward) IsExpired() bool {
	if rr.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*rr.ExpiresAt)
}

// CanBePaid checks if the reward can be paid out
func (rr *ReferralReward) CanBePaid() bool {
	if rr.Status != constants.RewardStatusEarned {
		return false
	}
	if rr.IsExpired() {
		return false
	}
	if rr.RequiresApproval && rr.ApprovedAt == nil {
		return false
	}
	return true
}

// IsPaymentCompleted checks if payment has been completed
func (rr *ReferralReward) IsPaymentCompleted() bool {
	return rr.Status == constants.RewardStatusPaid && rr.PaidAt != nil
}


// ToResponse should be implemented in service layer to avoid import cycles
// Use dto.ToReferralRewardResponse(rr) instead

// ToPublicResponse should be implemented in service layer to avoid import cycles
// Use dto.ToReferralRewardResponse(rr) and clean sensitive data instead
