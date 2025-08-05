package repositories

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/auth/entities"
	"linke/internal/domains/auth/usecases/interfaces"
	"linke/internal/shared/framework"
	"linke/internal/shared/repository"

	"gorm.io/gorm"
)

// jwtBlacklistRepository implements the JWTBlacklistRepository interface
type jwtBlacklistRepository struct {
	*repository.TimeBasedRepositoryImpl[entities.JWTBlacklist, string]
}

// NewJWTBlacklistRepository creates a new JWTBlacklistRepository implementation
func NewJWTBlacklistRepository(db *gorm.DB, logger framework.Logger) interfaces.JWTBlacklistRepository {
	return &jwtBlacklistRepository{
		TimeBasedRepositoryImpl: repository.NewTimeBasedRepository[entities.JWTBlacklist, string](db, logger),
	}
}

// GetByTokenHash retrieves a JWT blacklist entry by token hash
func (r *jwtBlacklistRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entities.JWTBlacklist, error) {
	var blacklist entities.JWTBlacklist
	if err := r.GetDB().WithContext(ctx).Where("token_hash = ?", tokenHash).First(&blacklist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("JWT blacklist entry not found")
		}
		return nil, fmt.Errorf("failed to get JWT blacklist entry: %w", err)
	}
	return &blacklist, nil
}

// Delete deletes a JWT blacklist entry by token hash
func (r *jwtBlacklistRepository) Delete(ctx context.Context, tokenHash string) error {
	result := r.GetDB().WithContext(ctx).Delete(&entities.JWTBlacklist{}, "token_hash = ?", tokenHash)
	if result.Error != nil {
		return fmt.Errorf("failed to delete JWT blacklist entry: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("JWT blacklist entry not found")
	}

	return nil
}

// IsTokenBlacklisted checks if a token is blacklisted
func (r *jwtBlacklistRepository) IsTokenBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.JWTBlacklist{}).
		Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check token blacklist status: %w", err)
	}
	return count > 0, nil
}

// BlacklistToken adds a token to the blacklist
func (r *jwtBlacklistRepository) BlacklistToken(ctx context.Context, tokenHash string, userID *uint, reason string, expiresAt time.Time) error {
	blacklist := &entities.JWTBlacklist{
		TokenHash: tokenHash,
		UserID:    userID,
		Reason:    reason,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	return r.Create(ctx, blacklist)
}

// BlacklistAllUserTokens adds all user tokens to the blacklist (tokens issued before a specific time)
func (r *jwtBlacklistRepository) BlacklistAllUserTokens(ctx context.Context, userID uint, reason string, beforeTime time.Time, expiresAt time.Time) error {
	// Create a special entry to mark all user tokens as blacklisted before a certain time
	blacklist := &entities.JWTBlacklist{
		TokenHash: fmt.Sprintf("user_%d_before_%d", userID, beforeTime.Unix()),
		UserID:    &userID,
		Reason:    reason,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	if err := r.Create(ctx, blacklist); err != nil {
		return fmt.Errorf("failed to blacklist all user tokens: %w", err)
	}

	return nil
}

// IsUserTokensBlacklisted checks if user tokens are blacklisted for tokens issued at a specific time
func (r *jwtBlacklistRepository) IsUserTokensBlacklisted(ctx context.Context, userID uint, tokenIssuedAt time.Time) (bool, error) {
	var count int64
	beforeHash := fmt.Sprintf("user_%d_before_%d", userID, tokenIssuedAt.Unix())

	if err := r.GetDB().WithContext(ctx).Model(&entities.JWTBlacklist{}).
		Where("user_id = ? AND expires_at > ? AND (token_hash = ? OR token_hash LIKE ?)",
			userID, time.Now(), beforeHash, fmt.Sprintf("user_%d_before_%%", userID)).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user token blacklist status: %w", err)
	}
	return count > 0, nil
}

// GetUserBlacklistedTokens retrieves all blacklisted tokens for a user
func (r *jwtBlacklistRepository) GetUserBlacklistedTokens(ctx context.Context, userID uint) ([]*entities.JWTBlacklist, error) {
	var blacklists []*entities.JWTBlacklist
	if err := r.GetDB().WithContext(ctx).Where("user_id = ?", userID).Find(&blacklists).Error; err != nil {
		return nil, fmt.Errorf("failed to get user blacklisted tokens: %w", err)
	}
	return blacklists, nil
}

// ListByUser lists JWT blacklist entries for a specific user with pagination
func (r *jwtBlacklistRepository) ListByUser(ctx context.Context, userID uint, limit, offset int) ([]*entities.JWTBlacklist, int64, error) {
	var blacklists []*entities.JWTBlacklist
	var total int64

	// Count total entries for user
	if err := r.GetDB().WithContext(ctx).Model(&entities.JWTBlacklist{}).
		Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count JWT blacklist entries by user: %w", err)
	}

	// Get entries with pagination
	if err := r.GetDB().WithContext(ctx).Where("user_id = ?", userID).
		Limit(limit).Offset(offset).Order("created_at DESC").Find(&blacklists).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list JWT blacklist entries by user: %w", err)
	}

	return blacklists, total, nil
}

// ListByReason lists JWT blacklist entries by reason with pagination
func (r *jwtBlacklistRepository) ListByReason(ctx context.Context, reason string, limit, offset int) ([]*entities.JWTBlacklist, int64, error) {
	var blacklists []*entities.JWTBlacklist
	var total int64

	// Count total entries for reason
	if err := r.GetDB().WithContext(ctx).Model(&entities.JWTBlacklist{}).
		Where("reason = ?", reason).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count JWT blacklist entries by reason: %w", err)
	}

	// Get entries with pagination
	if err := r.GetDB().WithContext(ctx).Where("reason = ?", reason).
		Limit(limit).Offset(offset).Order("created_at DESC").Find(&blacklists).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list JWT blacklist entries by reason: %w", err)
	}

	return blacklists, total, nil
}

// DeleteExpired removes all expired JWT blacklist entries
func (r *jwtBlacklistRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.GetDB().WithContext(ctx).Where("expires_at <= ?", time.Now()).Delete(&entities.JWTBlacklist{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete expired JWT blacklist entries: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// DeleteByUser removes all JWT blacklist entries for a specific user
func (r *jwtBlacklistRepository) DeleteByUser(ctx context.Context, userID uint) (int64, error) {
	result := r.GetDB().WithContext(ctx).Where("user_id = ?", userID).Delete(&entities.JWTBlacklist{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete JWT blacklist entries by user: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// DeleteOlderThan removes JWT blacklist entries older than a specific time
func (r *jwtBlacklistRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result := r.GetDB().WithContext(ctx).Where("created_at < ?", before).Delete(&entities.JWTBlacklist{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete old JWT blacklist entries: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// CountByReason returns the count of JWT blacklist entries by reason
func (r *jwtBlacklistRepository) CountByReason(ctx context.Context, reason string) (int64, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.JWTBlacklist{}).
		Where("reason = ?", reason).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count JWT blacklist entries by reason %s: %w", reason, err)
	}
	return count, nil
}

// CountByUser returns the count of JWT blacklist entries for a specific user
func (r *jwtBlacklistRepository) CountByUser(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.JWTBlacklist{}).
		Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count JWT blacklist entries for user %d: %w", userID, err)
	}
	return count, nil
}

// CountExpired returns the count of expired JWT blacklist entries
func (r *jwtBlacklistRepository) CountExpired(ctx context.Context) (int64, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.JWTBlacklist{}).
		Where("expires_at <= ?", time.Now()).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count expired JWT blacklist entries: %w", err)
	}
	return count, nil
}

// CountActive returns the count of active (non-expired) JWT blacklist entries
func (r *jwtBlacklistRepository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.JWTBlacklist{}).
		Where("expires_at > ?", time.Now()).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count active JWT blacklist entries: %w", err)
	}
	return count, nil
}

// GetExpiredBefore retrieves JWT blacklist entries that expired before a specific time
func (r *jwtBlacklistRepository) GetExpiredBefore(ctx context.Context, before time.Time, limit int) ([]*entities.JWTBlacklist, error) {
	var blacklists []*entities.JWTBlacklist
	if err := r.GetDB().WithContext(ctx).Where("expires_at <= ?", before).
		Limit(limit).Order("expires_at ASC").Find(&blacklists).Error; err != nil {
		return nil, fmt.Errorf("failed to get expired JWT blacklist entries: %w", err)
	}
	return blacklists, nil
}
