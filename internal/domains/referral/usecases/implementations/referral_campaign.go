package implementations

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/referral/entities"
	"linke/internal/domains/referral/usecases/interfaces"
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


// CreateReferralCampaign creates a new referral campaign
func (s *ReferralCampaignService) CreateReferralCampaign(ctx context.Context, req *interfaces.CreateReferralCampaignRequest) (*entities.ReferralCampaign, error) {
	// Check if name already exists
	var existingCampaign entities.ReferralCampaign
	if err := s.db.DB.Where("name = ?", req.Name).First(&existingCampaign).Error; err == nil {
		return nil, fmt.Errorf("campaign with name '%s' already exists", req.Name)
	}

	// Validate dates
	if req.StartDate != nil && req.EndDate != nil && req.StartDate.After(*req.EndDate) {
		return nil, fmt.Errorf("start date cannot be after end date")
	}

	// Create campaign
	campaign := &entities.ReferralCampaign{
		Name:                   req.Name,
		Description:            req.Description,
		CampaignType:           req.Type,
		Status:                 entities.CampaignStatusActive,
		StartDate:              req.StartDate,
		EndDate:                req.EndDate,
		ReferrerRewardAmount:   req.ReferrerRewardAmount,
		ReferrerRewardCurrency: req.RewardCurrency,
		RefereeRewardAmount:    req.RefereeRewardAmount,
		RefereeRewardCurrency:  req.RewardCurrency,
		MaxRewards:             req.MaxRewards,
		MaxReferrals:           req.MaxReferrals,
		IsPublic:               req.IsPublic,
	}


	if err := s.db.DB.Create(campaign).Error; err != nil {
		return nil, fmt.Errorf("failed to create referral campaign: %w", err)
	}

	logger.Info("Referral campaign created",
		logger.Uint("campaign_id", campaign.ID),
		logger.String("campaign_name", campaign.Name),
	)

	return campaign, nil
}

// GetReferralCampaigns gets referral campaigns with filters
func (s *ReferralCampaignService) GetReferralCampaigns(ctx context.Context, req *interfaces.GetReferralCampaignsRequest) ([]*entities.ReferralCampaign, int64, error) {
	var campaigns []*entities.ReferralCampaign
	var total int64

	query := s.db.DB.Model(&entities.ReferralCampaign{})

	// Apply filters
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Type != "" {
		query = query.Where("campaign_type = ?", req.Type)
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

// GetReferralCampaign gets a referral campaign by ID
func (s *ReferralCampaignService) GetReferralCampaign(ctx context.Context, campaignID uint) (*entities.ReferralCampaign, error) {
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
func (s *ReferralCampaignService) UpdateReferralCampaign(ctx context.Context, campaignID uint, req *interfaces.UpdateReferralCampaignRequest) (*entities.ReferralCampaign, error) {
	var campaign entities.ReferralCampaign
	if err := s.db.DB.First(&campaign, campaignID).Error; err != nil {
		return nil, fmt.Errorf("campaign not found: %w", err)
	}

	// Prepare updates
	updates := make(map[string]any)

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.EndDate != nil {
		updates["end_date"] = *req.EndDate
	}
	if req.ReferrerRewardAmount != nil {
		updates["referrer_reward_amount"] = *req.ReferrerRewardAmount
	}
	if req.RefereeRewardAmount != nil {
		updates["referee_reward_amount"] = *req.RefereeRewardAmount
	}
	if req.MaxRewards != nil {
		updates["max_rewards"] = *req.MaxRewards
	}
	if req.MaxReferrals != nil {
		updates["max_referrals"] = *req.MaxReferrals
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
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
		logger.String("campaign_name", campaign.Name),
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
		logger.String("campaign_name", campaign.Name),
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

// GetCampaignStatistics gets statistics for a campaign
func (s *ReferralCampaignService) GetCampaignStatistics(ctx context.Context, campaignID uint) (map[string]any, error) {
	stats := make(map[string]any)

	// Get basic campaign info
	var campaign entities.ReferralCampaign
	if err := s.db.DB.First(&campaign, campaignID).Error; err != nil {
		return nil, fmt.Errorf("campaign not found: %w", err)
	}

	stats["campaign_id"] = campaign.ID
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

// ActivateCampaign activates a referral campaign
func (s *ReferralCampaignService) ActivateCampaign(ctx context.Context, campaignID uint) error {
	updates := map[string]any{
		"status":     entities.CampaignStatusActive,
		"updated_at": time.Now(),
	}

	if err := s.db.DB.WithContext(ctx).Model(&entities.ReferralCampaign{}).Where("id = ?", campaignID).Updates(updates).Error; err != nil {
		logger.Error("Failed to activate campaign", logger.Error2("error", err), logger.Uint("campaign_id", campaignID))
		return fmt.Errorf("failed to activate campaign: %w", err)
	}

	logger.Info("Campaign activated", logger.Uint("campaign_id", campaignID))
	return nil
}

// DeactivateCampaign deactivates a referral campaign
func (s *ReferralCampaignService) DeactivateCampaign(ctx context.Context, campaignID uint) error {
	updates := map[string]any{
		"status":     entities.CampaignStatusPaused,
		"updated_at": time.Now(),
	}

	if err := s.db.DB.WithContext(ctx).Model(&entities.ReferralCampaign{}).Where("id = ?", campaignID).Updates(updates).Error; err != nil {
		logger.Error("Failed to deactivate campaign", logger.Error2("error", err), logger.Uint("campaign_id", campaignID))
		return fmt.Errorf("failed to deactivate campaign: %w", err)
	}

	logger.Info("Campaign deactivated", logger.Uint("campaign_id", campaignID))
	return nil
}

// ExpireCampaign expires a referral campaign
func (s *ReferralCampaignService) ExpireCampaign(ctx context.Context, campaignID uint) error {
	updates := map[string]any{
		"status":     entities.CampaignStatusEnded,
		"updated_at": time.Now(),
	}

	if err := s.db.DB.WithContext(ctx).Model(&entities.ReferralCampaign{}).Where("id = ?", campaignID).Updates(updates).Error; err != nil {
		logger.Error("Failed to expire campaign", logger.Error2("error", err), logger.Uint("campaign_id", campaignID))
		return fmt.Errorf("failed to expire campaign: %w", err)
	}

	logger.Info("Campaign expired", logger.Uint("campaign_id", campaignID))
	return nil
}

// GetCampaignPerformance gets campaign performance metrics for a date range
func (s *ReferralCampaignService) GetCampaignPerformance(ctx context.Context, campaignID uint, fromDate, toDate time.Time) (map[string]any, error) {
	stats := make(map[string]any)

	// Get campaign info
	var campaign entities.ReferralCampaign
	if err := s.db.DB.First(&campaign, campaignID).Error; err != nil {
		return nil, fmt.Errorf("campaign not found: %w", err)
	}

	stats["campaign_id"] = campaign.ID
	stats["campaign_name"] = campaign.Name
	stats["from_date"] = fromDate
	stats["to_date"] = toDate

	// Referrals created in date range
	var referralsInPeriod int64
	if err := s.db.DB.Model(&entities.Referral{}).
		Where("campaign_id = ? AND created_at BETWEEN ? AND ?", campaignID, fromDate, toDate).
		Count(&referralsInPeriod).Error; err != nil {
		return nil, fmt.Errorf("failed to count referrals in period: %w", err)
	}
	stats["referrals_in_period"] = referralsInPeriod

	// Conversions in date range
	var conversionsInPeriod int64
	if err := s.db.DB.Model(&entities.Referral{}).
		Where("campaign_id = ? AND converted_at BETWEEN ? AND ?", campaignID, fromDate, toDate).
		Count(&conversionsInPeriod).Error; err != nil {
		return nil, fmt.Errorf("failed to count conversions in period: %w", err)
	}
	stats["conversions_in_period"] = conversionsInPeriod

	// Conversion rate for the period
	conversionRate := float64(0)
	if referralsInPeriod > 0 {
		conversionRate = float64(conversionsInPeriod) / float64(referralsInPeriod)
	}
	stats["conversion_rate_in_period"] = conversionRate

	// Rewards paid in date range
	var rewardsPaidInPeriod float64
	if err := s.db.DB.Model(&entities.ReferralReward{}).
		Where("campaign_id = ? AND paid_at BETWEEN ? AND ?", campaignID, fromDate, toDate).
		Select("COALESCE(SUM(reward_amount), 0)").
		Scan(&rewardsPaidInPeriod).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate rewards paid in period: %w", err)
	}
	stats["rewards_paid_in_period"] = rewardsPaidInPeriod

	return stats, nil
}
