package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"linke/internal/logger"
	"linke/internal/model"

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

// CreateInviteCodeRequest represents the request to create an invite code
type CreateInviteCodeRequest struct {
	MaxUses     int    `json:"max_uses" binding:"min=1,max=100" example:"10"`                       // Maximum number of times the code can be used
	Description string `json:"description" binding:"max=255" example:"Friend invitation code"`     // Description of the invite code
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
	var existingCode model.InviteCode
	if err := s.db.Where("code = ?", code).First(&existingCode).Error; err == nil {
		// Code exists, try again (recursive call)
		return s.GenerateInviteCode()
	}
	
	return code, nil
}

// CreateInviteCode creates a new invite code
func (s *InviteCodeService) CreateInviteCode(ctx context.Context, createdByID uint, req *CreateInviteCodeRequest) (*model.InviteCode, error) {
	// Generate unique code
	code, err := s.GenerateInviteCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invite code: %w", err)
	}

	// Create invite code
	inviteCode := &model.InviteCode{
		Code:        code,
		CreatedByID: createdByID,
		Status:      model.InviteCodeStatusActive,
		MaxUses:     req.MaxUses,
		UsedCount:   0,
		Description: req.Description,
	}

	if err := s.db.WithContext(ctx).Create(inviteCode).Error; err != nil {
		logger.Error("Failed to create invite code",
			logger.Uint("created_by_id", createdByID),
			logger.Error2("error", err),
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
func (s *InviteCodeService) GetInviteCodeByCode(ctx context.Context, code string) (*model.InviteCode, error) {
	var inviteCode model.InviteCode
	if err := s.db.WithContext(ctx).Where("code = ?", code).First(&inviteCode).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invite code not found")
		}
		return nil, fmt.Errorf("failed to get invite code: %w", err)
	}
	return &inviteCode, nil
}

// GetInviteCodeByID retrieves an invite code by its ID
func (s *InviteCodeService) GetInviteCodeByID(ctx context.Context, id uint) (*model.InviteCode, error) {
	var inviteCode model.InviteCode
	if err := s.db.WithContext(ctx).First(&inviteCode, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invite code not found")
		}
		return nil, fmt.Errorf("failed to get invite code: %w", err)
	}
	return &inviteCode, nil
}

// GetInviteCodeByIDWithRelations retrieves an invite code by its ID with related data
func (s *InviteCodeService) GetInviteCodeByIDWithRelations(ctx context.Context, id uint) (*model.InviteCode, error) {
	var inviteCode model.InviteCode
	if err := s.db.WithContext(ctx).First(&inviteCode, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invite code not found")
		}
		return nil, fmt.Errorf("failed to get invite code: %w", err)
	}

	// Load creator
	var creator model.User
	if err := s.db.WithContext(ctx).First(&creator, inviteCode.CreatedByID).Error; err == nil {
		inviteCode.CreatedBy = &creator
	}

	// Load usage records
	var usageRecords []*model.InviteCodeUsage
	if err := s.db.WithContext(ctx).Where("invite_code_id = ?", id).Order("used_at DESC").Find(&usageRecords).Error; err == nil {
		// Load users for each usage record
		var userIDs []uint
		for _, usage := range usageRecords {
			userIDs = append(userIDs, usage.UsedByID)
		}
		
		if len(userIDs) > 0 {
			var users []*model.User
			if err := s.db.WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error; err == nil {
				userMap := make(map[uint]*model.User)
				for _, user := range users {
					userMap[user.ID] = user
				}
				
				for _, usage := range usageRecords {
					if user, exists := userMap[usage.UsedByID]; exists {
						usage.UsedBy = user
					}
				}
			}
		}
		
		inviteCode.UsageRecords = usageRecords
	}

	return &inviteCode, nil
}

// ValidateInviteCode validates if an invite code can be used
func (s *InviteCodeService) ValidateInviteCode(ctx context.Context, code string) (*model.InviteCode, error) {
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
func (s *InviteCodeService) UseInviteCode(ctx context.Context, code string, userID uint, ipAddress, userAgent string) (*model.InviteCode, error) {
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
		inviteCode.Status = model.InviteCodeStatusUsed
	}

	// Update invite code
	if err := tx.Save(inviteCode).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to update invite code usage",
			logger.Uint("invite_code_id", inviteCode.ID),
			logger.Uint("user_id", userID),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to update invite code: %w", err)
	}

	// Create usage record
	usage := &model.InviteCodeUsage{
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
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to create usage record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit invite code usage transaction",
			logger.Uint("invite_code_id", inviteCode.ID),
			logger.Uint("user_id", userID),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Info("Invite code used successfully",
		logger.Uint("invite_code_id", inviteCode.ID),
		logger.String("code", code),
		logger.Uint("user_id", userID),
		logger.Int("used_count", inviteCode.UsedCount),
	)

	return inviteCode, nil
}


// ListAllInviteCodes lists all invite codes
func (s *InviteCodeService) ListAllInviteCodes(ctx context.Context, limit, offset int) ([]*model.InviteCode, int64, error) {
	var codes []*model.InviteCode
	var total int64

	// Count total codes
	if err := s.db.WithContext(ctx).Model(&model.InviteCode{}).Count(&total).Error; err != nil {
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
func (s *InviteCodeService) ListInviteCodesByCreator(ctx context.Context, creatorID uint, limit, offset int) ([]*model.InviteCode, int64, error) {
	var codes []*model.InviteCode
	var total int64

	// Count total codes
	if err := s.db.WithContext(ctx).Model(&model.InviteCode{}).Where("created_by_id = ?", creatorID).Count(&total).Error; err != nil {
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
func (s *InviteCodeService) UpdateInviteCodeStatus(ctx context.Context, id uint, status string) (*model.InviteCode, error) {
	var inviteCode model.InviteCode
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
			logger.Error2("error", err),
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
	result := s.db.WithContext(ctx).Delete(&model.InviteCode{}, id)
	if result.Error != nil {
		logger.Error("Failed to delete invite code",
			logger.Uint("invite_code_id", id),
			logger.Error2("error", result.Error),
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
func (s *InviteCodeService) GetInviteCodeStats(ctx context.Context) (map[string]interface{}, error) {
	var stats map[string]interface{} = make(map[string]interface{})

	// Total invite codes
	var totalCodes int64
	if err := s.db.WithContext(ctx).Model(&model.InviteCode{}).Count(&totalCodes).Error; err != nil {
		return nil, fmt.Errorf("failed to count total invite codes: %w", err)
	}
	stats["total_codes"] = totalCodes

	// Active invite codes
	var activeCodes int64
	if err := s.db.WithContext(ctx).Model(&model.InviteCode{}).Where("status = ?", model.InviteCodeStatusActive).Count(&activeCodes).Error; err != nil {
		return nil, fmt.Errorf("failed to count active invite codes: %w", err)
	}
	stats["active_codes"] = activeCodes

	// Used invite codes
	var usedCodes int64
	if err := s.db.WithContext(ctx).Model(&model.InviteCode{}).Where("status = ?", model.InviteCodeStatusUsed).Count(&usedCodes).Error; err != nil {
		return nil, fmt.Errorf("failed to count used invite codes: %w", err)
	}
	stats["used_codes"] = usedCodes

	// Disabled invite codes
	var disabledCodes int64
	if err := s.db.WithContext(ctx).Model(&model.InviteCode{}).Where("status = ?", model.InviteCodeStatusDisabled).Count(&disabledCodes).Error; err != nil {
		return nil, fmt.Errorf("failed to count disabled invite codes: %w", err)
	}
	stats["disabled_codes"] = disabledCodes

	// Total usage count
	var totalUsage int64
	if err := s.db.WithContext(ctx).Model(&model.InviteCode{}).Select("COALESCE(SUM(used_count), 0)").Scan(&totalUsage).Error; err != nil {
		return nil, fmt.Errorf("failed to count total usage: %w", err)
	}
	stats["total_usage"] = totalUsage

	return stats, nil
}