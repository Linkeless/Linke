package implementations

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/auth/entities"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

// LoginSecurityService handles login security features like failure tracking and account lockouts
type LoginSecurityService struct {
	db                 *gorm.DB
	maxFailures        int           // Maximum failed attempts before lockout
	lockoutDuration    time.Duration // How long to lock account
	failureWindow      time.Duration // Time window to count failures
	progressiveLockout bool          // Whether to use progressive lockout (longer locks for repeat offenders)
}

// LoginSecurityConfig contains configuration for login security
type LoginSecurityConfig struct {
	MaxFailures        int           `json:"max_failures"`        // Default: 5
	LockoutDuration    time.Duration `json:"lockout_duration"`    // Default: 15 minutes
	FailureWindow      time.Duration `json:"failure_window"`      // Default: 1 hour
	ProgressiveLockout bool          `json:"progressive_lockout"` // Default: true
}

// NewLoginSecurityService creates a new login security service
func NewLoginSecurityService(db *gorm.DB, config *LoginSecurityConfig) *LoginSecurityService {
	// Set defaults if config is nil
	if config == nil {
		config = &LoginSecurityConfig{
			MaxFailures:        5,
			LockoutDuration:    15 * time.Minute,
			FailureWindow:      1 * time.Hour,
			ProgressiveLockout: true,
		}
	}

	return &LoginSecurityService{
		db:                 db,
		maxFailures:        config.MaxFailures,
		lockoutDuration:    config.LockoutDuration,
		failureWindow:      config.FailureWindow,
		progressiveLockout: config.ProgressiveLockout,
	}
}

// RecordLoginAttempt records a login attempt (success or failure)
func (l *LoginSecurityService) RecordLoginAttempt(ctx context.Context, email, ip, userAgent, reason string, success bool, userID *uint) error {
	attempt := &entities.LoginAttempt{
		Email:     email,
		IP:        ip,
		UserAgent: userAgent,
		Success:   success,
		Reason:    reason,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	if err := l.db.WithContext(ctx).Create(attempt).Error; err != nil {
		logger.Error("Failed to record login attempt",
			logger.String("email", email),
			logger.String("ip", ip),
			logger.String("success", fmt.Sprintf("%t", success)),
			logger.ErrorField(err))
		return fmt.Errorf("failed to record login attempt: %w", err)
	}

	// If this was a failed attempt, update lockout tracking
	if !success {
		if err := l.updateFailureTracking(ctx, email, userID); err != nil {
			logger.Error("Failed to update failure tracking",
				logger.String("email", email),
				logger.ErrorField(err))
			// Don't return error here as the login attempt was recorded
		}
	} else {
		// Reset failure count on successful login
		if err := l.resetFailureTracking(ctx, email); err != nil {
			logger.Warn("Failed to reset failure tracking after successful login",
				logger.String("email", email),
				logger.ErrorField(err))
		}
	}

	return nil
}

// IsAccountLocked checks if an account is currently locked
func (l *LoginSecurityService) IsAccountLocked(ctx context.Context, email string) (bool, *entities.AccountLockout, error) {
	var lockout entities.AccountLockout
	err := l.db.WithContext(ctx).Where("email = ?", email).First(&lockout).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, nil // Account not locked
		}
		return false, nil, fmt.Errorf("failed to check account lockout: %w", err)
	}

	if lockout.IsLocked() {
		return true, &lockout, nil
	}

	return false, &lockout, nil
}

// GetFailureCount returns the current failure count for an email within the failure window
func (l *LoginSecurityService) GetFailureCount(ctx context.Context, email string) (int, error) {
	windowStart := time.Now().Add(-l.failureWindow)

	var count int64
	err := l.db.WithContext(ctx).Model(&entities.LoginAttempt{}).
		Where("email = ? AND success = false AND created_at > ?", email, windowStart).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to get failure count: %w", err)
	}

	return int(count), nil
}

// updateFailureTracking updates the failure tracking for an email
func (l *LoginSecurityService) updateFailureTracking(ctx context.Context, email string, userID *uint) error {
	// Get current lockout record or create new one
	var lockout entities.AccountLockout
	err := l.db.WithContext(ctx).Where("email = ?", email).First(&lockout).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to get lockout record: %w", err)
	}

	// Get current failure count in the window
	failureCount, err := l.GetFailureCount(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get failure count: %w", err)
	}

	now := time.Now()

	if err == gorm.ErrRecordNotFound {
		// Create new lockout record
		lockout = entities.AccountLockout{
			Email:       email,
			UserID:      userID,
			FailedCount: failureCount,
			LastFailure: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	} else {
		// Update existing record
		lockout.FailedCount = failureCount
		lockout.LastFailure = now
		lockout.UpdatedAt = now

		if userID != nil && lockout.UserID == nil {
			lockout.UserID = userID
		}
	}

	// Check if account should be locked
	if failureCount >= l.maxFailures && (lockout.LockedUntil == nil || lockout.LockedUntil.Before(now)) {
		lockDuration := l.calculateLockoutDuration(&lockout)
		lockedUntil := now.Add(lockDuration)
		lockout.LockedUntil = &lockedUntil
		lockout.LockReason = entities.LockReasonMultipleFailures

		logger.Warn("Account locked due to multiple failed attempts",
			logger.String("email", email),
			logger.Int("failed_count", failureCount),
			logger.Duration("lock_duration", lockDuration))
	}

	// Save or update the lockout record
	if err == gorm.ErrRecordNotFound {
		return l.db.WithContext(ctx).Create(&lockout).Error
	} else {
		return l.db.WithContext(ctx).Save(&lockout).Error
	}
}

// resetFailureTracking resets the failure tracking for an email after successful login
func (l *LoginSecurityService) resetFailureTracking(ctx context.Context, email string) error {
	return l.db.WithContext(ctx).Model(&entities.AccountLockout{}).
		Where("email = ?", email).
		Updates(map[string]any{
			"failed_count": 0,
			"locked_until": nil,
			"lock_reason":  "",
			"updated_at":   time.Now(),
		}).Error
}

// calculateLockoutDuration calculates the lockout duration, considering progressive lockouts
func (l *LoginSecurityService) calculateLockoutDuration(lockout *entities.AccountLockout) time.Duration {
	if !l.progressiveLockout {
		return l.lockoutDuration
	}

	// Count how many times this account has been locked before
	var lockCount int64
	l.db.Model(&entities.LoginAttempt{}).
		Where("email = ? AND reason = ? AND created_at > ?",
			lockout.Email, entities.LoginFailureAccountLocked, time.Now().Add(-24*time.Hour)).
		Count(&lockCount)

	// Progressive lockout: base duration * (2 ^ lock_count)
	multiplier := int64(1)
	for i := int64(0); i < lockCount && i < 5; i++ { // Cap at 32x (2^5)
		multiplier *= 2
	}

	return time.Duration(multiplier) * l.lockoutDuration
}

// UnlockAccount manually unlocks an account (admin function)
func (l *LoginSecurityService) UnlockAccount(ctx context.Context, email string, reason string) error {
	result := l.db.WithContext(ctx).Model(&entities.AccountLockout{}).
		Where("email = ?", email).
		Updates(map[string]any{
			"locked_until": nil,
			"lock_reason":  "",
			"failed_count": 0,
			"updated_at":   time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to unlock account: %w", result.Error)
	}

	logger.Info("Account manually unlocked",
		logger.String("email", email),
		logger.String("reason", reason))

	return nil
}

// GetLoginAttemptStats returns statistics about login attempts
func (l *LoginSecurityService) GetLoginAttemptStats(ctx context.Context, since time.Time) (map[string]any, error) {
	var totalAttempts, successfulAttempts, failedAttempts int64

	// Count total attempts
	if err := l.db.WithContext(ctx).Model(&entities.LoginAttempt{}).
		Where("created_at > ?", since).Count(&totalAttempts).Error; err != nil {
		return nil, fmt.Errorf("failed to count total attempts: %w", err)
	}

	// Count successful attempts
	if err := l.db.WithContext(ctx).Model(&entities.LoginAttempt{}).
		Where("created_at > ? AND success = true", since).Count(&successfulAttempts).Error; err != nil {
		return nil, fmt.Errorf("failed to count successful attempts: %w", err)
	}

	failedAttempts = totalAttempts - successfulAttempts

	// Count currently locked accounts
	var lockedAccounts int64
	if err := l.db.WithContext(ctx).Model(&entities.AccountLockout{}).
		Where("locked_until > ?", time.Now()).Count(&lockedAccounts).Error; err != nil {
		return nil, fmt.Errorf("failed to count locked accounts: %w", err)
	}

	// Get top failed IPs
	var topFailedIPs []struct {
		IP    string `json:"ip"`
		Count int64  `json:"count"`
	}

	if err := l.db.WithContext(ctx).Model(&entities.LoginAttempt{}).
		Select("ip, COUNT(*) as count").
		Where("created_at > ? AND success = false", since).
		Group("ip").
		Order("count DESC").
		Limit(10).
		Scan(&topFailedIPs).Error; err != nil {
		return nil, fmt.Errorf("failed to get top failed IPs: %w", err)
	}

	return map[string]any{
		"total_attempts":      totalAttempts,
		"successful_attempts": successfulAttempts,
		"failed_attempts":     failedAttempts,
		"success_rate":        float64(successfulAttempts) / float64(totalAttempts) * 100,
		"locked_accounts":     lockedAccounts,
		"top_failed_ips":      topFailedIPs,
	}, nil
}

// CleanupOldAttempts removes old login attempt records
func (l *LoginSecurityService) CleanupOldAttempts(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	result := l.db.WithContext(ctx).Where("created_at < ?", cutoff).Delete(&entities.LoginAttempt{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old login attempts: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		logger.Info("Cleaned up old login attempts",
			logger.Int64("count", result.RowsAffected),
			logger.String("cutoff", cutoff.Format(time.RFC3339)))
	}

	return nil
}
