package implementations

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"linke/internal/domains/referral/entities"
	"linke/internal/shared/database"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

type ReferralService struct {
	db *database.Database
}

func NewReferralService(db *database.Database) *ReferralService {
	return &ReferralService{
		db: db,
	}
}

// CreateReferralRequest represents the request to create a referral
type CreateReferralRequest struct {
	ReferrerID      uint                   `json:"referrer_id" binding:"required"`
	RefereeID       uint                   `json:"referee_id" binding:"required"`
	InviteCodeID    *uint                  `json:"invite_code_id,omitempty"`
	ReferralSource  string                 `json:"referral_source" binding:"required"`
	ReferralChannel string                 `json:"referral_channel,omitempty"`
	ReferralCode    string                 `json:"referral_code,omitempty"`
	CampaignID      *uint                  `json:"campaign_id,omitempty"`
	AttributionData map[string]interface{} `json:"attribution_data,omitempty"`
	ConversionValue float64                `json:"conversion_value,omitempty"`
	ConversionType  string                 `json:"conversion_type,omitempty"`
	ExpirationDays  int                    `json:"expiration_days,omitempty"`
}

// CreateReferral creates a new referral relationship
func (s *ReferralService) CreateReferral(ctx context.Context, req *CreateReferralRequest) (*entities.Referral, error) {
	// Validate that referrer and referee are different
	if req.ReferrerID == req.RefereeID {
		return nil, fmt.Errorf("referrer and referee cannot be the same user")
	}

	// Check if referral already exists
	var existingReferral entities.Referral
	result := s.db.DB.Where("referrer_id = ? AND referee_id = ?", req.ReferrerID, req.RefereeID).First(&existingReferral)
	if result.Error == nil {
		return nil, fmt.Errorf("referral already exists between these users")
	}

	// Generate referral code if not provided
	referralCode := req.ReferralCode
	if referralCode == "" {
		referralCode = s.generateReferralCode()
	}

	// Create referral
	referral := &entities.Referral{
		ReferrerID:      req.ReferrerID,
		RefereeID:       req.RefereeID,
		InviteCodeID:    req.InviteCodeID,
		ReferralSource:  req.ReferralSource,
		ReferralChannel: req.ReferralChannel,
		ReferralCode:    referralCode,
		CampaignID:      req.CampaignID,
		Status:          entities.ReferralStatusPending,
		RefereeStatus:   entities.RefereeStatusRegistered,
		RewardStatus:    entities.RewardStatusPending,
		RewardCurrency:  "USD",
		ConversionValue: req.ConversionValue,
		ConversionType:  req.ConversionType,
	}

	// Handle attribution data
	if req.AttributionData != nil {
		if ipAddress, ok := req.AttributionData["ip_address"].(string); ok {
			referral.IPAddress = ipAddress
		}
		if userAgent, ok := req.AttributionData["user_agent"].(string); ok {
			referral.UserAgent = userAgent
		}
		if referrerURL, ok := req.AttributionData["referrer_url"].(string); ok {
			referral.ReferrerURL = referrerURL
		}
		if landingPage, ok := req.AttributionData["landing_page"].(string); ok {
			referral.LandingPage = landingPage
		}
		if utmSource, ok := req.AttributionData["utm_source"].(string); ok {
			referral.UTMSource = utmSource
		}
		if utmCampaign, ok := req.AttributionData["utm_campaign"].(string); ok {
			referral.UTMCampaign = utmCampaign
		}
		if utmMedium, ok := req.AttributionData["utm_medium"].(string); ok {
			referral.UTMMedium = utmMedium
		}
		if utmTerm, ok := req.AttributionData["utm_term"].(string); ok {
			referral.UTMTerm = utmTerm
		}
		if utmContent, ok := req.AttributionData["utm_content"].(string); ok {
			referral.UTMContent = utmContent
		}
	}

	// Set expiration if specified
	if req.ExpirationDays > 0 {
		expiresAt := time.Now().AddDate(0, 0, req.ExpirationDays)
		referral.ExpiresAt = &expiresAt
	}

	// Save referral
	if err := s.db.DB.Create(referral).Error; err != nil {
		return nil, fmt.Errorf("failed to create referral: %w", err)
	}

	// Log referral creation event
	s.createReferralEvent(ctx, referral.ID, req.RefereeID, entities.EventTypeRegistration,
		"User registered via referral", req.AttributionData)

	logger.Info("Referral created successfully",
		logger.Uint("referral_id", referral.ID),
		logger.Uint("referrer_id", req.ReferrerID),
		logger.Uint("referee_id", req.RefereeID),
		logger.String("source", req.ReferralSource),
	)

	return referral, nil
}

// CreateReferralFromInviteCode creates a referral when an invite code is used
func (s *ReferralService) CreateReferralFromInviteCode(ctx context.Context, inviteCode *entities.InviteCode, refereeID uint, attributionData map[string]interface{}) (*entities.Referral, error) {
	// Use the invite code's campaign ID if available, otherwise use default campaign
	var campaignID *uint
	if inviteCode.ReferralCampaignID != nil {
		campaignID = inviteCode.ReferralCampaignID
	} else {
		var defaultCampaign entities.ReferralCampaign
		if err := s.db.DB.Where("code = ? AND status = ?", "DEFAULT", entities.CampaignStatusActive).First(&defaultCampaign).Error; err == nil {
			campaignID = &defaultCampaign.ID
		}
	}

	req := &CreateReferralRequest{
		ReferrerID:      inviteCode.CreatedByID,
		RefereeID:       refereeID,
		InviteCodeID:    &inviteCode.ID,
		ReferralSource:  entities.ReferralSourceInviteCode,
		ReferralChannel: "organic",
		ReferralCode:    inviteCode.Code,
		CampaignID:      campaignID,
		AttributionData: attributionData,
		ConversionType:  "registration",
		ExpirationDays:  90, // Default 90 days expiration
	}

	return s.CreateReferral(ctx, req)
}

// TrackReferralClick tracks a click on a referral link
func (s *ReferralService) TrackReferralClick(ctx context.Context, referralCode string, attributionData map[string]interface{}) error {
	// Find referral by code
	var referral entities.Referral
	if err := s.db.DB.Where("referral_code = ?", referralCode).First(&referral).Error; err != nil {
		return fmt.Errorf("referral not found: %w", err)
	}

	// Update click tracking
	now := time.Now()
	updates := map[string]interface{}{
		"click_count":   gorm.Expr("click_count + 1"),
		"last_click_at": now,
		"updated_at":    now,
	}

	if referral.FirstClickAt == nil {
		updates["first_click_at"] = now
	}

	if err := s.db.DB.Model(&referral).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update referral click tracking: %w", err)
	}

	// Create click event
	s.createReferralEvent(ctx, referral.ID, referral.RefereeID, entities.EventTypeClick,
		"Referral link clicked", attributionData)

	return nil
}

// ConfirmReferral confirms a referral and marks it as active
func (s *ReferralService) ConfirmReferral(ctx context.Context, referralID uint) error {
	var referral entities.Referral
	if err := s.db.DB.First(&referral, referralID).Error; err != nil {
		return fmt.Errorf("referral not found: %w", err)
	}

	if referral.Status != entities.ReferralStatusPending {
		return fmt.Errorf("referral is not in pending status")
	}

	// Update referral status
	if err := s.db.DB.Model(&referral).Updates(map[string]interface{}{
		"status":     entities.ReferralStatusConfirmed,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to confirm referral: %w", err)
	}

	// Create confirmation event
	s.createReferralEvent(ctx, referral.ID, referral.RefereeID, entities.EventTypeActivation,
		"Referral confirmed", nil)

	logger.Info("Referral confirmed",
		logger.Uint("referral_id", referralID),
		logger.Uint("referrer_id", referral.ReferrerID),
		logger.Uint("referee_id", referral.RefereeID),
	)

	return nil
}

// TrackConversion tracks a conversion event for a referral
func (s *ReferralService) TrackConversion(ctx context.Context, userID uint, conversionType string, conversionValue float64, conversionID *uint) error {
	// Find referral where user is the referee
	var referral entities.Referral
	if err := s.db.DB.Where("referee_id = ? AND status = ?", userID, entities.ReferralStatusConfirmed).First(&referral).Error; err != nil {
		// No active referral found, this is normal
		return nil
	}

	// Check if already converted
	if referral.ConvertedAt != nil {
		return nil // Already converted
	}

	now := time.Now()

	// Update referral with conversion data
	updates := map[string]interface{}{
		"converted_at":     now,
		"conversion_value": conversionValue,
		"conversion_type":  conversionType,
		"referee_status":   entities.RefereeStatusActivated,
		"updated_at":       now,
	}

	if conversionType == "subscription" {
		updates["referee_status"] = entities.RefereeStatusSubscribed
	}

	if err := s.db.DB.Model(&referral).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to track conversion: %w", err)
	}

	// Create conversion event
	eventData := map[string]interface{}{
		"conversion_type":  conversionType,
		"conversion_value": conversionValue,
	}
	if conversionID != nil {
		eventData["conversion_id"] = *conversionID
	}

	s.createReferralEvent(ctx, referral.ID, userID, entities.EventTypeConversion,
		fmt.Sprintf("User converted: %s", conversionType), eventData)

	// Process rewards if applicable
	go s.processReferralRewards(context.Background(), referral.ID, conversionValue)

	logger.Info("Conversion tracked",
		logger.Uint("referral_id", referral.ID),
		logger.Uint("user_id", userID),
		logger.String("conversion_type", conversionType),
		logger.Any("conversion_value", conversionValue),
	)

	return nil
}

// GetReferralsByReferrer gets referrals created by a specific referrer
func (s *ReferralService) GetReferralsByReferrer(ctx context.Context, referrerID uint, limit, offset int) ([]*entities.Referral, int64, error) {
	var referrals []*entities.Referral
	var total int64

	query := s.db.DB.Where("referrer_id = ?", referrerID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count referrals: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&referrals).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get referrals: %w", err)
	}

	return referrals, total, nil
}

// GetReferralsByReferee gets referrals where user is the referee
func (s *ReferralService) GetReferralsByReferee(ctx context.Context, refereeID uint, limit, offset int) ([]*entities.Referral, int64, error) {
	var referrals []*entities.Referral
	var total int64

	query := s.db.DB.Where("referee_id = ?", refereeID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count referrals: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&referrals).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get referrals: %w", err)
	}

	return referrals, total, nil
}

// GetReferralByID gets a referral by ID
func (s *ReferralService) GetReferralByID(ctx context.Context, referralID uint) (*entities.Referral, error) {
	var referral entities.Referral
	if err := s.db.DB.First(&referral, referralID).Error; err != nil {
		return nil, fmt.Errorf("referral not found: %w", err)
	}
	return &referral, nil
}

// GetReferralWithRelations gets a referral with its related data
func (s *ReferralService) GetReferralWithRelations(ctx context.Context, referralID uint) (*entities.Referral, error) {
	var referral entities.Referral
	if err := s.db.DB.First(&referral, referralID).Error; err != nil {
		return nil, fmt.Errorf("referral not found: %w", err)
	}

	// Load related data
	if err := s.loadReferralRelations(&referral); err != nil {
		return nil, fmt.Errorf("failed to load referral relations: %w", err)
	}

	return &referral, nil
}

// GetReferralStats gets referral statistics for a user
func (s *ReferralService) GetReferralStats(ctx context.Context, userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count total referrals made by user
	var totalReferrals int64
	if err := s.db.DB.Model(&entities.Referral{}).Where("referrer_id = ?", userID).Count(&totalReferrals).Error; err != nil {
		return nil, fmt.Errorf("failed to count referrals: %w", err)
	}
	stats["total_referrals"] = totalReferrals

	// Count conversions
	var totalConversions int64
	if err := s.db.DB.Model(&entities.Referral{}).Where("referrer_id = ? AND converted_at IS NOT NULL", userID).Count(&totalConversions).Error; err != nil {
		return nil, fmt.Errorf("failed to count conversions: %w", err)
	}
	stats["total_conversions"] = totalConversions

	// Calculate conversion rate
	conversionRate := float64(0)
	if totalReferrals > 0 {
		conversionRate = float64(totalConversions) / float64(totalReferrals)
	}
	stats["conversion_rate"] = conversionRate

	// Total rewards earned
	var totalRewards float64
	if err := s.db.DB.Model(&entities.ReferralReward{}).Where("user_id = ? AND status = ?", userID, entities.RewardStatusPaid).Select("COALESCE(SUM(reward_amount), 0)").Scan(&totalRewards).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate total rewards: %w", err)
	}
	stats["total_rewards"] = totalRewards

	// Pending rewards
	var pendingRewards float64
	if err := s.db.DB.Model(&entities.ReferralReward{}).Where("user_id = ? AND status IN (?)", userID, []string{entities.RewardStatusPending, entities.RewardStatusEarned}).Select("COALESCE(SUM(reward_amount), 0)").Scan(&pendingRewards).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate pending rewards: %w", err)
	}
	stats["pending_rewards"] = pendingRewards

	return stats, nil
}

// generateReferralCode generates a unique referral code
func (s *ReferralService) generateReferralCode() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return strings.ToUpper(hex.EncodeToString(bytes))
}

// createReferralEvent creates a referral event
func (s *ReferralService) createReferralEvent(ctx context.Context, referralID, userID uint, eventType, description string, eventData map[string]interface{}) {
	event := &entities.ReferralEvent{
		ReferralID:       referralID,
		UserID:           userID,
		EventType:        eventType,
		EventDescription: description,
	}

	if eventData != nil {
		// Convert eventData to JSON string
		if jsonData, err := json.Marshal(eventData); err == nil {
			event.EventData = string(jsonData)
		}

		// Extract attribution data
		if ipAddress, ok := eventData["ip_address"].(string); ok {
			event.IPAddress = ipAddress
		}
		if userAgent, ok := eventData["user_agent"].(string); ok {
			event.UserAgent = userAgent
		}
		if referrerURL, ok := eventData["referrer_url"].(string); ok {
			event.ReferrerURL = referrerURL
		}
		if pageURL, ok := eventData["page_url"].(string); ok {
			event.PageURL = pageURL
		}
		if utmSource, ok := eventData["utm_source"].(string); ok {
			event.UTMSource = utmSource
		}
		if utmCampaign, ok := eventData["utm_campaign"].(string); ok {
			event.UTMCampaign = utmCampaign
		}
		if utmMedium, ok := eventData["utm_medium"].(string); ok {
			event.UTMMedium = utmMedium
		}
		if utmTerm, ok := eventData["utm_term"].(string); ok {
			event.UTMTerm = utmTerm
		}
		if utmContent, ok := eventData["utm_content"].(string); ok {
			event.UTMContent = utmContent
		}
	}

	if err := s.db.DB.Create(event).Error; err != nil {
		logger.Error("Failed to create referral event",
			logger.Uint("referral_id", referralID),
			logger.Uint("user_id", userID),
			logger.String("event_type", eventType),
			logger.Error2("error", err),
		)
	}
}

// loadReferralRelations loads related data for a referral
func (s *ReferralService) loadReferralRelations(referral *entities.Referral) error {
	// Note: Cross-domain entity loading has been removed to maintain clean architecture
	// Related data (User, InviteCode, Campaign) should be loaded and assembled at the application layer
	// This maintains domain boundaries and prevents circular dependencies

	// TODO: The following related data loading should be handled at the application layer:
	// - Referrer user data (via User domain service)
	// - Referee user data (via User domain service)
	// - Invite code data (if InviteCodeID is not nil)
	// - Campaign data (if CampaignID is not nil)

	logger.Info("Referral relations should be loaded at application layer",
		logger.Uint("referral_id", referral.ID),
		logger.Uint("referrer_id", referral.ReferrerID),
		logger.Uint("referee_id", referral.RefereeID),
	)

	return nil
}

// processReferralRewards processes rewards for a converted referral
func (s *ReferralService) processReferralRewards(ctx context.Context, referralID uint, conversionValue float64) {
	// This would be implemented to create and process rewards
	// For now, just log the action
	logger.Info("Processing referral rewards",
		logger.Uint("referral_id", referralID),
		logger.Any("conversion_value", conversionValue),
	)
}
