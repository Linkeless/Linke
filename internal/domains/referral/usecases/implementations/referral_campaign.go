package implementations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"linke/internal/domains/referral/entities"
	"linke/internal/shared/database"
	"linke/internal/shared/logger"
)

type ReferralCampaignService struct {
	db *database.Database
}

func NewReferralCampaignService(db *database.Database) *ReferralCampaignService {
	return &ReferralCampaignService{
		db: db,
	}
}

// CreateReferralCampaignRequest represents the request to create a referral campaign
type CreateReferralCampaignRequest struct {
	Name                   string                 `json:"name" binding:"required,max=100"`
	Code                   string                 `json:"code" binding:"required,max=50"`
	Description            string                 `json:"description,omitempty"`
	CampaignType           string                 `json:"campaign_type" binding:"required,oneof=standard bonus seasonal influencer partner"`
	IsPublic               bool                   `json:"is_public"`
	RequiresApproval       bool                   `json:"requires_approval"`
	StartDate              *time.Time             `json:"start_date,omitempty"`
	EndDate                *time.Time             `json:"end_date,omitempty"`
	ReferrerRewardType     string                 `json:"referrer_reward_type" binding:"required,oneof=fixed percentage tiered"`
	ReferrerRewardAmount   float64                `json:"referrer_reward_amount" binding:"required,min=0"`
	ReferrerRewardCurrency string                 `json:"referrer_reward_currency" binding:"required"`
	ReferrerRewardCap      float64                `json:"referrer_reward_cap,omitempty"`
	RefereeRewardType      string                 `json:"referee_reward_type" binding:"required,oneof=fixed percentage discount"`
	RefereeRewardAmount    float64                `json:"referee_reward_amount" binding:"required,min=0"`
	RefereeRewardCurrency  string                 `json:"referee_reward_currency" binding:"required"`
	MinimumPurchaseAmount  float64                `json:"minimum_purchase_amount,omitempty"`
	RewardTrigger          string                 `json:"reward_trigger" binding:"required,oneof=registration first_purchase subscription activation"`
	RewardDelay            int                    `json:"reward_delay,omitempty"`
	MaxReferrals           int                    `json:"max_referrals,omitempty"`
	MaxRewards             int                    `json:"max_rewards,omitempty"`
	TotalRewardBudget      float64                `json:"total_reward_budget,omitempty"`
	TargetAudience         string                 `json:"target_audience,omitempty"`
	EligibleUserSegments   []string               `json:"eligible_user_segments,omitempty"`
	RestrictedCountries    []string               `json:"restricted_countries,omitempty"`
	TrackingEnabled        bool                   `json:"tracking_enabled"`
	ConversionGoal         string                 `json:"conversion_goal,omitempty"`
	ConversionValue        float64                `json:"conversion_value,omitempty"`
	Metadata               map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateReferralCampaignRequest represents the request to update a referral campaign
type UpdateReferralCampaignRequest struct {
	Name                   *string                `json:"name,omitempty"`
	Description            *string                `json:"description,omitempty"`
	Status                 *string                `json:"status,omitempty"`
	IsPublic               *bool                  `json:"is_public,omitempty"`
	RequiresApproval       *bool                  `json:"requires_approval,omitempty"`
	StartDate              *time.Time             `json:"start_date,omitempty"`
	EndDate                *time.Time             `json:"end_date,omitempty"`
	ReferrerRewardType     *string                `json:"referrer_reward_type,omitempty"`
	ReferrerRewardAmount   *float64               `json:"referrer_reward_amount,omitempty"`
	ReferrerRewardCurrency *string                `json:"referrer_reward_currency,omitempty"`
	ReferrerRewardCap      *float64               `json:"referrer_reward_cap,omitempty"`
	RefereeRewardType      *string                `json:"referee_reward_type,omitempty"`
	RefereeRewardAmount    *float64               `json:"referee_reward_amount,omitempty"`
	RefereeRewardCurrency  *string                `json:"referee_reward_currency,omitempty"`
	MinimumPurchaseAmount  *float64               `json:"minimum_purchase_amount,omitempty"`
	RewardTrigger          *string                `json:"reward_trigger,omitempty"`
	RewardDelay            *int                   `json:"reward_delay,omitempty"`
	MaxReferrals           *int                   `json:"max_referrals,omitempty"`
	MaxRewards             *int                   `json:"max_rewards,omitempty"`
	TotalRewardBudget      *float64               `json:"total_reward_budget,omitempty"`
	TargetAudience         *string                `json:"target_audience,omitempty"`
	EligibleUserSegments   []string               `json:"eligible_user_segments,omitempty"`
	RestrictedCountries    []string               `json:"restricted_countries,omitempty"`
	TrackingEnabled        *bool                  `json:"tracking_enabled,omitempty"`
	ConversionGoal         *string                `json:"conversion_goal,omitempty"`
	ConversionValue        *float64               `json:"conversion_value,omitempty"`
	Metadata               map[string]interface{} `json:"metadata,omitempty"`
}

// GetReferralCampaignsRequest represents the request to get referral campaigns
type GetReferralCampaignsRequest struct {
	Status       string `form:"status,omitempty"`
	CampaignType string `form:"campaign_type,omitempty"`
	IsPublic     *bool  `form:"is_public,omitempty"`
	Limit        int    `form:"limit,omitempty"`
	Offset       int    `form:"offset,omitempty"`
}

// CreateReferralCampaign creates a new referral campaign
func (s *ReferralCampaignService) CreateReferralCampaign(ctx context.Context, adminUserID uint, req *CreateReferralCampaignRequest) (*entities.ReferralCampaign, error) {
	// Check if code already exists
	var existingCampaign entities.ReferralCampaign
	if err := s.db.DB.Where("code = ?", req.Code).First(&existingCampaign).Error; err == nil {
		return nil, fmt.Errorf("campaign with code '%s' already exists", req.Code)
	}

	// Validate dates
	if req.StartDate != nil && req.EndDate != nil && req.StartDate.After(*req.EndDate) {
		return nil, fmt.Errorf("start date cannot be after end date")
	}

	// Create campaign
	campaign := &entities.ReferralCampaign{
		Name:                   req.Name,
		Code:                   req.Code,
		Description:            req.Description,
		CampaignType:           req.CampaignType,
		Status:                 entities.CampaignStatusActive,
		IsPublic:               req.IsPublic,
		RequiresApproval:       req.RequiresApproval,
		StartDate:              req.StartDate,
		EndDate:                req.EndDate,
		ReferrerRewardType:     req.ReferrerRewardType,
		ReferrerRewardAmount:   req.ReferrerRewardAmount,
		ReferrerRewardCurrency: req.ReferrerRewardCurrency,
		ReferrerRewardCap:      req.ReferrerRewardCap,
		RefereeRewardType:      req.RefereeRewardType,
		RefereeRewardAmount:    req.RefereeRewardAmount,
		RefereeRewardCurrency:  req.RefereeRewardCurrency,
		MinimumPurchaseAmount:  req.MinimumPurchaseAmount,
		RewardTrigger:          req.RewardTrigger,
		RewardDelay:            req.RewardDelay,
		MaxReferrals:           req.MaxReferrals,
		MaxRewards:             req.MaxRewards,
		TotalRewardBudget:      req.TotalRewardBudget,
		TargetAudience:         req.TargetAudience,
		TrackingEnabled:        req.TrackingEnabled,
		ConversionGoal:         req.ConversionGoal,
		ConversionValue:        req.ConversionValue,
		CreatedByID:            adminUserID,
	}

	// Handle array fields
	if req.EligibleUserSegments != nil {
		if jsonData, err := json.Marshal(req.EligibleUserSegments); err == nil {
			campaign.EligibleUserSegments = string(jsonData)
		}
	}
	if req.RestrictedCountries != nil {
		if jsonData, err := json.Marshal(req.RestrictedCountries); err == nil {
			campaign.RestrictedCountries = string(jsonData)
		}
	}
	if req.Metadata != nil {
		if jsonData, err := json.Marshal(req.Metadata); err == nil {
			campaign.Metadata = string(jsonData)
		}
	}

	if err := s.db.DB.Create(campaign).Error; err != nil {
		return nil, fmt.Errorf("failed to create referral campaign: %w", err)
	}

	logger.Info("Referral campaign created",
		logger.Uint("campaign_id", campaign.ID),
		logger.String("campaign_code", campaign.Code),
		logger.Uint("created_by", adminUserID),
	)

	return campaign, nil
}

// GetReferralCampaigns gets referral campaigns with filters
func (s *ReferralCampaignService) GetReferralCampaigns(ctx context.Context, req *GetReferralCampaignsRequest) ([]*entities.ReferralCampaign, int64, error) {
	var campaigns []*entities.ReferralCampaign
	var total int64

	query := s.db.DB.Model(&entities.ReferralCampaign{})

	// Apply filters
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.CampaignType != "" {
		query = query.Where("campaign_type = ?", req.CampaignType)
	}
	if req.IsPublic != nil {
		query = query.Where("is_public = ?", *req.IsPublic)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count campaigns: %w", err)
	}

	// Apply pagination
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	// Get campaigns
	if err := query.Order("created_at DESC").Find(&campaigns).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get campaigns: %w", err)
	}

	return campaigns, total, nil
}

// GetReferralCampaignByID gets a referral campaign by ID
func (s *ReferralCampaignService) GetReferralCampaignByID(ctx context.Context, campaignID uint) (*entities.ReferralCampaign, error) {
	var campaign entities.ReferralCampaign
	if err := s.db.DB.First(&campaign, campaignID).Error; err != nil {
		return nil, fmt.Errorf("campaign not found: %w", err)
	}
	return &campaign, nil
}

// GetReferralCampaignByCode gets a referral campaign by code
func (s *ReferralCampaignService) GetReferralCampaignByCode(ctx context.Context, code string) (*entities.ReferralCampaign, error) {
	var campaign entities.ReferralCampaign
	if err := s.db.DB.Where("code = ?", code).First(&campaign).Error; err != nil {
		return nil, fmt.Errorf("campaign not found: %w", err)
	}
	return &campaign, nil
}

// UpdateReferralCampaign updates a referral campaign
func (s *ReferralCampaignService) UpdateReferralCampaign(ctx context.Context, campaignID uint, req *UpdateReferralCampaignRequest) (*entities.ReferralCampaign, error) {
	var campaign entities.ReferralCampaign
	if err := s.db.DB.First(&campaign, campaignID).Error; err != nil {
		return nil, fmt.Errorf("campaign not found: %w", err)
	}

	// Prepare updates
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}
	if req.RequiresApproval != nil {
		updates["requires_approval"] = *req.RequiresApproval
	}
	if req.StartDate != nil {
		updates["start_date"] = *req.StartDate
	}
	if req.EndDate != nil {
		updates["end_date"] = *req.EndDate
	}
	if req.ReferrerRewardType != nil {
		updates["referrer_reward_type"] = *req.ReferrerRewardType
	}
	if req.ReferrerRewardAmount != nil {
		updates["referrer_reward_amount"] = *req.ReferrerRewardAmount
	}
	if req.ReferrerRewardCurrency != nil {
		updates["referrer_reward_currency"] = *req.ReferrerRewardCurrency
	}
	if req.ReferrerRewardCap != nil {
		updates["referrer_reward_cap"] = *req.ReferrerRewardCap
	}
	if req.RefereeRewardType != nil {
		updates["referee_reward_type"] = *req.RefereeRewardType
	}
	if req.RefereeRewardAmount != nil {
		updates["referee_reward_amount"] = *req.RefereeRewardAmount
	}
	if req.RefereeRewardCurrency != nil {
		updates["referee_reward_currency"] = *req.RefereeRewardCurrency
	}
	if req.MinimumPurchaseAmount != nil {
		updates["minimum_purchase_amount"] = *req.MinimumPurchaseAmount
	}
	if req.RewardTrigger != nil {
		updates["reward_trigger"] = *req.RewardTrigger
	}
	if req.RewardDelay != nil {
		updates["reward_delay"] = *req.RewardDelay
	}
	if req.MaxReferrals != nil {
		updates["max_referrals"] = *req.MaxReferrals
	}
	if req.MaxRewards != nil {
		updates["max_rewards"] = *req.MaxRewards
	}
	if req.TotalRewardBudget != nil {
		updates["total_reward_budget"] = *req.TotalRewardBudget
	}
	if req.TargetAudience != nil {
		updates["target_audience"] = *req.TargetAudience
	}
	if req.TrackingEnabled != nil {
		updates["tracking_enabled"] = *req.TrackingEnabled
	}
	if req.ConversionGoal != nil {
		updates["conversion_goal"] = *req.ConversionGoal
	}
	if req.ConversionValue != nil {
		updates["conversion_value"] = *req.ConversionValue
	}

	// Handle array fields
	if req.EligibleUserSegments != nil {
		if jsonData, err := json.Marshal(req.EligibleUserSegments); err == nil {
			updates["eligible_user_segments"] = string(jsonData)
		}
	}
	if req.RestrictedCountries != nil {
		if jsonData, err := json.Marshal(req.RestrictedCountries); err == nil {
			updates["restricted_countries"] = string(jsonData)
		}
	}
	if req.Metadata != nil {
		if jsonData, err := json.Marshal(req.Metadata); err == nil {
			updates["metadata"] = string(jsonData)
		}
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.db.DB.Model(&campaign).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update campaign: %w", err)
		}
	}

	// Reload campaign
	if err := s.db.DB.First(&campaign, campaignID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload campaign: %w", err)
	}

	logger.Info("Referral campaign updated",
		logger.Uint("campaign_id", campaignID),
		logger.String("campaign_code", campaign.Code),
	)

	return &campaign, nil
}

// DeleteReferralCampaign soft deletes a referral campaign
func (s *ReferralCampaignService) DeleteReferralCampaign(ctx context.Context, campaignID uint) error {
	var campaign entities.ReferralCampaign
	if err := s.db.DB.First(&campaign, campaignID).Error; err != nil {
		return fmt.Errorf("campaign not found: %w", err)
	}

	if err := s.db.DB.Delete(&campaign).Error; err != nil {
		return fmt.Errorf("failed to delete campaign: %w", err)
	}

	logger.Info("Referral campaign deleted",
		logger.Uint("campaign_id", campaignID),
		logger.String("campaign_code", campaign.Code),
	)

	return nil
}

// GetActiveCampaigns gets all active and public referral campaigns
func (s *ReferralCampaignService) GetActiveCampaigns(ctx context.Context) ([]*entities.ReferralCampaign, error) {
	var campaigns []*entities.ReferralCampaign

	now := time.Now()
	query := s.db.DB.Where("status = ? AND is_public = ?", entities.CampaignStatusActive, true)

	// Add date filters
	query = query.Where("(start_date IS NULL OR start_date <= ?)", now)
	query = query.Where("(end_date IS NULL OR end_date >= ?)", now)

	if err := query.Order("created_at DESC").Find(&campaigns).Error; err != nil {
		return nil, fmt.Errorf("failed to get active campaigns: %w", err)
	}

	return campaigns, nil
}

// GetCampaignStats gets statistics for a campaign
func (s *ReferralCampaignService) GetCampaignStats(ctx context.Context, campaignID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get basic campaign info
	var campaign entities.ReferralCampaign
	if err := s.db.DB.First(&campaign, campaignID).Error; err != nil {
		return nil, fmt.Errorf("campaign not found: %w", err)
	}

	stats["campaign_id"] = campaign.ID
	stats["campaign_code"] = campaign.Code
	stats["campaign_name"] = campaign.Name
	stats["status"] = campaign.Status
	stats["total_referrals"] = campaign.TotalReferrals
	stats["total_conversions"] = campaign.TotalConversions
	stats["total_rewards_paid"] = campaign.TotalRewardsPaid
	stats["conversion_rate"] = campaign.ConversionRate

	// Get additional stats
	var pendingReferrals int64
	if err := s.db.DB.Model(&entities.Referral{}).Where("campaign_id = ? AND status = ?", campaignID, entities.ReferralStatusPending).Count(&pendingReferrals).Error; err != nil {
		return nil, fmt.Errorf("failed to count pending referrals: %w", err)
	}
	stats["pending_referrals"] = pendingReferrals

	var activeReferrals int64
	if err := s.db.DB.Model(&entities.Referral{}).Where("campaign_id = ? AND status = ?", campaignID, entities.ReferralStatusConfirmed).Count(&activeReferrals).Error; err != nil {
		return nil, fmt.Errorf("failed to count active referrals: %w", err)
	}
	stats["active_referrals"] = activeReferrals

	// Get reward stats
	var pendingRewards float64
	if err := s.db.DB.Model(&entities.ReferralReward{}).Where("campaign_id = ? AND status IN (?)", campaignID, []string{entities.RewardStatusPending, entities.RewardStatusEarned}).Select("COALESCE(SUM(reward_amount), 0)").Scan(&pendingRewards).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate pending rewards: %w", err)
	}
	stats["pending_rewards"] = pendingRewards

	// Calculate remaining budget
	remainingBudget := campaign.TotalRewardBudget - campaign.TotalRewardsPaid - pendingRewards
	if campaign.TotalRewardBudget > 0 {
		stats["remaining_budget"] = remainingBudget
		stats["budget_utilization"] = (campaign.TotalRewardsPaid + pendingRewards) / campaign.TotalRewardBudget
	} else {
		stats["remaining_budget"] = nil
		stats["budget_utilization"] = 0
	}

	return stats, nil
}
