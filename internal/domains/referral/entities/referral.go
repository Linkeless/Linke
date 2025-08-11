package entities

import (
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/referral/constants"
)

// Referral represents a referral relationship between users
type Referral struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Core Fields
	ReferrerID uint `json:"referrer_id" gorm:"not null;index"` // User who made the referral
	RefereeID  uint `json:"referee_id" gorm:"not null;index"`  // User who was referred

	// Referral Source
	InviteCodeID    *uint  `json:"invite_code_id,omitempty" gorm:"index"`         // Associated invite code
	ReferralSource  string `json:"referral_source" gorm:"size:50;not null;index"` // Source of referral (invite_code, link, email, etc.)
	ReferralChannel string `json:"referral_channel" gorm:"size:50;index"`         // Channel (organic, social, email, etc.)
	ReferralCode    string `json:"referral_code" gorm:"size:100;index"`           // Referral tracking code
	CampaignID      *uint  `json:"campaign_id,omitempty" gorm:"index"`            // Associated campaign

	// Status and Tracking
	Status        string `json:"status" gorm:"size:20;not null;default:'pending';index"`            // pending, confirmed, rewarded, cancelled
	RefereeStatus string `json:"referee_status" gorm:"size:20;not null;default:'registered';index"` // registered, activated, subscribed, churned

	// Conversion Tracking
	ConvertedAt     *time.Time `json:"converted_at,omitempty" gorm:"index"`                  // When referee converted (first purchase, activation, etc.)
	ConversionValue float64    `json:"conversion_value" gorm:"type:decimal(10,2);default:0"` // Value of conversion
	ConversionType  string     `json:"conversion_type" gorm:"size:50;index"`                 // Type of conversion (subscription, purchase, etc.)

	// Reward Tracking
	RewardStatus   string     `json:"reward_status" gorm:"size:20;not null;default:'pending';index"` // pending, earned, paid, cancelled
	RewardAmount   float64    `json:"reward_amount" gorm:"type:decimal(10,2);default:0"`             // Reward amount for referrer
	RewardCurrency string     `json:"reward_currency" gorm:"size:10;default:'USD'"`                  // Reward currency
	RefereeReward  float64    `json:"referee_reward" gorm:"type:decimal(10,2);default:0"`            // Reward for referee
	RewardedAt     *time.Time `json:"rewarded_at,omitempty" gorm:"index"`                            // When reward was paid

	// Attribution Data
	FirstClickAt *time.Time `json:"first_click_at,omitempty" gorm:"index"` // First click timestamp
	LastClickAt  *time.Time `json:"last_click_at,omitempty" gorm:"index"`  // Last click timestamp
	ClickCount   int        `json:"click_count" gorm:"default:0"`          // Number of clicks
	IPAddress    string     `json:"ip_address" gorm:"size:45"`             // IP address of referee
	UserAgent    string     `json:"user_agent" gorm:"size:500"`            // User agent
	ReferrerURL  string     `json:"referrer_url" gorm:"size:500"`          // Referrer URL
	LandingPage  string     `json:"landing_page" gorm:"size:500"`          // Landing page URL
	UTMSource    string     `json:"utm_source" gorm:"size:100"`            // UTM source
	UTMCampaign  string     `json:"utm_campaign" gorm:"size:100"`          // UTM campaign
	UTMMedium    string     `json:"utm_medium" gorm:"size:100"`            // UTM medium
	UTMTerm      string     `json:"utm_term" gorm:"size:100"`              // UTM term
	UTMContent   string     `json:"utm_content" gorm:"size:100"`           // UTM content

	// Expiration
	ExpiresAt *time.Time `json:"expires_at,omitempty" gorm:"index"` // Referral expiration date

	// Metadata
	Metadata string `json:"metadata,omitempty" gorm:"type:text"` // Additional metadata (JSON)
	Notes    string `json:"notes,omitempty" gorm:"type:text"`    // Admin notes

	// Note: Relationships removed to avoid cross-domain dependencies
	// Related data should be fetched and assembled at the application layer

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for Referral model
func (Referral) TableName() string {
	return "referrals"
}


// IsActive checks if the referral is active
func (r *Referral) IsActive() bool {
	return r.Status == constants.ReferralStatusConfirmed || r.Status == constants.ReferralStatusRewarded
}

// IsExpired checks if the referral has expired
func (r *Referral) IsExpired() bool {
	if r.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*r.ExpiresAt)
}

// CanEarnReward checks if referral can earn rewards
func (r *Referral) CanEarnReward() bool {
	return r.IsActive() && !r.IsExpired() && r.RewardStatus == constants.RewardStatusPending
}

// IsConverted checks if the referee has converted
func (r *Referral) IsConverted() bool {
	return r.ConvertedAt != nil
}


// ToResponse should be implemented in service layer to avoid import cycles
// Use dto.ToReferralResponse(r) instead

// ToPublicResponse should be implemented in service layer to avoid import cycles
// Use dto.ToReferralResponse(r) and clean sensitive data instead
