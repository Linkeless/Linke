package entities

import (
	"time"

	"gorm.io/gorm"

	"linke/internal/shared/dto"
)

// ReferralCampaign represents a referral marketing campaign
type ReferralCampaign struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Core Fields
	Name        string `json:"name" gorm:"size:100;not null;index"`        // Campaign name
	Code        string `json:"code" gorm:"uniqueIndex;size:50;not null"`   // Campaign code (unique identifier)
	Description string `json:"description" gorm:"type:text"`               // Campaign description
	
	// Campaign Settings
	CampaignType    string `json:"campaign_type" gorm:"size:50;not null;index"`      // Type of campaign (standard, bonus, seasonal, etc.)
	Status          string `json:"status" gorm:"size:20;not null;default:'active';index"` // active, paused, ended
	IsPublic        bool   `json:"is_public" gorm:"not null;default:true"`           // Whether campaign is public
	RequiresApproval bool  `json:"requires_approval" gorm:"not null;default:false"`  // Whether referrals need approval
	
	// Timing
	StartDate       *time.Time `json:"start_date,omitempty" gorm:"index"`    // Campaign start date
	EndDate         *time.Time `json:"end_date,omitempty" gorm:"index"`      // Campaign end date
	
	// Referrer Rewards
	ReferrerRewardType     string  `json:"referrer_reward_type" gorm:"size:20;not null;default:'fixed'"` // fixed, percentage, tiered
	ReferrerRewardAmount   float64 `json:"referrer_reward_amount" gorm:"type:decimal(10,2);default:0"`    // Reward amount for referrer
	ReferrerRewardCurrency string  `json:"referrer_reward_currency" gorm:"size:10;default:'USD'"`         // Reward currency
	ReferrerRewardCap      float64 `json:"referrer_reward_cap" gorm:"type:decimal(10,2);default:0"`       // Maximum reward per referrer (0 = no cap)
	
	// Referee Rewards
	RefereeRewardType     string  `json:"referee_reward_type" gorm:"size:20;not null;default:'fixed'"` // fixed, percentage, discount
	RefereeRewardAmount   float64 `json:"referee_reward_amount" gorm:"type:decimal(10,2);default:0"`    // Reward amount for referee
	RefereeRewardCurrency string  `json:"referee_reward_currency" gorm:"size:10;default:'USD'"`         // Reward currency
	
	// Reward Conditions
	MinimumPurchaseAmount float64 `json:"minimum_purchase_amount" gorm:"type:decimal(10,2);default:0"`   // Minimum purchase for reward
	RewardTrigger         string  `json:"reward_trigger" gorm:"size:50;not null;default:'registration'"` // registration, first_purchase, subscription, etc.
	RewardDelay           int     `json:"reward_delay" gorm:"default:0"`                                 // Delay in days before reward is earned
	
	// Limits
	MaxReferrals          int `json:"max_referrals" gorm:"default:0"`           // Max referrals per user (0 = unlimited)
	MaxRewards            int `json:"max_rewards" gorm:"default:0"`             // Max rewards per user (0 = unlimited)
	TotalRewardBudget     float64 `json:"total_reward_budget" gorm:"type:decimal(10,2);default:0"` // Total budget for rewards (0 = unlimited)
	
	// Targeting
	TargetAudience        string `json:"target_audience" gorm:"size:100"`        // Target audience description
	EligibleUserSegments  string `json:"eligible_user_segments" gorm:"type:text"` // JSON array of eligible user segments
	RestrictedCountries   string `json:"restricted_countries" gorm:"type:text"`   // JSON array of restricted countries
	
	// Tracking
	TrackingEnabled       bool   `json:"tracking_enabled" gorm:"not null;default:true"`    // Whether to track this campaign
	ConversionGoal        string `json:"conversion_goal" gorm:"size:100"`                  // Conversion goal (subscription, purchase, etc.)
	ConversionValue       float64 `json:"conversion_value" gorm:"type:decimal(10,2);default:0"` // Expected conversion value
	
	// Statistics (updated via background jobs)
	TotalReferrals        int     `json:"total_referrals" gorm:"default:0"`                 // Total referrals created
	TotalConversions      int     `json:"total_conversions" gorm:"default:0"`               // Total conversions
	TotalRewardsPaid      float64 `json:"total_rewards_paid" gorm:"type:decimal(10,2);default:0"` // Total rewards paid out
	ConversionRate        float64 `json:"conversion_rate" gorm:"type:decimal(5,4);default:0"`     // Conversion rate (%)
	
	// Metadata
	Metadata              string  `json:"metadata,omitempty" gorm:"type:text"`   // Additional campaign metadata (JSON)
	CreatedByID           uint    `json:"created_by_id" gorm:"not null;index"`   // Admin who created the campaign
	
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

// Campaign Type constants
const (
	CampaignTypeStandard  = "standard"
	CampaignTypeBonus     = "bonus"
	CampaignTypeSeasonal  = "seasonal"
	CampaignTypeInfluencer = "influencer"
	CampaignTypePartner   = "partner"
)

// Campaign Status constants
const (
	CampaignStatusActive = "active"
	CampaignStatusPaused = "paused"
	CampaignStatusEnded  = "ended"
)

// Reward Type constants (for campaigns)
const (
	RewardTypeFixed      = "fixed"
	RewardTypePercentage = "percentage"
	RewardTypeTiered     = "tiered"
	RewardTypeDiscount   = "discount"
)

// Reward Trigger constants
const (
	RewardTriggerRegistration   = "registration"
	RewardTriggerFirstPurchase  = "first_purchase"
	RewardTriggerSubscription   = "subscription"
	RewardTriggerActivation     = "activation"
)

// IsActive checks if the campaign is currently active
func (rc *ReferralCampaign) IsActive() bool {
	if rc.Status != CampaignStatusActive {
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
	case RewardTypeFixed:
		return rc.ReferrerRewardAmount
	case RewardTypePercentage:
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
	case RewardTypeFixed:
		return rc.RefereeRewardAmount
	case RewardTypePercentage:
		return conversionValue * (rc.RefereeRewardAmount / 100)
	case RewardTypeDiscount:
		return conversionValue * (rc.RefereeRewardAmount / 100)
	default:
		return rc.RefereeRewardAmount
	}
}

// ReferralCampaignResponse represents the campaign data structure for API responses
type ReferralCampaignResponse struct {
	ID                     uint       `json:"id" example:"1"`
	Name                   string     `json:"name" example:"Summer Referral Campaign"`
	Code                   string     `json:"code" example:"SUMMER2024"`
	Description            string     `json:"description" example:"Summer referral campaign with bonus rewards"`
	CampaignType           string     `json:"campaign_type" example:"seasonal"`
	Status                 string     `json:"status" example:"active"`
	IsPublic               bool       `json:"is_public" example:"true"`
	RequiresApproval       bool       `json:"requires_approval" example:"false"`
	StartDate              *time.Time `json:"start_date,omitempty" example:"2024-06-01T00:00:00Z"`
	EndDate                *time.Time `json:"end_date,omitempty" example:"2024-08-31T23:59:59Z"`
	ReferrerRewardType     string     `json:"referrer_reward_type" example:"fixed"`
	ReferrerRewardAmount   float64    `json:"referrer_reward_amount" example:"10.00"`
	ReferrerRewardCurrency string     `json:"referrer_reward_currency" example:"USD"`
	ReferrerRewardCap      float64    `json:"referrer_reward_cap" example:"100.00"`
	RefereeRewardType      string     `json:"referee_reward_type" example:"discount"`
	RefereeRewardAmount    float64    `json:"referee_reward_amount" example:"20.00"`
	RefereeRewardCurrency  string     `json:"referee_reward_currency" example:"USD"`
	MinimumPurchaseAmount  float64    `json:"minimum_purchase_amount" example:"25.00"`
	RewardTrigger          string     `json:"reward_trigger" example:"first_purchase"`
	RewardDelay            int        `json:"reward_delay" example:"7"`
	MaxReferrals           int        `json:"max_referrals" example:"50"`
	MaxRewards             int        `json:"max_rewards" example:"50"`
	TotalRewardBudget      float64    `json:"total_reward_budget" example:"5000.00"`
	TotalReferrals         int        `json:"total_referrals" example:"123"`
	TotalConversions       int        `json:"total_conversions" example:"45"`
	TotalRewardsPaid       float64    `json:"total_rewards_paid" example:"450.00"`
	ConversionRate         float64    `json:"conversion_rate" example:"0.3659"`
	CreatedByID            uint       `json:"created_by_id" example:"1"`
	CreatedAt              time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt              time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	
	// Optional related data
	CreatedBy              *dto.UserBasicDTO   `json:"created_by,omitempty"`
	Referrals              []*ReferralResponse `json:"referrals,omitempty"`
}

// ToResponse converts ReferralCampaign to ReferralCampaignResponse
func (rc *ReferralCampaign) ToResponse() *ReferralCampaignResponse {
	resp := &ReferralCampaignResponse{
		ID:                     rc.ID,
		Name:                   rc.Name,
		Code:                   rc.Code,
		Description:            rc.Description,
		CampaignType:           rc.CampaignType,
		Status:                 rc.Status,
		IsPublic:               rc.IsPublic,
		RequiresApproval:       rc.RequiresApproval,
		StartDate:              rc.StartDate,
		EndDate:                rc.EndDate,
		ReferrerRewardType:     rc.ReferrerRewardType,
		ReferrerRewardAmount:   rc.ReferrerRewardAmount,
		ReferrerRewardCurrency: rc.ReferrerRewardCurrency,
		ReferrerRewardCap:      rc.ReferrerRewardCap,
		RefereeRewardType:      rc.RefereeRewardType,
		RefereeRewardAmount:    rc.RefereeRewardAmount,
		RefereeRewardCurrency:  rc.RefereeRewardCurrency,
		MinimumPurchaseAmount:  rc.MinimumPurchaseAmount,
		RewardTrigger:          rc.RewardTrigger,
		RewardDelay:            rc.RewardDelay,
		MaxReferrals:           rc.MaxReferrals,
		MaxRewards:             rc.MaxRewards,
		TotalRewardBudget:      rc.TotalRewardBudget,
		TotalReferrals:         rc.TotalReferrals,
		TotalConversions:       rc.TotalConversions,
		TotalRewardsPaid:       rc.TotalRewardsPaid,
		ConversionRate:         rc.ConversionRate,
		CreatedByID:            rc.CreatedByID,
		CreatedAt:              rc.CreatedAt,
		UpdatedAt:              rc.UpdatedAt,
	}
	
	// Note: Related data should be populated at the application layer
	// to avoid cross-domain dependencies
	
	return resp
}

// ToPublicResponse converts ReferralCampaign to a public response (limited info)
func (rc *ReferralCampaign) ToPublicResponse() *ReferralCampaignResponse {
	return &ReferralCampaignResponse{
		ID:                     rc.ID,
		Name:                   rc.Name,
		Code:                   rc.Code,
		Description:            rc.Description,
		CampaignType:           rc.CampaignType,
		Status:                 rc.Status,
		StartDate:              rc.StartDate,
		EndDate:                rc.EndDate,
		ReferrerRewardType:     rc.ReferrerRewardType,
		ReferrerRewardAmount:   rc.ReferrerRewardAmount,
		ReferrerRewardCurrency: rc.ReferrerRewardCurrency,
		RefereeRewardType:      rc.RefereeRewardType,
		RefereeRewardAmount:    rc.RefereeRewardAmount,
		RefereeRewardCurrency:  rc.RefereeRewardCurrency,
		MinimumPurchaseAmount:  rc.MinimumPurchaseAmount,
		RewardTrigger:          rc.RewardTrigger,
		MaxReferrals:           rc.MaxReferrals,
	}
}