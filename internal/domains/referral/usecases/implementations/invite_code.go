package implementations

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"linke/internal/domains/referral/constants"
	"linke/internal/domains/referral/dto"
	"linke/internal/domains/referral/entities"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

type InviteCodeService struct {
	db *gorm.DB
}

func NewInviteCodeService(db *gorm.DB) *InviteCodeService {
	return &InviteCodeService{
		db: db,
	}
}


// GenerateInviteCode generates a random invite code
func (s *InviteCodeService) GenerateInviteCode() (string, error) {
	// Generate 16 bytes of random data
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to hex string (32 characters)
	code := hex.EncodeToString(bytes)

	// Check if code already exists (very unlikely but possible)
	var existingCode entities.InviteCode
	if err := s.db.Where("code = ?", code).First(&existingCode).Error; err == nil {
		// Code exists, try again (recursive call)
		return s.GenerateInviteCode()
	}

	return code, nil
}

// CreateInviteCode creates a new invite code
func (s *InviteCodeService) CreateInviteCode(ctx context.Context, createdByID uint, req *dto.CreateInviteCodeRequest) (*entities.InviteCode, error) {
	// Generate unique code
	code, err := s.GenerateInviteCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invite code: %w", err)
	}

	// Create invite code
	inviteCode := &entities.InviteCode{
		Code:                   code,
		CreatedByID:            createdByID,
		Status:                 constants.InviteCodeStatusActive,
		MaxUses:                req.MaxUses,
		UsedCount:              0,
		Description:            req.Description,
		ReferralCampaignID:     req.ReferralCampaignID,
		ReferralRewardAmount:   req.ReferralRewardAmount,
		ReferralRewardCurrency: "USD",
	}

	if err := s.db.WithContext(ctx).Create(inviteCode).Error; err != nil {
		logger.Error("Failed to create invite code",
			logger.Uint("created_by_id", createdByID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to create invite code: %w", err)
	}

	logger.Info("Invite code created successfully",
		logger.Uint("invite_code_id", inviteCode.ID),
		logger.String("code", code),
		logger.Uint("created_by_id", createdByID),
	)

	return inviteCode, nil
}

// GetInviteCodeByCode retrieves an invite code by its code
func (s *InviteCodeService) GetInviteCodeByCode(ctx context.Context, code string) (*entities.InviteCode, error) {
	var inviteCode entities.InviteCode
	if err := s.db.WithContext(ctx).Where("code = ?", code).First(&inviteCode).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invite code not found")
		}
		return nil, fmt.Errorf("failed to get invite code: %w", err)
	}
	return &inviteCode, nil
}

// GetInviteCodeByID retrieves an invite code by its ID
func (s *InviteCodeService) GetInviteCodeByID(ctx context.Context, id uint) (*entities.InviteCode, error) {
	var inviteCode entities.InviteCode
	if err := s.db.WithContext(ctx).First(&inviteCode, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invite code not found")
		}
		return nil, fmt.Errorf("failed to get invite code: %w", err)
	}
	return &inviteCode, nil
}

// GetInviteCodeByIDWithRelations retrieves an invite code by its ID with related data
func (s *InviteCodeService) GetInviteCodeByIDWithRelations(ctx context.Context, id uint) (*entities.InviteCode, error) {
	var inviteCode entities.InviteCode
	if err := s.db.WithContext(ctx).First(&inviteCode, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invite code not found")
		}
		return nil, fmt.Errorf("failed to get invite code: %w", err)
	}

	// TODO: Load creator information through user service interface
	// For now, we'll just return the invite code with CreatedByID

	// TODO: Load usage records through appropriate service interfaces
	// For now, we'll just return the invite code without related data

	return &inviteCode, nil
}

// ValidateInviteCode validates if an invite code can be used
func (s *InviteCodeService) ValidateInviteCode(ctx context.Context, code string) (*entities.InviteCode, error) {
	inviteCode, err := s.GetInviteCodeByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Check if code can be used
	if !inviteCode.CanBeUsed() {
		if inviteCode.IsExhausted() {
			return nil, fmt.Errorf("invite code has reached maximum uses")
		}
		return nil, fmt.Errorf("invite code is not active")
	}

	return inviteCode, nil
}

// UseInviteCode marks an invite code as used by a user and creates usage record
func (s *InviteCodeService) UseInviteCode(ctx context.Context, code string, userID uint, ipAddress, userAgent string) (*entities.InviteCodeUsage, error) {
	// Start a transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get and validate invite code
	inviteCode, err := s.ValidateInviteCode(ctx, code)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Increment used count
	inviteCode.UsedCount++

	// Update status if exhausted
	if inviteCode.UsedCount >= inviteCode.MaxUses {
		inviteCode.Status = constants.InviteCodeStatusUsed
	}

	// Update invite code
	if err := tx.Save(inviteCode).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to update invite code usage",
			logger.Uint("invite_code_id", inviteCode.ID),
			logger.Uint("user_id", userID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to update invite code: %w", err)
	}

	// Create usage record
	usage := &entities.InviteCodeUsage{
		InviteCodeID: inviteCode.ID,
		UsedByID:     userID,
		UsedAt:       time.Now(),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	if err := tx.Create(usage).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to create invite code usage record",
			logger.Uint("invite_code_id", inviteCode.ID),
			logger.Uint("user_id", userID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to create usage record: %w", err)
	}

	// TODO: Create referral record through referral service interface
	// This should be handled at the application layer to avoid circular dependencies

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit invite code usage transaction",
			logger.Uint("invite_code_id", inviteCode.ID),
			logger.Uint("user_id", userID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Info("Invite code used successfully",
		logger.Uint("invite_code_id", inviteCode.ID),
		logger.String("code", code),
		logger.Uint("user_id", userID),
		logger.Int("used_count", inviteCode.UsedCount),
	)

	return usage, nil
}

// ListAllInviteCodes lists all invite codes
func (s *InviteCodeService) ListAllInviteCodes(ctx context.Context, limit, offset int) ([]*entities.InviteCode, int64, error) {
	var codes []*entities.InviteCode
	var total int64

	// Count total codes
	if err := s.db.WithContext(ctx).Model(&entities.InviteCode{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invite codes: %w", err)
	}

	// Get codes with pagination
	if err := s.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&codes).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list invite codes: %w", err)
	}

	return codes, total, nil
}

// ListInviteCodesByCreator lists invite codes created by a specific user
func (s *InviteCodeService) ListInviteCodesByCreator(ctx context.Context, creatorID uint, limit, offset int) ([]*entities.InviteCode, int64, error) {
	var codes []*entities.InviteCode
	var total int64

	// Count total codes
	if err := s.db.WithContext(ctx).Model(&entities.InviteCode{}).Where("created_by_id = ?", creatorID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invite codes: %w", err)
	}

	// Get codes with pagination
	if err := s.db.WithContext(ctx).
		Where("created_by_id = ?", creatorID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&codes).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list invite codes: %w", err)
	}

	return codes, total, nil
}

// UpdateInviteCodeStatus updates the status of an invite code
func (s *InviteCodeService) UpdateInviteCodeStatus(ctx context.Context, id uint, status string) (*entities.InviteCode, error) {
	var inviteCode entities.InviteCode
	if err := s.db.WithContext(ctx).First(&inviteCode, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invite code not found")
		}
		return nil, fmt.Errorf("failed to get invite code: %w", err)
	}

	// Update status
	inviteCode.Status = status
	if err := s.db.WithContext(ctx).Save(&inviteCode).Error; err != nil {
		logger.Error("Failed to update invite code status",
			logger.Uint("invite_code_id", id),
			logger.String("status", status),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to update invite code status: %w", err)
	}

	logger.Info("Invite code status updated",
		logger.Uint("invite_code_id", id),
		logger.String("status", status),
	)

	return &inviteCode, nil
}

// DeleteInviteCode soft deletes an invite code
func (s *InviteCodeService) DeleteInviteCode(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&entities.InviteCode{}, id)
	if result.Error != nil {
		logger.Error("Failed to delete invite code",
			logger.Uint("invite_code_id", id),
			logger.ErrorField(result.Error),
		)
		return fmt.Errorf("failed to delete invite code: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("invite code not found")
	}

	logger.Info("Invite code deleted successfully",
		logger.Uint("invite_code_id", id),
	)

	return nil
}

// GetInviteCodeStats returns statistics about invite codes
func (s *InviteCodeService) GetInviteCodeStats(ctx context.Context) (map[string]any, error) {
	var stats map[string]any = make(map[string]any)

	// Total invite codes
	var totalCodes int64
	if err := s.db.WithContext(ctx).Model(&entities.InviteCode{}).Count(&totalCodes).Error; err != nil {
		return nil, fmt.Errorf("failed to count total invite codes: %w", err)
	}
	stats["total_codes"] = totalCodes

	// Active invite codes
	var activeCodes int64
	if err := s.db.WithContext(ctx).Model(&entities.InviteCode{}).Where("status = ?", constants.InviteCodeStatusActive).Count(&activeCodes).Error; err != nil {
		return nil, fmt.Errorf("failed to count active invite codes: %w", err)
	}
	stats["active_codes"] = activeCodes

	// Used invite codes
	var usedCodes int64
	if err := s.db.WithContext(ctx).Model(&entities.InviteCode{}).Where("status = ?", constants.InviteCodeStatusUsed).Count(&usedCodes).Error; err != nil {
		return nil, fmt.Errorf("failed to count used invite codes: %w", err)
	}
	stats["used_codes"] = usedCodes

	// Disabled invite codes
	var disabledCodes int64
	if err := s.db.WithContext(ctx).Model(&entities.InviteCode{}).Where("status = ?", constants.InviteCodeStatusDisabled).Count(&disabledCodes).Error; err != nil {
		return nil, fmt.Errorf("failed to count disabled invite codes: %w", err)
	}
	stats["disabled_codes"] = disabledCodes

	// Total usage count
	var totalUsage int64
	if err := s.db.WithContext(ctx).Model(&entities.InviteCode{}).Select("COALESCE(SUM(used_count), 0)").Scan(&totalUsage).Error; err != nil {
		return nil, fmt.Errorf("failed to count total usage: %w", err)
	}
	stats["total_usage"] = totalUsage

	return stats, nil
}

// GetInviteCode gets an invite code by ID
func (s *InviteCodeService) GetInviteCode(ctx context.Context, inviteCodeID uint) (*entities.InviteCode, error) {
	return s.GetInviteCodeByID(ctx, inviteCodeID)
}

// UpdateInviteCode updates an invite code
func (s *InviteCodeService) UpdateInviteCode(ctx context.Context, inviteCodeID uint, req *dto.UpdateInviteCodeRequest) (*entities.InviteCode, error) {
	// Get existing invite code
	inviteCode, err := s.GetInviteCode(ctx, inviteCodeID)
	if err != nil {
		return nil, err
	}

	// Prepare updates
	updates := make(map[string]any)

	if req.MaxUses != nil {
		updates["max_uses"] = *req.MaxUses
	}

	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if req.ReferralRewardAmount != nil {
		updates["referral_reward_amount"] = *req.ReferralRewardAmount
	}

	updates["updated_at"] = time.Now()

	// Update the invite code
	if err := s.db.WithContext(ctx).Model(inviteCode).Updates(updates).Error; err != nil {
		logger.Error("Failed to update invite code", logger.Uint("inviteCodeID", uint(inviteCodeID)))
		return nil, fmt.Errorf("failed to update invite code: %w", err)
	}

	// Reload the invite code
	updatedInviteCode, err := s.GetInviteCode(ctx, inviteCodeID)
	if err != nil {
		return nil, err
	}

	logger.Info("Invite code updated successfully", logger.Uint("invite_code_id", inviteCodeID))

	return updatedInviteCode, nil
}

// GetInviteCodes gets invite codes with filtering and pagination
func (s *InviteCodeService) GetInviteCodes(ctx context.Context, req *dto.GetInviteCodesRequest) ([]*entities.InviteCode, int64, error) {
	query := s.db.WithContext(ctx).Model(&entities.InviteCode{})

	// Apply filters
	if req.CreatedByID != 0 {
		query = query.Where("created_by_id = ?", req.CreatedByID)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.ReferralCampaignID != nil {
		query = query.Where("referral_campaign_id = ?", *req.ReferralCampaignID)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		logger.Error("Failed to count invite codes", logger.ErrorField(err))
		return nil, 0, fmt.Errorf("failed to count invite codes: %w", err)
	}

	// Apply pagination and ordering
	query = query.Order("created_at DESC")

	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	var inviteCodes []*entities.InviteCode
	if err := query.Find(&inviteCodes).Error; err != nil {
		logger.Error("Failed to get invite codes", logger.ErrorField(err))
		return nil, 0, fmt.Errorf("failed to get invite codes: %w", err)
	}

	return inviteCodes, totalCount, nil
}

// GetUserInviteCodes gets invite codes created by a specific user
func (s *InviteCodeService) GetUserInviteCodes(ctx context.Context, userID uint, limit, offset int) ([]*entities.InviteCode, int64, error) {
	return s.ListInviteCodesByCreator(ctx, userID, limit, offset)
}

// ActivateInviteCode activates an invite code
func (s *InviteCodeService) ActivateInviteCode(ctx context.Context, inviteCodeID uint) error {
	_, err := s.UpdateInviteCodeStatus(ctx, inviteCodeID, constants.InviteCodeStatusActive)
	return err
}

// DeactivateInviteCode deactivates an invite code
func (s *InviteCodeService) DeactivateInviteCode(ctx context.Context, inviteCodeID uint) error {
	_, err := s.UpdateInviteCodeStatus(ctx, inviteCodeID, constants.InviteCodeStatusDisabled)
	return err
}

// ExpireInviteCode expires an invite code
func (s *InviteCodeService) ExpireInviteCode(ctx context.Context, inviteCodeID uint) error {
	_, err := s.UpdateInviteCodeStatus(ctx, inviteCodeID, constants.InviteCodeStatusDisabled)
	return err
}

// GetInviteCodeUsage gets usage records for an invite code
func (s *InviteCodeService) GetInviteCodeUsage(ctx context.Context, inviteCodeID uint, limit, offset int) ([]*entities.InviteCodeUsage, int64, error) {
	var usages []*entities.InviteCodeUsage
	var total int64

	query := s.db.WithContext(ctx).Model(&entities.InviteCodeUsage{}).Where("invite_code_id = ?", inviteCodeID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invite code usages: %w", err)
	}

	// Apply pagination and ordering
	query = query.Order("used_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&usages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get invite code usages: %w", err)
	}

	return usages, total, nil
}

// GetUserInviteCodeUsage gets invite code usage records for a specific user
func (s *InviteCodeService) GetUserInviteCodeUsage(ctx context.Context, userID uint, limit, offset int) ([]*entities.InviteCodeUsage, int64, error) {
	var usages []*entities.InviteCodeUsage
	var total int64

	query := s.db.WithContext(ctx).Model(&entities.InviteCodeUsage{}).Where("used_by_id = ?", userID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user invite code usages: %w", err)
	}

	// Apply pagination and ordering
	query = query.Order("used_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&usages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get user invite code usages: %w", err)
	}

	return usages, total, nil
}

// GetInviteCodeStatistics gets statistics for a specific invite code
func (s *InviteCodeService) GetInviteCodeStatistics(ctx context.Context, inviteCodeID uint) (map[string]any, error) {
	// Get invite code
	inviteCode, err := s.GetInviteCode(ctx, inviteCodeID)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]any)
	stats["invite_code_id"] = inviteCode.ID
	stats["code"] = inviteCode.Code
	stats["max_uses"] = inviteCode.MaxUses
	stats["used_count"] = inviteCode.UsedCount
	stats["remaining_uses"] = inviteCode.MaxUses - inviteCode.UsedCount
	stats["status"] = inviteCode.Status
	stats["created_at"] = inviteCode.CreatedAt

	// Get usage statistics
	var totalUsages int64
	if err := s.db.WithContext(ctx).Model(&entities.InviteCodeUsage{}).Where("invite_code_id = ?", inviteCodeID).Count(&totalUsages).Error; err != nil {
		return nil, fmt.Errorf("failed to count invite code usages: %w", err)
	}
	stats["total_usages"] = totalUsages

	// Get first and last usage dates
	var firstUsage, lastUsage entities.InviteCodeUsage
	if err := s.db.WithContext(ctx).Where("invite_code_id = ?", inviteCodeID).Order("used_at ASC").First(&firstUsage).Error; err == nil {
		stats["first_used_at"] = firstUsage.UsedAt
	}

	if err := s.db.WithContext(ctx).Where("invite_code_id = ?", inviteCodeID).Order("used_at DESC").First(&lastUsage).Error; err == nil {
		stats["last_used_at"] = lastUsage.UsedAt
	}

	return stats, nil
}

// GetUserInviteCodeStatistics gets invite code statistics for a specific user
func (s *InviteCodeService) GetUserInviteCodeStatistics(ctx context.Context, userID uint) (map[string]any, error) {
	stats := make(map[string]any)

	// Count total invite codes created by user
	var totalCodes int64
	if err := s.db.WithContext(ctx).Model(&entities.InviteCode{}).Where("created_by_id = ?", userID).Count(&totalCodes).Error; err != nil {
		return nil, fmt.Errorf("failed to count user invite codes: %w", err)
	}
	stats["total_codes"] = totalCodes

	// Count active codes
	var activeCodes int64
	if err := s.db.WithContext(ctx).Model(&entities.InviteCode{}).Where("created_by_id = ? AND status = ?", userID, constants.InviteCodeStatusActive).Count(&activeCodes).Error; err != nil {
		return nil, fmt.Errorf("failed to count active user invite codes: %w", err)
	}
	stats["active_codes"] = activeCodes

	// Sum total usages of user's codes
	var totalUsages int64
	if err := s.db.WithContext(ctx).Model(&entities.InviteCode{}).Where("created_by_id = ?", userID).Select("COALESCE(SUM(used_count), 0)").Scan(&totalUsages).Error; err != nil {
		return nil, fmt.Errorf("failed to count total user invite code usages: %w", err)
	}
	stats["total_usages"] = totalUsages

	// Count times user has used invite codes
	var userUsages int64
	if err := s.db.WithContext(ctx).Model(&entities.InviteCodeUsage{}).Where("used_by_id = ?", userID).Count(&userUsages).Error; err != nil {
		return nil, fmt.Errorf("failed to count user invite code usages: %w", err)
	}
	stats["used_codes"] = userUsages

	return stats, nil
}
