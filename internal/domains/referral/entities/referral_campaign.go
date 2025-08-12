package entities

import (
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/referral/constants"
)

// ReferralCampaign represents a referral marketing campaign
type ReferralCampaign struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Core Fields
	Name        string `json:"name" gorm:"size:100;not null;index"`      // Campaign name
	Code        string `json:"code" gorm:"uniqueIndex;size:50;not null"` // Campaign code (unique identifier)
	Description string `json:"description" gorm:"type:text"`             // Campaign description

	// Campaign Settings
	CampaignType     string `json:"campaign_type" gorm:"size:50;not null;index"`           // Type of campaign (standard, bonus, seasonal, etc.)
	Status           string `json:"status" gorm:"size:20;not null;default:'active';index"` // active, paused, ended
	IsPublic         bool   `json:"is_public" gorm:"not null;default:true"`                // Whether campaign is public
	RequiresApproval bool   `json:"requires_approval" gorm:"not null;default:false"`       // Whether referrals need approval

	// Timing
	StartDate *time.Time `json:"start_date,omitempty" gorm:"index"` // Campaign start date
	EndDate   *time.Time `json:"end_date,omitempty" gorm:"index"`   // Campaign end date

	// Referrer Rewards
	ReferrerRewardType     string  `json:"referrer_reward_type" gorm:"size:20;not null;default:'fixed'"` // fixed, percentage, tiered
	ReferrerRewardAmount   float64 `json:"referrer_reward_amount" gorm:"type:decimal(10,2);default:0"`   // Reward amount for referrer
	ReferrerRewardCurrency string  `json:"referrer_reward_currency" gorm:"size:10;default:'CNY'"`        // Reward currency
	ReferrerRewardCap      float64 `json:"referrer_reward_cap" gorm:"type:decimal(10,2);default:0"`      // Maximum reward per referrer (0 = no cap)

	// Referee Rewards
	RefereeRewardType     string  `json:"referee_reward_type" gorm:"size:20;not null;default:'fixed'"` // fixed, percentage, discount
	RefereeRewardAmount   float64 `json:"referee_reward_amount" gorm:"type:decimal(10,2);default:0"`   // Reward amount for referee
	RefereeRewardCurrency string  `json:"referee_reward_currency" gorm:"size:10;default:'CNY'"`        // Reward currency

	// Reward Conditions
	MinimumPurchaseAmount float64 `json:"minimum_purchase_amount" gorm:"type:decimal(10,2);default:0"`   // Minimum purchase for reward
	RewardTrigger         string  `json:"reward_trigger" gorm:"size:50;not null;default:'registration'"` // registration, first_purchase, subscription, etc.
	RewardDelay           int     `json:"reward_delay" gorm:"default:0"`                                 // Delay in days before reward is earned

	// Limits
	MaxReferrals      int     `json:"max_referrals" gorm:"default:0"`                          // Max referrals per user (0 = unlimited)
	MaxRewards        int     `json:"max_rewards" gorm:"default:0"`                            // Max rewards per user (0 = unlimited)
	TotalRewardBudget float64 `json:"total_reward_budget" gorm:"type:decimal(10,2);default:0"` // Total budget for rewards (0 = unlimited)

	// Targeting
	TargetAudience       string `json:"target_audience" gorm:"size:100"`         // Target audience description
	EligibleUserSegments string `json:"eligible_user_segments" gorm:"type:text"` // JSON array of eligible user segments
	RestrictedCountries  string `json:"restricted_countries" gorm:"type:text"`   // JSON array of restricted countries

	// Tracking
	TrackingEnabled bool    `json:"tracking_enabled" gorm:"not null;default:true"`        // Whether to track this campaign
	ConversionGoal  string  `json:"conversion_goal" gorm:"size:100"`                      // Conversion goal (subscription, purchase, etc.)
	ConversionValue float64 `json:"conversion_value" gorm:"type:decimal(10,2);default:0"` // Expected conversion value

	// Statistics (updated via background jobs)
	TotalReferrals   int     `json:"total_referrals" gorm:"default:0"`                       // Total referrals created
	TotalConversions int     `json:"total_conversions" gorm:"default:0"`                     // Total conversions
	TotalRewardsPaid float64 `json:"total_rewards_paid" gorm:"type:decimal(10,2);default:0"` // Total rewards paid out
	ConversionRate   float64 `json:"conversion_rate" gorm:"type:decimal(5,4);default:0"`     // Conversion rate (%)

	// Metadata
	Metadata    string `json:"metadata,omitempty" gorm:"type:text"` // Additional campaign metadata (JSON)
	CreatedByID uint   `json:"created_by_id" gorm:"not null;index"` // Admin who created the campaign

	// Note: Relationships removed to avoid cross-domain dependencies
	// Related data should be fetched and assembled at the application layer

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for ReferralCampaign model
func (ReferralCampaign) TableName() string {
	return "referral_campaigns"
}


// IsActive checks if the campaign is currently active
func (rc *ReferralCampaign) IsActive() bool {
	if rc.Status != constants.CampaignStatusActive {
		return false
	}

	now := time.Now()

	// Check start date
	if rc.StartDate != nil && now.Before(*rc.StartDate) {
		return false
	}

	// Check end date
	if rc.EndDate != nil && now.After(*rc.EndDate) {
		return false
	}

	return true
}

// IsExpired checks if the campaign has expired
func (rc *ReferralCampaign) IsExpired() bool {
	if rc.EndDate == nil {
		return false
	}
	return time.Now().After(*rc.EndDate)
}

// HasBudgetRemaining checks if campaign has budget remaining
func (rc *ReferralCampaign) HasBudgetRemaining() bool {
	if rc.TotalRewardBudget <= 0 {
		return true // No budget limit
	}
	return rc.TotalRewardsPaid < rc.TotalRewardBudget
}

// CalculateReferrerReward calculates the referrer reward based on campaign settings
func (rc *ReferralCampaign) CalculateReferrerReward(conversionValue float64) float64 {
	switch rc.ReferrerRewardType {
	case constants.RewardTypeFixed:
		return rc.ReferrerRewardAmount
	case constants.RewardTypePercentage:
		reward := conversionValue * (rc.ReferrerRewardAmount / 100)
		if rc.ReferrerRewardCap > 0 && reward > rc.ReferrerRewardCap {
			return rc.ReferrerRewardCap
		}
		return reward
	default:
		return rc.ReferrerRewardAmount
	}
}

// CalculateRefereeReward calculates the referee reward based on campaign settings
func (rc *ReferralCampaign) CalculateRefereeReward(conversionValue float64) float64 {
	switch rc.RefereeRewardType {
	case constants.RewardTypeFixed:
		return rc.RefereeRewardAmount
	case constants.RewardTypePercentage:
		return conversionValue * (rc.RefereeRewardAmount / 100)
	case constants.RewardTypeDiscount:
		return conversionValue * (rc.RefereeRewardAmount / 100)
	default:
		return rc.RefereeRewardAmount
	}
}


// ToResponse should be implemented in service layer to avoid import cycles
// Use dto.ToReferralCampaignResponse(rc) instead

// ToPublicResponse should be implemented in service layer to avoid import cycles
// Use dto.ToReferralCampaignResponse(rc) and clean sensitive data instead
