package service

import (
	"context"
	"fmt"
	"time"

	"linke/internal/logger"
	"linke/internal/model"

	"gorm.io/gorm"
)

type InviteCodeUsageService struct {
	db *gorm.DB
}

func NewInviteCodeUsageService(db *gorm.DB) *InviteCodeUsageService {
	return &InviteCodeUsageService{
		db: db,
	}
}

// CreateUsageRecord creates a new usage record for an invite code
func (s *InviteCodeUsageService) CreateUsageRecord(ctx context.Context, inviteCodeID uint, userID uint, ipAddress, userAgent string) (*model.InviteCodeUsage, error) {
	usage := &model.InviteCodeUsage{
		InviteCodeID: inviteCodeID,
		UsedByID:     userID,
		UsedAt:       time.Now(),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	if err := s.db.WithContext(ctx).Create(usage).Error; err != nil {
		logger.Error("Failed to create invite code usage record",
			logger.Uint("invite_code_id", inviteCodeID),
			logger.Uint("user_id", userID),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to create usage record: %w", err)
	}

	logger.Info("Invite code usage record created",
		logger.Uint("usage_id", usage.ID),
		logger.Uint("invite_code_id", inviteCodeID),
		logger.Uint("user_id", userID),
	)

	return usage, nil
}

// GetUsagesByInviteCode retrieves all usage records for a specific invite code
func (s *InviteCodeUsageService) GetUsagesByInviteCode(ctx context.Context, inviteCodeID uint, limit, offset int) ([]*model.InviteCodeUsage, int64, error) {
	var usages []*model.InviteCodeUsage
	var total int64

	// Count total usages
	if err := s.db.WithContext(ctx).Model(&model.InviteCodeUsage{}).Where("invite_code_id = ?", inviteCodeID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invite code usages: %w", err)
	}

	// Get usages with pagination
	if err := s.db.WithContext(ctx).
		Where("invite_code_id = ?", inviteCodeID).
		Order("used_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&usages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list invite code usages: %w", err)
	}

	return usages, total, nil
}

// GetUsagesByUser retrieves all usage records for a specific user
func (s *InviteCodeUsageService) GetUsagesByUser(ctx context.Context, userID uint, limit, offset int) ([]*model.InviteCodeUsage, int64, error) {
	var usages []*model.InviteCodeUsage
	var total int64

	// Count total usages
	if err := s.db.WithContext(ctx).Model(&model.InviteCodeUsage{}).Where("used_by_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user invite code usages: %w", err)
	}

	// Get usages with pagination
	if err := s.db.WithContext(ctx).
		Where("used_by_id = ?", userID).
		Order("used_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&usages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list user invite code usages: %w", err)
	}

	return usages, total, nil
}

// GetUsagesByCreator retrieves all usage records for invite codes created by a specific user
func (s *InviteCodeUsageService) GetUsagesByCreator(ctx context.Context, creatorID uint, limit, offset int) ([]*model.InviteCodeUsage, int64, error) {
	var usages []*model.InviteCodeUsage
	var total int64

	// Count total usages
	if err := s.db.WithContext(ctx).Model(&model.InviteCodeUsage{}).
		Joins("JOIN invite_codes ON invite_codes.id = invite_code_usages.invite_code_id").
		Where("invite_codes.created_by_id = ?", creatorID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count creator invite code usages: %w", err)
	}

	// Get usages with pagination
	if err := s.db.WithContext(ctx).
		Joins("JOIN invite_codes ON invite_codes.id = invite_code_usages.invite_code_id").
		Where("invite_codes.created_by_id = ?", creatorID).
		Order("invite_code_usages.used_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&usages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list creator invite code usages: %w", err)
	}

	return usages, total, nil
}

// LoadRelatedData loads related user and invite code data for usage records
func (s *InviteCodeUsageService) LoadRelatedData(ctx context.Context, usages []*model.InviteCodeUsage) error {
	if len(usages) == 0 {
		return nil
	}

	// Collect unique IDs
	var userIDs []uint
	var inviteCodeIDs []uint
	
	for _, usage := range usages {
		userIDs = append(userIDs, usage.UsedByID)
		inviteCodeIDs = append(inviteCodeIDs, usage.InviteCodeID)
	}

	// Load users
	var users []*model.User
	if err := s.db.WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return fmt.Errorf("failed to load users: %w", err)
	}

	// Load invite codes
	var inviteCodes []*model.InviteCode
	if err := s.db.WithContext(ctx).Where("id IN ?", inviteCodeIDs).Find(&inviteCodes).Error; err != nil {
		return fmt.Errorf("failed to load invite codes: %w", err)
	}

	// Create maps for quick lookup
	userMap := make(map[uint]*model.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	inviteCodeMap := make(map[uint]*model.InviteCode)
	for _, code := range inviteCodes {
		inviteCodeMap[code.ID] = code
	}

	// Assign related data
	for _, usage := range usages {
		if user, exists := userMap[usage.UsedByID]; exists {
			usage.UsedBy = user
		}
		if inviteCode, exists := inviteCodeMap[usage.InviteCodeID]; exists {
			usage.InviteCode = inviteCode
		}
	}

	return nil
}

// GetUsageStats returns statistics about invite code usage
func (s *InviteCodeUsageService) GetUsageStats(ctx context.Context) (map[string]interface{}, error) {
	var stats map[string]interface{} = make(map[string]interface{})

	// Total usages
	var totalUsages int64
	if err := s.db.WithContext(ctx).Model(&model.InviteCodeUsage{}).Count(&totalUsages).Error; err != nil {
		return nil, fmt.Errorf("failed to count total usages: %w", err)
	}
	stats["total_usages"] = totalUsages

	// Usages today
	today := time.Now().Truncate(24 * time.Hour)
	var todayUsages int64
	if err := s.db.WithContext(ctx).Model(&model.InviteCodeUsage{}).Where("used_at >= ?", today).Count(&todayUsages).Error; err != nil {
		return nil, fmt.Errorf("failed to count today's usages: %w", err)
	}
	stats["today_usages"] = todayUsages

	// Usages this week
	weekAgo := time.Now().AddDate(0, 0, -7)
	var weekUsages int64
	if err := s.db.WithContext(ctx).Model(&model.InviteCodeUsage{}).Where("used_at >= ?", weekAgo).Count(&weekUsages).Error; err != nil {
		return nil, fmt.Errorf("failed to count week's usages: %w", err)
	}
	stats["week_usages"] = weekUsages

	// Usages this month
	monthAgo := time.Now().AddDate(0, -1, 0)
	var monthUsages int64
	if err := s.db.WithContext(ctx).Model(&model.InviteCodeUsage{}).Where("used_at >= ?", monthAgo).Count(&monthUsages).Error; err != nil {
		return nil, fmt.Errorf("failed to count month's usages: %w", err)
	}
	stats["month_usages"] = monthUsages

	return stats, nil
}