package implementations

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/auth/entities"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

// JWTBlacklistService handles JWT token blacklisting
type JWTBlacklistService struct {
	db *gorm.DB
}

// NewJWTBlacklistService creates a new JWT blacklist service
func NewJWTBlacklistService(db *gorm.DB) *JWTBlacklistService {
	return &JWTBlacklistService{
		db: db,
	}
}

// BlacklistToken adds a token to the blacklist
func (j *JWTBlacklistService) BlacklistToken(ctx context.Context, token string, userID *uint, reason string, expiresAt time.Time) error {
	tokenHash := entities.HashToken(token)

	blacklistEntry := &entities.JWTBlacklist{
		TokenHash: tokenHash,
		UserID:    userID,
		Reason:    reason,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	if err := j.db.WithContext(ctx).Create(blacklistEntry).Error; err != nil {
		logger.Error("Failed to blacklist token",
			logger.String("reason", reason),
			logger.String("token_hash", tokenHash[:8]+"..."),
			logger.Error2("error", err))
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	logger.Info("Token blacklisted successfully",
		logger.String("reason", reason),
		logger.String("token_hash", tokenHash[:8]+"..."),
		logger.Uint("user_id", getUintValue(userID)))

	return nil
}

// IsTokenBlacklisted checks if a token is blacklisted
func (j *JWTBlacklistService) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	tokenHash := entities.HashToken(token)

	var blacklistEntry entities.JWTBlacklist
	err := j.db.WithContext(ctx).Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).First(&blacklistEntry).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil // Token not blacklisted
		}
		logger.Error("Failed to check token blacklist",
			logger.String("token_hash", tokenHash[:8]+"..."),
			logger.Error2("error", err))
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	return true, nil
}

// BlacklistAllUserTokens blacklists all tokens for a specific user
func (j *JWTBlacklistService) BlacklistAllUserTokens(ctx context.Context, userID uint, reason string, tokenExpiresAt time.Time) error {
	// This creates a "user blacklist" entry that will invalidate all tokens for the user
	// We use a special token hash pattern to identify user-wide blacklists
	userTokenHash := fmt.Sprintf("user_%d_%d", userID, time.Now().Unix())

	blacklistEntry := &entities.JWTBlacklist{
		TokenHash: userTokenHash,
		UserID:    &userID,
		Reason:    reason,
		ExpiresAt: tokenExpiresAt,
		CreatedAt: time.Now(),
	}

	if err := j.db.WithContext(ctx).Create(blacklistEntry).Error; err != nil {
		logger.Error("Failed to blacklist all user tokens",
			logger.Uint("user_id", userID),
			logger.String("reason", reason),
			logger.Error2("error", err))
		return fmt.Errorf("failed to blacklist all user tokens: %w", err)
	}

	logger.Info("All user tokens blacklisted",
		logger.Uint("user_id", userID),
		logger.String("reason", reason))

	return nil
}

// IsUserTokensBlacklisted checks if all tokens for a user are blacklisted
func (j *JWTBlacklistService) IsUserTokensBlacklisted(ctx context.Context, userID uint, tokenIssuedAt time.Time) (bool, error) {
	var count int64
	err := j.db.WithContext(ctx).Model(&entities.JWTBlacklist{}).
		Where("user_id = ? AND token_hash LIKE ? AND created_at > ? AND expires_at > ?",
			userID, "user_%", tokenIssuedAt, time.Now()).
		Count(&count).Error

	if err != nil {
		logger.Error("Failed to check user token blacklist",
			logger.Uint("user_id", userID),
			logger.Error2("error", err))
		return false, fmt.Errorf("failed to check user token blacklist: %w", err)
	}

	return count > 0, nil
}

// CleanupExpiredEntries removes expired blacklist entries
func (j *JWTBlacklistService) CleanupExpiredEntries(ctx context.Context) error {
	result := j.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&entities.JWTBlacklist{})

	if result.Error != nil {
		logger.Error("Failed to cleanup expired blacklist entries",
			logger.Error2("error", result.Error))
		return fmt.Errorf("failed to cleanup expired blacklist entries: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		logger.Info("Cleaned up expired blacklist entries",
			logger.Int64("count", result.RowsAffected))
	}

	return nil
}

// GetBlacklistStats returns statistics about blacklisted tokens
func (j *JWTBlacklistService) GetBlacklistStats(ctx context.Context) (map[string]any, error) {
	var totalCount, expiredCount int64

	// Count total blacklisted tokens
	if err := j.db.WithContext(ctx).Model(&entities.JWTBlacklist{}).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count total blacklisted tokens: %w", err)
	}

	// Count expired tokens
	if err := j.db.WithContext(ctx).Model(&entities.JWTBlacklist{}).
		Where("expires_at < ?", time.Now()).Count(&expiredCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count expired blacklisted tokens: %w", err)
	}

	// Count by reason
	var reasonStats []struct {
		Reason string `json:"reason"`
		Count  int64  `json:"count"`
	}

	if err := j.db.WithContext(ctx).Model(&entities.JWTBlacklist{}).
		Select("reason, COUNT(*) as count").
		Where("expires_at > ?", time.Now()).
		Group("reason").
		Scan(&reasonStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get reason statistics: %w", err)
	}

	return map[string]any{
		"total_blacklisted":  totalCount,
		"active_blacklisted": totalCount - expiredCount,
		"expired_entries":    expiredCount,
		"reason_breakdown":   reasonStats,
	}, nil
}

// Helper function to safely get uint value from pointer
func getUintValue(ptr *uint) uint {
	if ptr == nil {
		return 0
	}
	return *ptr
}
