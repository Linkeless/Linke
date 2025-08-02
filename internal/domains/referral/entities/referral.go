package entities

import (
	"time"

	"gorm.io/gorm"

	"linke/internal/shared/dto"
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

// Status constants
const (
	ReferralStatusPending   = "pending"
	ReferralStatusConfirmed = "confirmed"
	ReferralStatusRewarded  = "rewarded"
	ReferralStatusCancelled = "cancelled"
)

// Referee Status constants
const (
	RefereeStatusRegistered = "registered"
	RefereeStatusActivated  = "activated"
	RefereeStatusSubscribed = "subscribed"
	RefereeStatusChurned    = "churned"
)

// Reward Status constants
const (
	RewardStatusPending   = "pending"
	RewardStatusEarned    = "earned"
	RewardStatusPaid      = "paid"
	RewardStatusCancelled = "cancelled"
)

// Referral Source constants
const (
	ReferralSourceInviteCode = "invite_code"
	ReferralSourceLink       = "link"
	ReferralSourceEmail      = "email"
	ReferralSourceSocial     = "social"
	ReferralSourceOrganic    = "organic"
)

// IsActive checks if the referral is active
func (r *Referral) IsActive() bool {
	return r.Status == ReferralStatusConfirmed || r.Status == ReferralStatusRewarded
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
	return r.IsActive() && !r.IsExpired() && r.RewardStatus == RewardStatusPending
}

// IsConverted checks if the referee has converted
func (r *Referral) IsConverted() bool {
	return r.ConvertedAt != nil
}

// ReferralResponse represents the referral data structure for API responses
type ReferralResponse struct {
	ID              uint       `json:"id" example:"1"`
	ReferrerID      uint       `json:"referrer_id" example:"1"`
	RefereeID       uint       `json:"referee_id" example:"2"`
	InviteCodeID    *uint      `json:"invite_code_id,omitempty" example:"1"`
	ReferralSource  string     `json:"referral_source" example:"invite_code"`
	ReferralChannel string     `json:"referral_channel" example:"organic"`
	ReferralCode    string     `json:"referral_code" example:"REF123"`
	CampaignID      *uint      `json:"campaign_id,omitempty" example:"1"`
	Status          string     `json:"status" example:"confirmed"`
	RefereeStatus   string     `json:"referee_status" example:"activated"`
	ConvertedAt     *time.Time `json:"converted_at,omitempty" example:"2024-01-01T00:00:00Z"`
	ConversionValue float64    `json:"conversion_value" example:"29.99"`
	ConversionType  string     `json:"conversion_type" example:"subscription"`
	RewardStatus    string     `json:"reward_status" example:"earned"`
	RewardAmount    float64    `json:"reward_amount" example:"5.00"`
	RewardCurrency  string     `json:"reward_currency" example:"USD"`
	RefereeReward   float64    `json:"referee_reward" example:"2.50"`
	RewardedAt      *time.Time `json:"rewarded_at,omitempty" example:"2024-01-01T00:00:00Z"`
	FirstClickAt    *time.Time `json:"first_click_at,omitempty" example:"2024-01-01T00:00:00Z"`
	LastClickAt     *time.Time `json:"last_click_at,omitempty" example:"2024-01-01T00:00:00Z"`
	ClickCount      int        `json:"click_count" example:"3"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z"`
	CreatedAt       time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt       time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`

	// Optional related data (to be populated at application layer)
	Referrer        *dto.UserBasicDTO         `json:"referrer,omitempty"`
	Referee         *dto.UserBasicDTO         `json:"referee,omitempty"`
	InviteCode      *InviteCodeResponse       `json:"invite_code,omitempty"`
	Campaign        *ReferralCampaignResponse `json:"campaign,omitempty"`
	ReferralEvents  []*ReferralEventResponse  `json:"referral_events,omitempty"`
	ReferralRewards []*ReferralRewardResponse `json:"referral_rewards,omitempty"`
}

// ToResponse converts Referral to ReferralResponse
func (r *Referral) ToResponse() *ReferralResponse {
	resp := &ReferralResponse{
		ID:              r.ID,
		ReferrerID:      r.ReferrerID,
		RefereeID:       r.RefereeID,
		InviteCodeID:    r.InviteCodeID,
		ReferralSource:  r.ReferralSource,
		ReferralChannel: r.ReferralChannel,
		ReferralCode:    r.ReferralCode,
		CampaignID:      r.CampaignID,
		Status:          r.Status,
		RefereeStatus:   r.RefereeStatus,
		ConvertedAt:     r.ConvertedAt,
		ConversionValue: r.ConversionValue,
		ConversionType:  r.ConversionType,
		RewardStatus:    r.RewardStatus,
		RewardAmount:    r.RewardAmount,
		RewardCurrency:  r.RewardCurrency,
		RefereeReward:   r.RefereeReward,
		RewardedAt:      r.RewardedAt,
		FirstClickAt:    r.FirstClickAt,
		LastClickAt:     r.LastClickAt,
		ClickCount:      r.ClickCount,
		ExpiresAt:       r.ExpiresAt,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}

	// Note: Related data should be populated at the application layer
	// to avoid cross-domain dependencies

	return resp
}

// ToPublicResponse converts Referral to a public response (limited info)
func (r *Referral) ToPublicResponse() *ReferralResponse {
	return &ReferralResponse{
		ID:              r.ID,
		ReferralSource:  r.ReferralSource,
		ReferralChannel: r.ReferralChannel,
		Status:          r.Status,
		RefereeStatus:   r.RefereeStatus,
		ConvertedAt:     r.ConvertedAt,
		ConversionType:  r.ConversionType,
		RewardStatus:    r.RewardStatus,
		CreatedAt:       r.CreatedAt,
	}
}
