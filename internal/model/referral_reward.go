package model

import (
	"time"

	"gorm.io/gorm"
)

// ReferralReward represents rewards earned from referrals
type ReferralReward struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	ReferralID uint `json:"referral_id" gorm:"not null;index"`
	UserID     uint `json:"user_id" gorm:"not null;index"`       // User who earned the reward
	CampaignID *uint `json:"campaign_id,omitempty" gorm:"index"` // Associated campaign
	
	// Reward Details
	RewardType        string  `json:"reward_type" gorm:"size:50;not null;index"`        // Type of reward (cash, credit, discount, etc.)
	RewardAmount      float64 `json:"reward_amount" gorm:"type:decimal(10,2);not null"` // Reward amount
	RewardCurrency    string  `json:"reward_currency" gorm:"size:10;not null;default:'USD'"` // Currency
	RewardDescription string  `json:"reward_description" gorm:"size:255"`               // Description of the reward
	
	// Reward Status
	Status            string     `json:"status" gorm:"size:20;not null;default:'pending';index"` // pending, earned, paid, cancelled, expired
	EarnedAt          *time.Time `json:"earned_at,omitempty" gorm:"index"`                       // When reward was earned
	PaidAt            *time.Time `json:"paid_at,omitempty" gorm:"index"`                         // When reward was paid
	ExpiresAt         *time.Time `json:"expires_at,omitempty" gorm:"index"`                      // When reward expires
	
	// Payment Information
	PaymentMethod     string     `json:"payment_method" gorm:"size:50;index"`                    // Payment method used
	PaymentReference  string     `json:"payment_reference" gorm:"size:100;index"`                // Payment reference/transaction ID
	PaymentData       string     `json:"payment_data,omitempty" gorm:"type:text"`                // Payment details (JSON)
	
	// Conversion Details
	ConversionValue   float64    `json:"conversion_value" gorm:"type:decimal(10,2);default:0"`   // Value of conversion that triggered reward
	ConversionType    string     `json:"conversion_type" gorm:"size:50;index"`                   // Type of conversion
	ConversionID      *uint      `json:"conversion_id,omitempty" gorm:"index"`                   // ID of conversion record (e.g., subscription order)
	
	// Payout Information
	PayoutBatchID     *uint      `json:"payout_batch_id,omitempty" gorm:"index"`                 // Associated payout batch
	PayoutFee         float64    `json:"payout_fee" gorm:"type:decimal(10,2);default:0"`         // Fee charged for payout
	NetAmount         float64    `json:"net_amount" gorm:"type:decimal(10,2);default:0"`         // Net amount after fees
	
	// Approval Workflow
	RequiresApproval  bool       `json:"requires_approval" gorm:"not null;default:false"`        // Whether reward requires approval
	ApprovedAt        *time.Time `json:"approved_at,omitempty" gorm:"index"`                     // When reward was approved
	ApprovedByID      *uint      `json:"approved_by_id,omitempty" gorm:"index"`                  // Admin who approved
	RejectedAt        *time.Time `json:"rejected_at,omitempty" gorm:"index"`                     // When reward was rejected
	RejectedByID      *uint      `json:"rejected_by_id,omitempty" gorm:"index"`                  // Admin who rejected
	RejectionReason   string     `json:"rejection_reason" gorm:"size:255"`                       // Reason for rejection
	
	// Metadata
	Metadata          string     `json:"metadata,omitempty" gorm:"type:text"`                    // Additional reward metadata (JSON)
	Notes             string     `json:"notes,omitempty" gorm:"type:text"`                       // Admin notes
	
	// Relationships (no foreign key constraints for performance)
	Referral          *Referral         `json:"referral,omitempty" gorm:"-"`
	User              *User             `json:"user,omitempty" gorm:"-"`
	Campaign          *ReferralCampaign `json:"campaign,omitempty" gorm:"-"`
	ApprovedBy        *User             `json:"approved_by,omitempty" gorm:"-"`
	RejectedBy        *User             `json:"rejected_by,omitempty" gorm:"-"`

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for ReferralReward model
func (ReferralReward) TableName() string {
	return "referral_rewards"
}

// Reward Type constants
const (
	RewardTypeCash          = "cash"
	RewardTypeCredit        = "credit"
	RewardTypeBonus         = "bonus"
	RewardTypeCoupon        = "coupon"
	RewardTypeSubscription  = "subscription"
	RewardTypeProduct       = "product"
	RewardTypeService       = "service"
)

// Reward Status constants (specific to referral rewards)
const (
	RewardStatusExpired   = "expired"
	RewardStatusRejected  = "rejected"
)

// Payment Method constants
const (
	PaymentMethodPayPal     = "paypal"
	PaymentMethodStripe     = "stripe"
	PaymentMethodBankTransfer = "bank_transfer"
	PaymentMethodCredit     = "credit"
	PaymentMethodCrypto     = "crypto"
	PaymentMethodCheck      = "check"
)

// IsActive checks if the reward is in an active state
func (rr *ReferralReward) IsActive() bool {
	activeStatuses := map[string]bool{
		RewardStatusPending: true,
		RewardStatusEarned:  true,
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
	if rr.Status != RewardStatusEarned {
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
	return rr.Status == RewardStatusPaid && rr.PaidAt != nil
}

// ReferralRewardResponse represents the referral reward data structure for API responses
type ReferralRewardResponse struct {
	ID                uint       `json:"id" example:"1"`
	ReferralID        uint       `json:"referral_id" example:"1"`
	UserID            uint       `json:"user_id" example:"2"`
	CampaignID        *uint      `json:"campaign_id,omitempty" example:"1"`
	RewardType        string     `json:"reward_type" example:"cash"`
	RewardAmount      float64    `json:"reward_amount" example:"10.00"`
	RewardCurrency    string     `json:"reward_currency" example:"USD"`
	RewardDescription string     `json:"reward_description" example:"Referral bonus for new subscriber"`
	Status            string     `json:"status" example:"earned"`
	EarnedAt          *time.Time `json:"earned_at,omitempty" example:"2024-01-01T00:00:00Z"`
	PaidAt            *time.Time `json:"paid_at,omitempty" example:"2024-01-01T00:00:00Z"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z"`
	PaymentMethod     string     `json:"payment_method" example:"paypal"`
	PaymentReference  string     `json:"payment_reference" example:"PAY-123456789"`
	ConversionValue   float64    `json:"conversion_value" example:"29.99"`
	ConversionType    string     `json:"conversion_type" example:"subscription"`
	ConversionID      *uint      `json:"conversion_id,omitempty" example:"5"`
	PayoutBatchID     *uint      `json:"payout_batch_id,omitempty" example:"1"`
	PayoutFee         float64    `json:"payout_fee" example:"0.50"`
	NetAmount         float64    `json:"net_amount" example:"9.50"`
	RequiresApproval  bool       `json:"requires_approval" example:"false"`
	ApprovedAt        *time.Time `json:"approved_at,omitempty" example:"2024-01-01T00:00:00Z"`
	ApprovedByID      *uint      `json:"approved_by_id,omitempty" example:"1"`
	RejectedAt        *time.Time `json:"rejected_at,omitempty" example:"2024-01-01T00:00:00Z"`
	RejectedByID      *uint      `json:"rejected_by_id,omitempty" example:"1"`
	RejectionReason   string     `json:"rejection_reason" example:"Fraudulent activity detected"`
	CreatedAt         time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt         time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	
	// Optional related data
	Referral          *ReferralResponse         `json:"referral,omitempty"`
	User              *UserResponse             `json:"user,omitempty"`
	Campaign          *ReferralCampaignResponse `json:"campaign,omitempty"`
	ApprovedBy        *UserResponse             `json:"approved_by,omitempty"`
	RejectedBy        *UserResponse             `json:"rejected_by,omitempty"`
}

// ToResponse converts ReferralReward to ReferralRewardResponse
func (rr *ReferralReward) ToResponse() *ReferralRewardResponse {
	resp := &ReferralRewardResponse{
		ID:                rr.ID,
		ReferralID:        rr.ReferralID,
		UserID:            rr.UserID,
		CampaignID:        rr.CampaignID,
		RewardType:        rr.RewardType,
		RewardAmount:      rr.RewardAmount,
		RewardCurrency:    rr.RewardCurrency,
		RewardDescription: rr.RewardDescription,
		Status:            rr.Status,
		EarnedAt:          rr.EarnedAt,
		PaidAt:            rr.PaidAt,
		ExpiresAt:         rr.ExpiresAt,
		PaymentMethod:     rr.PaymentMethod,
		PaymentReference:  rr.PaymentReference,
		ConversionValue:   rr.ConversionValue,
		ConversionType:    rr.ConversionType,
		ConversionID:      rr.ConversionID,
		PayoutBatchID:     rr.PayoutBatchID,
		PayoutFee:         rr.PayoutFee,
		NetAmount:         rr.NetAmount,
		RequiresApproval:  rr.RequiresApproval,
		ApprovedAt:        rr.ApprovedAt,
		ApprovedByID:      rr.ApprovedByID,
		RejectedAt:        rr.RejectedAt,
		RejectedByID:      rr.RejectedByID,
		RejectionReason:   rr.RejectionReason,
		CreatedAt:         rr.CreatedAt,
		UpdatedAt:         rr.UpdatedAt,
	}
	
	// Include related data if loaded
	if rr.Referral != nil {
		resp.Referral = rr.Referral.ToResponse()
	}
	if rr.User != nil {
		resp.User = rr.User.ToResponse()
	}
	if rr.Campaign != nil {
		resp.Campaign = rr.Campaign.ToResponse()
	}
	if rr.ApprovedBy != nil {
		resp.ApprovedBy = rr.ApprovedBy.ToResponse()
	}
	if rr.RejectedBy != nil {
		resp.RejectedBy = rr.RejectedBy.ToResponse()
	}
	
	return resp
}

// ToPublicResponse converts ReferralReward to a public response (limited info)
func (rr *ReferralReward) ToPublicResponse() *ReferralRewardResponse {
	return &ReferralRewardResponse{
		ID:                rr.ID,
		RewardType:        rr.RewardType,
		RewardAmount:      rr.RewardAmount,
		RewardCurrency:    rr.RewardCurrency,
		RewardDescription: rr.RewardDescription,
		Status:            rr.Status,
		EarnedAt:          rr.EarnedAt,
		PaidAt:            rr.PaidAt,
		ExpiresAt:         rr.ExpiresAt,
		ConversionType:    rr.ConversionType,
		CreatedAt:         rr.CreatedAt,
	}
}