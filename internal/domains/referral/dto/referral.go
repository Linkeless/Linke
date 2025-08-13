package dto

import (
	"time"

	"linke/internal/domains/referral/entities"
	"linke/internal/shared/dto"
)

// Referral DTOs

// CreateReferralRequest represents the request to create a referral
type CreateReferralRequest struct {
	ReferrerID      uint           `json:"referrer_id" binding:"required" example:"1"`
	RefereeID       uint           `json:"referee_id" binding:"required" example:"2"`
	InviteCodeID    *uint          `json:"invite_code_id,omitempty" example:"1"`
	ReferralSource  string         `json:"referral_source" binding:"required" example:"invite_code"`
	ReferralChannel string         `json:"referral_channel,omitempty" example:"organic"`
	ReferralCode    string         `json:"referral_code,omitempty" example:"REF123"`
	CampaignID      *uint          `json:"campaign_id,omitempty" example:"1"`
	ConversionValue float64        `json:"conversion_value,omitempty" example:"29.99"`
	ConversionType  string         `json:"conversion_type,omitempty" example:"subscription"`
	ExpirationDays  int            `json:"expiration_days,omitempty" example:"30"`
	AttributionData map[string]any `json:"attribution_data,omitempty"`
}

// UpdateReferralRequest represents the request to update a referral
type UpdateReferralRequest struct {
	Status          *string  `json:"status,omitempty" binding:"omitempty,oneof=pending confirmed rewarded cancelled"`
	RefereeStatus   *string  `json:"referee_status,omitempty" binding:"omitempty,oneof=registered activated subscribed churned"`
	RewardStatus    *string  `json:"reward_status,omitempty" binding:"omitempty,oneof=pending earned paid cancelled"`
	RewardAmount    *float64 `json:"reward_amount,omitempty" binding:"omitempty,min=0"`
	ConversionValue *float64 `json:"conversion_value,omitempty" binding:"omitempty,min=0"`
	ConversionType  *string  `json:"conversion_type,omitempty"`
}

// GetReferralsRequest represents the request to get referrals with filters
type GetReferralsRequest struct {
	ReferrerID   uint   `form:"referrer_id,omitempty"`
	RefereeID    uint   `form:"referee_id,omitempty"`
	Status       string `form:"status,omitempty"`
	RewardStatus string `form:"reward_status,omitempty"`
	CampaignID   *uint  `form:"campaign_id,omitempty"`
	DateFrom     string `form:"date_from,omitempty"`
	DateTo       string `form:"date_to,omitempty"`
	Limit        int    `form:"limit,omitempty"`
	Offset       int    `form:"offset,omitempty"`
}

// ApproveReferralRequest represents the request for approving referrals
type ApproveReferralRequest struct {
	RewardAmount *float64 `json:"reward_amount,omitempty" binding:"omitempty,min=0" example:"10.00"`
	Note         string   `json:"note,omitempty" example:"Approved by admin"`
}

// RejectReferralRequest represents the request for rejecting referrals
type RejectReferralRequest struct {
	Reason string `json:"reason" binding:"required" example:"Invalid referral"`
	Note   string `json:"note,omitempty" example:"Rejected due to policy violation"`
}

// PayoutReferralRequest represents the request for processing payouts
type PayoutReferralRequest struct {
	PaymentMethod string  `json:"payment_method" binding:"required" example:"paypal"`
	PaymentInfo   string  `json:"payment_info" binding:"required" example:"user@example.com"`
	Amount        float64 `json:"amount" binding:"required,min=0" example:"10.00"`
	Note          string  `json:"note,omitempty" example:"Monthly referral payout"`
}

// BulkReferralRequest represents the request for bulk operations
type BulkReferralRequest struct {
	IDs    []uint  `json:"ids" binding:"required,min=1,max=100"`
	Action string  `json:"action" binding:"required,oneof=approve reject payout"`
	Amount float64 `json:"amount,omitempty" binding:"omitempty,min=0"`
	Note   string  `json:"note,omitempty"`
}

// SearchReferralsRequest represents the search request
type SearchReferralsRequest struct {
	Query           string     `form:"q" binding:"omitempty,min=1,max=100"`
	ReferrerID      uint       `form:"referrer_id,omitempty"`
	RefereeID       uint       `form:"referee_id,omitempty"`
	Status          string     `form:"status,omitempty" binding:"omitempty,oneof=pending confirmed rewarded cancelled"`
	RefereeStatus   string     `form:"referee_status,omitempty" binding:"omitempty,oneof=registered activated subscribed churned"`
	RewardStatus    string     `form:"reward_status,omitempty" binding:"omitempty,oneof=pending earned paid cancelled"`
	CampaignID      *uint      `form:"campaign_id,omitempty"`
	ReferralSource  string     `form:"referral_source,omitempty"`
	ReferralChannel string     `form:"referral_channel,omitempty"`
	DateFrom        *time.Time `form:"date_from,omitempty"`
	DateTo          *time.Time `form:"date_to,omitempty"`
	MinReward       *float64   `form:"min_reward,omitempty" binding:"omitempty,min=0"`
	MaxReward       *float64   `form:"max_reward,omitempty" binding:"omitempty,min=0"`
	Page            int        `form:"page,omitempty" binding:"omitempty,min=1" example:"1"`
	Limit           int        `form:"limit,omitempty" binding:"omitempty,min=1,max=100" example:"10"`
}

// Referral Campaign DTOs

// CreateReferralCampaignRequest represents the request to create a referral campaign
type CreateReferralCampaignRequest struct {
	Name                 string     `json:"name" binding:"required,max=255" example:"Summer Referral Campaign"`
	Description          string     `json:"description,omitempty" binding:"max=1000" example:"Refer friends and earn rewards"`
	Type                 string     `json:"type" binding:"required,oneof=standard bonus limited" example:"standard"`
	Status               string     `json:"status,omitempty" example:"active"`
	StartDate            *time.Time `json:"start_date,omitempty" example:"2024-06-01T00:00:00Z"`
	EndDate              *time.Time `json:"end_date,omitempty" example:"2024-08-31T23:59:59Z"`
	ReferrerRewardAmount float64    `json:"referrer_reward_amount" binding:"required,min=0" example:"10.00"`
	RefereeRewardAmount  float64    `json:"referee_reward_amount,omitempty" example:"5.00"`
	RewardCurrency       string     `json:"reward_currency,omitempty" example:"CNY"`
	MaxRewards           int        `json:"max_rewards,omitempty" example:"1000"`
	MinReferrals         int        `json:"min_referrals,omitempty" example:"1"`
	MaxReferrals         int        `json:"max_referrals,omitempty" example:"10"`
	Terms                string     `json:"terms,omitempty" binding:"max=5000" example:"Terms and conditions"`
	IsPublic             bool       `json:"is_public,omitempty" example:"true"`
}

// UpdateReferralCampaignRequest represents the request to update a referral campaign
type UpdateReferralCampaignRequest struct {
	Name                 *string    `json:"name,omitempty" binding:"omitempty,max=255"`
	Description          *string    `json:"description,omitempty" binding:"omitempty,max=1000"`
	Status               *string    `json:"status,omitempty" binding:"omitempty,oneof=active inactive expired"`
	EndDate              *time.Time `json:"end_date,omitempty"`
	ReferrerRewardAmount *float64   `json:"referrer_reward_amount,omitempty" binding:"omitempty,min=0"`
	RefereeRewardAmount  *float64   `json:"referee_reward_amount,omitempty" binding:"omitempty,min=0"`
	MaxRewards           *int       `json:"max_rewards,omitempty"`
	MaxReferrals         *int       `json:"max_referrals,omitempty"`
	Terms                *string    `json:"terms,omitempty" binding:"omitempty,max=5000"`
	IsPublic             *bool      `json:"is_public,omitempty"`
}

// GetReferralCampaignsRequest represents the request to get referral campaigns with filters
type GetReferralCampaignsRequest struct {
	Status   string `form:"status,omitempty" binding:"omitempty,oneof=active inactive expired"`
	Type     string `form:"type,omitempty" binding:"omitempty,oneof=standard bonus limited"`
	IsPublic *bool  `form:"is_public,omitempty"`
	Limit    int    `form:"limit,omitempty" binding:"omitempty,min=1,max=100"`
	Offset   int    `form:"offset,omitempty" binding:"omitempty,min=0"`
}

// Invite Code DTOs

// CreateInviteCodeRequest represents the request to create an invite code
type CreateInviteCodeRequest struct {
	MaxUses              int     `json:"max_uses" binding:"min=1,max=100" example:"10"`
	Description          string  `json:"description" binding:"max=255" example:"Friend invitation code"`
	ReferralCampaignID   *uint   `json:"referral_campaign_id,omitempty" example:"1"`
	ReferralRewardAmount float64 `json:"referral_reward_amount,omitempty" example:"5.00"`
}

// UpdateInviteCodeRequest represents the request to update an invite code
type UpdateInviteCodeRequest struct {
	MaxUses              *int     `json:"max_uses,omitempty" binding:"omitempty,min=1,max=100"`
	Description          *string  `json:"description,omitempty" binding:"omitempty,max=255"`
	Status               *string  `json:"status,omitempty"`
	ReferralRewardAmount *float64 `json:"referral_reward_amount,omitempty"`
}

// GetInviteCodesRequest represents the request to get invite codes with filters
type GetInviteCodesRequest struct {
	CreatedByID        uint   `form:"created_by_id,omitempty"`
	Status             string `form:"status,omitempty"`
	ReferralCampaignID *uint  `form:"referral_campaign_id,omitempty"`
	Limit              int    `form:"limit,omitempty"`
	Offset             int    `form:"offset,omitempty"`
}

// Response DTOs

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
	RewardCurrency  string     `json:"reward_currency" example:"CNY"`
	RefereeReward   float64    `json:"referee_reward" example:"2.50"`
	RewardedAt      *time.Time `json:"rewarded_at,omitempty" example:"2024-01-01T00:00:00Z"`
	FirstClickAt    *time.Time `json:"first_click_at,omitempty" example:"2024-01-01T00:00:00Z"`
	LastClickAt     *time.Time `json:"last_click_at,omitempty" example:"2024-01-01T00:00:00Z"`
	ClickCount      int        `json:"click_count" example:"3"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z"`
	CreatedAt       time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt       time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`

	// Optional related data
	Referrer        *dto.UserBasicDTO         `json:"referrer,omitempty"`
	Referee         *dto.UserBasicDTO         `json:"referee,omitempty"`
	InviteCode      *InviteCodeResponse       `json:"invite_code,omitempty"`
	Campaign        *ReferralCampaignResponse `json:"campaign,omitempty"`
	ReferralEvents  []*ReferralEventResponse  `json:"referral_events,omitempty"`
	ReferralRewards []*ReferralRewardResponse `json:"referral_rewards,omitempty"`
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
	ReferrerRewardCurrency string     `json:"referrer_reward_currency" example:"CNY"`
	ReferrerRewardCap      float64    `json:"referrer_reward_cap" example:"100.00"`
	RefereeRewardType      string     `json:"referee_reward_type" example:"discount"`
	RefereeRewardAmount    float64    `json:"referee_reward_amount" example:"20.00"`
	RefereeRewardCurrency  string     `json:"referee_reward_currency" example:"CNY"`
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
	CreatedBy *dto.UserBasicDTO   `json:"created_by,omitempty"`
	Referrals []*ReferralResponse `json:"referrals,omitempty"`
}

// InviteCodeResponse represents the invite code data structure for API responses
type InviteCodeResponse struct {
	ID          uint      `json:"id" example:"1"`
	Code        string    `json:"code" example:"a1b2c3d4e5f6789012345678901234567890abcd"`
	CreatedByID uint      `json:"created_by_id" example:"1"`
	Status      string    `json:"status" example:"active" enums:"active,used,disabled"`
	MaxUses     int       `json:"max_uses" example:"10"`
	UsedCount   int       `json:"used_count" example:"0"`
	Description string    `json:"description" example:"Friend invitation code"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`

	// Referral Integration Fields
	ReferralCampaignID     *uint   `json:"referral_campaign_id,omitempty" example:"1"`
	ReferralRewardAmount   float64 `json:"referral_reward_amount" example:"5.00"`
	ReferralRewardCurrency string  `json:"referral_reward_currency" example:"CNY"`
}

// InviteCodeUsageResponse represents the invite code usage data structure for API responses
type InviteCodeUsageResponse struct {
	ID           uint      `json:"id" example:"1"`
	InviteCodeID uint      `json:"invite_code_id" example:"1"`
	UsedByID     uint      `json:"used_by_id" example:"2"`
	UsedAt       time.Time `json:"used_at" example:"2024-01-01T00:00:00Z"`
	IPAddress    string    `json:"ip_address" example:"192.168.1.100"`
	UserAgent    string    `json:"user_agent" example:"Mozilla/5.0..."`
	CreatedAt    time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`

	// Optional related data
	InviteCode *InviteCodeResponse `json:"invite_code,omitempty"`
	UsedBy     *dto.UserBasicDTO   `json:"used_by,omitempty"`
}

// ReferralEventResponse represents the referral event data structure for API responses
type ReferralEventResponse struct {
	ID               uint       `json:"id" example:"1"`
	ReferralID       uint       `json:"referral_id" example:"1"`
	UserID           uint       `json:"user_id" example:"2"`
	EventType        string     `json:"event_type" example:"registration"`
	EventDescription string     `json:"event_description" example:"User completed registration"`
	EventData        string     `json:"event_data,omitempty" example:"{\"signup_method\":\"email\"}"`
	IPAddress        string     `json:"ip_address" example:"192.168.1.100"`
	UserAgent        string     `json:"user_agent" example:"Mozilla/5.0..."`
	ReferrerURL      string     `json:"referrer_url" example:"https://example.com/ref"`
	PageURL          string     `json:"page_url" example:"https://example.com/signup"`
	UTMSource        string     `json:"utm_source" example:"facebook"`
	UTMCampaign      string     `json:"utm_campaign" example:"summer_referral"`
	UTMMedium        string     `json:"utm_medium" example:"social"`
	UTMTerm          string     `json:"utm_term" example:"referral"`
	UTMContent       string     `json:"utm_content" example:"banner"`
	EventValue       float64    `json:"event_value" example:"29.99"`
	EventCurrency    string     `json:"event_currency" example:"CNY"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty" example:"2024-01-01T00:00:00Z"`
	CreatedAt        time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt        time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`

	// Optional related data
	Referral *ReferralResponse `json:"referral,omitempty"`
	User     *dto.UserBasicDTO `json:"user,omitempty"`
}

// ReferralRewardResponse represents the referral reward data structure for API responses
type ReferralRewardResponse struct {
	ID                uint       `json:"id" example:"1"`
	ReferralID        uint       `json:"referral_id" example:"1"`
	UserID            uint       `json:"user_id" example:"2"`
	CampaignID        *uint      `json:"campaign_id,omitempty" example:"1"`
	RewardType        string     `json:"reward_type" example:"cash"`
	RewardAmount      float64    `json:"reward_amount" example:"10.00"`
	RewardCurrency    string     `json:"reward_currency" example:"CNY"`
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
	Referral   *ReferralResponse         `json:"referral,omitempty"`
	User       *dto.UserBasicDTO         `json:"user,omitempty"`
	Campaign   *ReferralCampaignResponse `json:"campaign,omitempty"`
	ApprovedBy *dto.UserBasicDTO         `json:"approved_by,omitempty"`
	RejectedBy *dto.UserBasicDTO         `json:"rejected_by,omitempty"`
}

// Conversion Functions

// ToResponse converts Referral to ReferralResponse
func ToReferralResponse(r *entities.Referral) *ReferralResponse {
	return &ReferralResponse{
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
}

// ToReferralCampaignResponse converts ReferralCampaign to ReferralCampaignResponse
func ToReferralCampaignResponse(rc *entities.ReferralCampaign) *ReferralCampaignResponse {
	return &ReferralCampaignResponse{
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
}

// ToInviteCodeResponse converts InviteCode to InviteCodeResponse
func ToInviteCodeResponse(ic *entities.InviteCode) *InviteCodeResponse {
	return &InviteCodeResponse{
		ID:                     ic.ID,
		Code:                   ic.Code,
		CreatedByID:            ic.CreatedByID,
		Status:                 ic.Status,
		MaxUses:                ic.MaxUses,
		UsedCount:              ic.UsedCount,
		Description:            ic.Description,
		CreatedAt:              ic.CreatedAt,
		UpdatedAt:              ic.UpdatedAt,
		ReferralCampaignID:     ic.ReferralCampaignID,
		ReferralRewardAmount:   ic.ReferralRewardAmount,
		ReferralRewardCurrency: ic.ReferralRewardCurrency,
	}
}

// ToInviteCodeUsageResponse converts InviteCodeUsage to InviteCodeUsageResponse
func ToInviteCodeUsageResponse(icu *entities.InviteCodeUsage) *InviteCodeUsageResponse {
	return &InviteCodeUsageResponse{
		ID:           icu.ID,
		InviteCodeID: icu.InviteCodeID,
		UsedByID:     icu.UsedByID,
		UsedAt:       icu.UsedAt,
		IPAddress:    icu.IPAddress,
		UserAgent:    icu.UserAgent,
		CreatedAt:    icu.CreatedAt,
	}
}

// ToReferralEventResponse converts ReferralEvent to ReferralEventResponse
func ToReferralEventResponse(re *entities.ReferralEvent) *ReferralEventResponse {
	return &ReferralEventResponse{
		ID:               re.ID,
		ReferralID:       re.ReferralID,
		UserID:           re.UserID,
		EventType:        re.EventType,
		EventDescription: re.EventDescription,
		EventData:        re.EventData,
		IPAddress:        re.IPAddress,
		UserAgent:        re.UserAgent,
		ReferrerURL:      re.ReferrerURL,
		PageURL:          re.PageURL,
		UTMSource:        re.UTMSource,
		UTMCampaign:      re.UTMCampaign,
		UTMMedium:        re.UTMMedium,
		UTMTerm:          re.UTMTerm,
		UTMContent:       re.UTMContent,
		EventValue:       re.EventValue,
		EventCurrency:    re.EventCurrency,
		ProcessedAt:      re.ProcessedAt,
		CreatedAt:        re.CreatedAt,
		UpdatedAt:        re.UpdatedAt,
	}
}

// ToReferralRewardResponse converts ReferralReward to ReferralRewardResponse
func ToReferralRewardResponse(rr *entities.ReferralReward) *ReferralRewardResponse {
	return &ReferralRewardResponse{
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
}
