package repositories

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/auth/entities"
	"linke/internal/domains/auth/usecases/interfaces"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

// loginAttemptRepository implements the LoginAttemptRepository interface
type loginAttemptRepository struct {
	db *gorm.DB
}

// accountLockoutRepository implements the AccountLockoutRepository interface
type accountLockoutRepository struct {
	db *gorm.DB
}

// NewLoginAttemptRepository creates a new LoginAttemptRepository implementation
func NewLoginAttemptRepository(db *gorm.DB) interfaces.LoginAttemptRepository {
	return &loginAttemptRepository{
		db: db,
	}
}

// NewAccountLockoutRepository creates a new AccountLockoutRepository implementation
func NewAccountLockoutRepository(db *gorm.DB) interfaces.AccountLockoutRepository {
	return &accountLockoutRepository{
		db: db,
	}
}

// === LoginAttemptRepository Implementation ===

// Create creates a new login attempt record
func (r *loginAttemptRepository) Create(ctx context.Context, attempt *entities.LoginAttempt) error {
	if err := r.db.WithContext(ctx).Create(attempt).Error; err != nil {
		logger.Error("Failed to create login attempt",
			logger.String("email", attempt.Email),
			logger.String("ip", attempt.IP),
			logger.String("success", fmt.Sprintf("%t", attempt.Success)),
			logger.Error2("error", err),
		)
		return fmt.Errorf("failed to create login attempt: %w", err)
	}

	logger.Debug("Login attempt recorded",
		logger.Uint("attempt_id", attempt.ID),
		logger.String("email", attempt.Email),
		logger.String("success", fmt.Sprintf("%t", attempt.Success)),
	)
	return nil
}

// GetByID retrieves a login attempt by ID
func (r *loginAttemptRepository) GetByID(ctx context.Context, id uint) (*entities.LoginAttempt, error) {
	var attempt entities.LoginAttempt
	if err := r.db.WithContext(ctx).First(&attempt, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("login attempt not found")
		}
		logger.Error("Failed to get login attempt by ID",
			logger.Uint("attempt_id", id),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get login attempt: %w", err)
	}
	return &attempt, nil
}

// Delete deletes a login attempt by ID
func (r *loginAttemptRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entities.LoginAttempt{}, id)
	if result.Error != nil {
		logger.Error("Failed to delete login attempt",
			logger.Uint("attempt_id", id),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to delete login attempt: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("login attempt not found")
	}

	logger.Debug("Login attempt deleted",
		logger.Uint("attempt_id", id),
	)
	return nil
}

// GetByEmail retrieves login attempts by email with pagination
func (r *loginAttemptRepository) GetByEmail(ctx context.Context, email string, limit, offset int) ([]*entities.LoginAttempt, int64, error) {
	var attempts []*entities.LoginAttempt
	var total int64

	// Count total attempts for email
	if err := r.db.WithContext(ctx).Model(&entities.LoginAttempt{}).
		Where("email = ?", email).Count(&total).Error; err != nil {
		logger.Error("Failed to count login attempts by email",
			logger.String("email", email),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count login attempts by email: %w", err)
	}

	// Get attempts with pagination
	if err := r.db.WithContext(ctx).Where("email = ?", email).
		Limit(limit).Offset(offset).Order("created_at DESC").Find(&attempts).Error; err != nil {
		logger.Error("Failed to get login attempts by email",
			logger.String("email", email),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to get login attempts by email: %w", err)
	}

	return attempts, total, nil
}

// GetByUser retrieves login attempts by user ID with pagination
func (r *loginAttemptRepository) GetByUser(ctx context.Context, userID uint, limit, offset int) ([]*entities.LoginAttempt, int64, error) {
	var attempts []*entities.LoginAttempt
	var total int64

	// Count total attempts for user
	if err := r.db.WithContext(ctx).Model(&entities.LoginAttempt{}).
		Where("user_id = ?", userID).Count(&total).Error; err != nil {
		logger.Error("Failed to count login attempts by user",
			logger.Uint("user_id", userID),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count login attempts by user: %w", err)
	}

	// Get attempts with pagination
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Limit(limit).Offset(offset).Order("created_at DESC").Find(&attempts).Error; err != nil {
		logger.Error("Failed to get login attempts by user",
			logger.Uint("user_id", userID),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to get login attempts by user: %w", err)
	}

	return attempts, total, nil
}

// GetByIP retrieves login attempts by IP with pagination
func (r *loginAttemptRepository) GetByIP(ctx context.Context, ip string, limit, offset int) ([]*entities.LoginAttempt, int64, error) {
	var attempts []*entities.LoginAttempt
	var total int64

	// Count total attempts for IP
	if err := r.db.WithContext(ctx).Model(&entities.LoginAttempt{}).
		Where("ip = ?", ip).Count(&total).Error; err != nil {
		logger.Error("Failed to count login attempts by IP",
			logger.String("ip", ip),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count login attempts by IP: %w", err)
	}

	// Get attempts with pagination
	if err := r.db.WithContext(ctx).Where("ip = ?", ip).
		Limit(limit).Offset(offset).Order("created_at DESC").Find(&attempts).Error; err != nil {
		logger.Error("Failed to get login attempts by IP",
			logger.String("ip", ip),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to get login attempts by IP: %w", err)
	}

	return attempts, total, nil
}

// GetSuccessfulAttempts retrieves successful login attempts by email with pagination
func (r *loginAttemptRepository) GetSuccessfulAttempts(ctx context.Context, email string, limit, offset int) ([]*entities.LoginAttempt, int64, error) {
	var attempts []*entities.LoginAttempt
	var total int64

	// Count total successful attempts for email
	if err := r.db.WithContext(ctx).Model(&entities.LoginAttempt{}).
		Where("email = ? AND success = ?", email, true).Count(&total).Error; err != nil {
		logger.Error("Failed to count successful login attempts by email",
			logger.String("email", email),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count successful login attempts by email: %w", err)
	}

	// Get attempts with pagination
	if err := r.db.WithContext(ctx).Where("email = ? AND success = ?", email, true).
		Limit(limit).Offset(offset).Order("created_at DESC").Find(&attempts).Error; err != nil {
		logger.Error("Failed to get successful login attempts by email",
			logger.String("email", email),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to get successful login attempts by email: %w", err)
	}

	return attempts, total, nil
}

// GetFailedAttempts retrieves failed login attempts by email with pagination
func (r *loginAttemptRepository) GetFailedAttempts(ctx context.Context, email string, limit, offset int) ([]*entities.LoginAttempt, int64, error) {
	var attempts []*entities.LoginAttempt
	var total int64

	// Count total failed attempts for email
	if err := r.db.WithContext(ctx).Model(&entities.LoginAttempt{}).
		Where("email = ? AND success = ?", email, false).Count(&total).Error; err != nil {
		logger.Error("Failed to count failed login attempts by email",
			logger.String("email", email),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count failed login attempts by email: %w", err)
	}

	// Get attempts with pagination
	if err := r.db.WithContext(ctx).Where("email = ? AND success = ?", email, false).
		Limit(limit).Offset(offset).Order("created_at DESC").Find(&attempts).Error; err != nil {
		logger.Error("Failed to get failed login attempts by email",
			logger.String("email", email),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to get failed login attempts by email: %w", err)
	}

	return attempts, total, nil
}

// GetRecentAttempts retrieves recent login attempts for an email since a specific time
func (r *loginAttemptRepository) GetRecentAttempts(ctx context.Context, email string, since time.Time) ([]*entities.LoginAttempt, error) {
	var attempts []*entities.LoginAttempt
	if err := r.db.WithContext(ctx).Where("email = ? AND created_at >= ?", email, since).
		Order("created_at DESC").Find(&attempts).Error; err != nil {
		logger.Error("Failed to get recent login attempts",
			logger.String("email", email),
			logger.String("since", since.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get recent login attempts: %w", err)
	}
	return attempts, nil
}

// GetRecentFailedAttempts retrieves recent failed login attempts for an email since a specific time
func (r *loginAttemptRepository) GetRecentFailedAttempts(ctx context.Context, email string, since time.Time) ([]*entities.LoginAttempt, error) {
	var attempts []*entities.LoginAttempt
	if err := r.db.WithContext(ctx).Where("email = ? AND success = ? AND created_at >= ?", email, false, since).
		Order("created_at DESC").Find(&attempts).Error; err != nil {
		logger.Error("Failed to get recent failed login attempts",
			logger.String("email", email),
			logger.String("since", since.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get recent failed login attempts: %w", err)
	}
	return attempts, nil
}

// GetAttemptsInTimeRange retrieves login attempts for an email within a time range
func (r *loginAttemptRepository) GetAttemptsInTimeRange(ctx context.Context, email string, start, end time.Time) ([]*entities.LoginAttempt, error) {
	var attempts []*entities.LoginAttempt
	if err := r.db.WithContext(ctx).Where("email = ? AND created_at >= ? AND created_at <= ?", email, start, end).
		Order("created_at DESC").Find(&attempts).Error; err != nil {
		logger.Error("Failed to get login attempts in time range",
			logger.String("email", email),
			logger.String("start", start.Format(time.RFC3339)),
			logger.String("end", end.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get login attempts in time range: %w", err)
	}
	return attempts, nil
}

// CountTotalAttempts returns the total count of login attempts
func (r *loginAttemptRepository) CountTotalAttempts(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.LoginAttempt{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count total login attempts: %w", err)
	}
	return count, nil
}

// CountSuccessfulAttempts returns the count of successful login attempts
func (r *loginAttemptRepository) CountSuccessfulAttempts(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.LoginAttempt{}).
		Where("success = ?", true).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count successful login attempts: %w", err)
	}
	return count, nil
}

// CountFailedAttempts returns the count of failed login attempts
func (r *loginAttemptRepository) CountFailedAttempts(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.LoginAttempt{}).
		Where("success = ?", false).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count failed login attempts: %w", err)
	}
	return count, nil
}

// CountAttemptsByEmail returns the count of login attempts for a specific email
func (r *loginAttemptRepository) CountAttemptsByEmail(ctx context.Context, email string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.LoginAttempt{}).
		Where("email = ?", email).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count login attempts by email: %w", err)
	}
	return count, nil
}

// CountFailedAttemptsByEmail returns the count of failed login attempts for a specific email since a time
func (r *loginAttemptRepository) CountFailedAttemptsByEmail(ctx context.Context, email string, since time.Time) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.LoginAttempt{}).
		Where("email = ? AND success = ? AND created_at >= ?", email, false, since).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count failed login attempts by email: %w", err)
	}
	return count, nil
}

// CountAttemptsByIP returns the count of login attempts from a specific IP since a time
func (r *loginAttemptRepository) CountAttemptsByIP(ctx context.Context, ip string, since time.Time) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.LoginAttempt{}).
		Where("ip = ? AND created_at >= ?", ip, since).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count login attempts by IP: %w", err)
	}
	return count, nil
}

// DeleteOlderThan removes login attempts older than a specific time
func (r *loginAttemptRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("created_at < ?", before).Delete(&entities.LoginAttempt{})
	if result.Error != nil {
		logger.Error("Failed to delete old login attempts",
			logger.String("before", before.Format(time.RFC3339)),
			logger.Error2("error", result.Error),
		)
		return 0, fmt.Errorf("failed to delete old login attempts: %w", result.Error)
	}

	logger.Info("Old login attempts deleted",
		logger.String("before", before.Format(time.RFC3339)),
		logger.Int64("deleted_count", result.RowsAffected),
	)
	return result.RowsAffected, nil
}

// DeleteByEmail removes all login attempts for a specific email
func (r *loginAttemptRepository) DeleteByEmail(ctx context.Context, email string) (int64, error) {
	result := r.db.WithContext(ctx).Where("email = ?", email).Delete(&entities.LoginAttempt{})
	if result.Error != nil {
		logger.Error("Failed to delete login attempts by email",
			logger.String("email", email),
			logger.Error2("error", result.Error),
		)
		return 0, fmt.Errorf("failed to delete login attempts by email: %w", result.Error)
	}

	logger.Info("Login attempts deleted by email",
		logger.String("email", email),
		logger.Int64("deleted_count", result.RowsAffected),
	)
	return result.RowsAffected, nil
}

// List lists all login attempts with pagination
func (r *loginAttemptRepository) List(ctx context.Context, limit, offset int) ([]*entities.LoginAttempt, int64, error) {
	var attempts []*entities.LoginAttempt
	var total int64

	// Count total attempts
	if err := r.db.WithContext(ctx).Model(&entities.LoginAttempt{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count login attempts", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count login attempts: %w", err)
	}

	// Get attempts with pagination
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).
		Order("created_at DESC").Find(&attempts).Error; err != nil {
		logger.Error("Failed to list login attempts", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to list login attempts: %w", err)
	}

	return attempts, total, nil
}

// ListBySuccess lists login attempts filtered by success status with pagination
func (r *loginAttemptRepository) ListBySuccess(ctx context.Context, success bool, limit, offset int) ([]*entities.LoginAttempt, int64, error) {
	var attempts []*entities.LoginAttempt
	var total int64

	// Count total attempts by success status
	if err := r.db.WithContext(ctx).Model(&entities.LoginAttempt{}).
		Where("success = ?", success).Count(&total).Error; err != nil {
		logger.Error("Failed to count login attempts by success",
			logger.String("success", fmt.Sprintf("%t", success)),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count login attempts by success: %w", err)
	}

	// Get attempts with pagination
	if err := r.db.WithContext(ctx).Where("success = ?", success).
		Limit(limit).Offset(offset).Order("created_at DESC").Find(&attempts).Error; err != nil {
		logger.Error("Failed to list login attempts by success",
			logger.String("success", fmt.Sprintf("%t", success)),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list login attempts by success: %w", err)
	}

	return attempts, total, nil
}

// ListRecent lists recent login attempts since a specific time with pagination
func (r *loginAttemptRepository) ListRecent(ctx context.Context, since time.Time, limit, offset int) ([]*entities.LoginAttempt, int64, error) {
	var attempts []*entities.LoginAttempt
	var total int64

	// Count total recent attempts
	if err := r.db.WithContext(ctx).Model(&entities.LoginAttempt{}).
		Where("created_at >= ?", since).Count(&total).Error; err != nil {
		logger.Error("Failed to count recent login attempts",
			logger.String("since", since.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count recent login attempts: %w", err)
	}

	// Get attempts with pagination
	if err := r.db.WithContext(ctx).Where("created_at >= ?", since).
		Limit(limit).Offset(offset).Order("created_at DESC").Find(&attempts).Error; err != nil {
		logger.Error("Failed to list recent login attempts",
			logger.String("since", since.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list recent login attempts: %w", err)
	}

	return attempts, total, nil
}

// === AccountLockoutRepository Implementation ===

// Create creates a new account lockout record
func (r *accountLockoutRepository) Create(ctx context.Context, lockout *entities.AccountLockout) error {
	if err := r.db.WithContext(ctx).Create(lockout).Error; err != nil {
		logger.Error("Failed to create account lockout",
			logger.String("email", lockout.Email),
			logger.String("reason", lockout.LockReason),
			logger.Error2("error", err),
		)
		return fmt.Errorf("failed to create account lockout: %w", err)
	}

	logger.Debug("Account lockout created",
		logger.Uint("lockout_id", lockout.ID),
		logger.String("email", lockout.Email),
	)
	return nil
}

// GetByID retrieves an account lockout by ID
func (r *accountLockoutRepository) GetByID(ctx context.Context, id uint) (*entities.AccountLockout, error) {
	var lockout entities.AccountLockout
	if err := r.db.WithContext(ctx).First(&lockout, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("account lockout not found")
		}
		logger.Error("Failed to get account lockout by ID",
			logger.Uint("lockout_id", id),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get account lockout: %w", err)
	}
	return &lockout, nil
}

// GetByEmail retrieves an account lockout by email
func (r *accountLockoutRepository) GetByEmail(ctx context.Context, email string) (*entities.AccountLockout, error) {
	var lockout entities.AccountLockout
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&lockout).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("account lockout not found")
		}
		logger.Error("Failed to get account lockout by email",
			logger.String("email", email),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get account lockout: %w", err)
	}
	return &lockout, nil
}

// Update updates an account lockout record
func (r *accountLockoutRepository) Update(ctx context.Context, lockout *entities.AccountLockout) error {
	if err := r.db.WithContext(ctx).Save(lockout).Error; err != nil {
		logger.Error("Failed to update account lockout",
			logger.Uint("lockout_id", lockout.ID),
			logger.Error2("error", err),
		)
		return fmt.Errorf("failed to update account lockout: %w", err)
	}

	logger.Debug("Account lockout updated",
		logger.Uint("lockout_id", lockout.ID),
	)
	return nil
}

// Delete deletes an account lockout by ID
func (r *accountLockoutRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entities.AccountLockout{}, id)
	if result.Error != nil {
		logger.Error("Failed to delete account lockout",
			logger.Uint("lockout_id", id),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to delete account lockout: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("account lockout not found")
	}

	logger.Debug("Account lockout deleted",
		logger.Uint("lockout_id", id),
	)
	return nil
}

// IsAccountLocked checks if an account is currently locked
func (r *accountLockoutRepository) IsAccountLocked(ctx context.Context, email string) (bool, *entities.AccountLockout, error) {
	var lockout entities.AccountLockout
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&lockout).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("failed to check account lockout: %w", err)
	}

	isLocked := lockout.IsLocked()
	return isLocked, &lockout, nil
}

// IncrementFailureCount increments the failure count for an account
func (r *accountLockoutRepository) IncrementFailureCount(ctx context.Context, email string, userID *uint) (*entities.AccountLockout, error) {
	var lockout entities.AccountLockout

	// Try to get existing lockout record
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&lockout).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get account lockout: %w", err)
	}

	now := time.Now()
	if err == gorm.ErrRecordNotFound {
		// Create new lockout record
		lockout = entities.AccountLockout{
			Email:       email,
			UserID:      userID,
			FailedCount: 1,
			LastFailure: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := r.Create(ctx, &lockout); err != nil {
			return nil, err
		}
	} else {
		// Update existing record
		lockout.FailedCount++
		lockout.LastFailure = now
		lockout.UpdatedAt = now
		if err := r.Update(ctx, &lockout); err != nil {
			return nil, err
		}
	}

	logger.Debug("Account failure count incremented",
		logger.String("email", email),
		logger.Int("failed_count", lockout.FailedCount),
	)
	return &lockout, nil
}

// LockAccount locks an account for a specific duration
func (r *accountLockoutRepository) LockAccount(ctx context.Context, email string, userID *uint, duration time.Duration, reason string) (*entities.AccountLockout, error) {
	var lockout entities.AccountLockout
	now := time.Now()
	lockedUntil := now.Add(duration)

	// Try to get existing lockout record
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&lockout).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to get account lockout: %w", err)
	}

	if err == gorm.ErrRecordNotFound {
		// Create new lockout record
		lockout = entities.AccountLockout{
			Email:       email,
			UserID:      userID,
			FailedCount: 0,
			LastFailure: now,
			LockedUntil: &lockedUntil,
			LockReason:  reason,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := r.Create(ctx, &lockout); err != nil {
			return nil, err
		}
	} else {
		// Update existing record
		lockout.LockedUntil = &lockedUntil
		lockout.LockReason = reason
		lockout.UpdatedAt = now
		if err := r.Update(ctx, &lockout); err != nil {
			return nil, err
		}
	}

	logger.Info("Account locked",
		logger.String("email", email),
		logger.String("reason", reason),
		logger.String("locked_until", lockedUntil.Format(time.RFC3339)),
	)
	return &lockout, nil
}

// UnlockAccount unlocks an account
func (r *accountLockoutRepository) UnlockAccount(ctx context.Context, email string, reason string) error {
	result := r.db.WithContext(ctx).Model(&entities.AccountLockout{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"locked_until": nil,
			"lock_reason":  reason,
			"updated_at":   time.Now(),
		})

	if result.Error != nil {
		logger.Error("Failed to unlock account",
			logger.String("email", email),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to unlock account: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("account lockout not found")
	}

	logger.Info("Account unlocked",
		logger.String("email", email),
		logger.String("reason", reason),
	)
	return nil
}

// ResetFailureCount resets the failure count for an account
func (r *accountLockoutRepository) ResetFailureCount(ctx context.Context, email string) error {
	result := r.db.WithContext(ctx).Model(&entities.AccountLockout{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"failed_count": 0,
			"updated_at":   time.Now(),
		})

	if result.Error != nil {
		logger.Error("Failed to reset failure count",
			logger.String("email", email),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to reset failure count: %w", result.Error)
	}

	logger.Debug("Account failure count reset",
		logger.String("email", email),
	)
	return nil
}

// GetLockedAccounts retrieves all currently locked accounts with pagination
func (r *accountLockoutRepository) GetLockedAccounts(ctx context.Context, limit, offset int) ([]*entities.AccountLockout, int64, error) {
	var lockouts []*entities.AccountLockout
	var total int64

	now := time.Now()
	condition := "locked_until IS NOT NULL AND locked_until > ?"

	// Count total locked accounts
	if err := r.db.WithContext(ctx).Model(&entities.AccountLockout{}).
		Where(condition, now).Count(&total).Error; err != nil {
		logger.Error("Failed to count locked accounts", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count locked accounts: %w", err)
	}

	// Get locked accounts with pagination
	if err := r.db.WithContext(ctx).Where(condition, now).
		Limit(limit).Offset(offset).Order("locked_until DESC").Find(&lockouts).Error; err != nil {
		logger.Error("Failed to list locked accounts", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to list locked accounts: %w", err)
	}

	return lockouts, total, nil
}

// GetAccountsWithFailures retrieves accounts with minimum failure count with pagination
func (r *accountLockoutRepository) GetAccountsWithFailures(ctx context.Context, minFailures int, limit, offset int) ([]*entities.AccountLockout, int64, error) {
	var lockouts []*entities.AccountLockout
	var total int64

	// Count total accounts with failures
	if err := r.db.WithContext(ctx).Model(&entities.AccountLockout{}).
		Where("failed_count >= ?", minFailures).Count(&total).Error; err != nil {
		logger.Error("Failed to count accounts with failures",
			logger.Int("min_failures", minFailures),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count accounts with failures: %w", err)
	}

	// Get accounts with pagination
	if err := r.db.WithContext(ctx).Where("failed_count >= ?", minFailures).
		Limit(limit).Offset(offset).Order("failed_count DESC").Find(&lockouts).Error; err != nil {
		logger.Error("Failed to list accounts with failures",
			logger.Int("min_failures", minFailures),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list accounts with failures: %w", err)
	}

	return lockouts, total, nil
}

// GetByUser retrieves account lockout by user ID
func (r *accountLockoutRepository) GetByUser(ctx context.Context, userID uint) (*entities.AccountLockout, error) {
	var lockout entities.AccountLockout
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&lockout).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("account lockout not found")
		}
		logger.Error("Failed to get account lockout by user",
			logger.Uint("user_id", userID),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get account lockout: %w", err)
	}
	return &lockout, nil
}

// CountTotalLockouts returns the total count of account lockout records
func (r *accountLockoutRepository) CountTotalLockouts(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.AccountLockout{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count total lockouts: %w", err)
	}
	return count, nil
}

// CountActiveLockouts returns the count of currently active lockouts
func (r *accountLockoutRepository) CountActiveLockouts(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.AccountLockout{}).
		Where("locked_until IS NOT NULL AND locked_until > ?", time.Now()).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count active lockouts: %w", err)
	}
	return count, nil
}

// CountLockoutsByReason returns the count of lockouts by reason
func (r *accountLockoutRepository) CountLockoutsByReason(ctx context.Context, reason string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.AccountLockout{}).
		Where("lock_reason = ?", reason).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count lockouts by reason: %w", err)
	}
	return count, nil
}

// GetLockoutStats returns lockout statistics since a specific time
func (r *accountLockoutRepository) GetLockoutStats(ctx context.Context, since time.Time) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total lockouts
	if count, err := r.CountTotalLockouts(ctx); err != nil {
		return nil, err
	} else {
		stats["total"] = count
	}

	// Active lockouts
	if count, err := r.CountActiveLockouts(ctx); err != nil {
		return nil, err
	} else {
		stats["active"] = count
	}

	// Recent lockouts
	var recentCount int64
	if err := r.db.WithContext(ctx).Model(&entities.AccountLockout{}).
		Where("created_at >= ?", since).Count(&recentCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count recent lockouts: %w", err)
	}
	stats["recent"] = recentCount

	return stats, nil
}

// CleanupExpiredLockouts removes expired lockout entries
func (r *accountLockoutRepository) CleanupExpiredLockouts(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Model(&entities.AccountLockout{}).
		Where("locked_until IS NOT NULL AND locked_until <= ?", time.Now()).
		Updates(map[string]interface{}{
			"locked_until": nil,
			"updated_at":   time.Now(),
		})

	if result.Error != nil {
		logger.Error("Failed to cleanup expired lockouts", logger.Error2("error", result.Error))
		return 0, fmt.Errorf("failed to cleanup expired lockouts: %w", result.Error)
	}

	logger.Info("Expired lockouts cleaned up",
		logger.Int64("cleaned_count", result.RowsAffected),
	)
	return result.RowsAffected, nil
}

// DeleteOlderThan removes lockout records older than a specific time
func (r *accountLockoutRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("created_at < ?", before).Delete(&entities.AccountLockout{})
	if result.Error != nil {
		logger.Error("Failed to delete old lockout records",
			logger.String("before", before.Format(time.RFC3339)),
			logger.Error2("error", result.Error),
		)
		return 0, fmt.Errorf("failed to delete old lockout records: %w", result.Error)
	}

	logger.Info("Old lockout records deleted",
		logger.String("before", before.Format(time.RFC3339)),
		logger.Int64("deleted_count", result.RowsAffected),
	)
	return result.RowsAffected, nil
}

// List lists all account lockout records with pagination
func (r *accountLockoutRepository) List(ctx context.Context, limit, offset int) ([]*entities.AccountLockout, int64, error) {
	var lockouts []*entities.AccountLockout
	var total int64

	// Count total lockouts
	if err := r.db.WithContext(ctx).Model(&entities.AccountLockout{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count account lockouts", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count account lockouts: %w", err)
	}

	// Get lockouts with pagination
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).
		Order("updated_at DESC").Find(&lockouts).Error; err != nil {
		logger.Error("Failed to list account lockouts", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to list account lockouts: %w", err)
	}

	return lockouts, total, nil
}

// ListByReason lists account lockouts filtered by reason with pagination
func (r *accountLockoutRepository) ListByReason(ctx context.Context, reason string, limit, offset int) ([]*entities.AccountLockout, int64, error) {
	var lockouts []*entities.AccountLockout
	var total int64

	// Count total lockouts by reason
	if err := r.db.WithContext(ctx).Model(&entities.AccountLockout{}).
		Where("lock_reason = ?", reason).Count(&total).Error; err != nil {
		logger.Error("Failed to count account lockouts by reason",
			logger.String("reason", reason),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count account lockouts by reason: %w", err)
	}

	// Get lockouts with pagination
	if err := r.db.WithContext(ctx).Where("lock_reason = ?", reason).
		Limit(limit).Offset(offset).Order("updated_at DESC").Find(&lockouts).Error; err != nil {
		logger.Error("Failed to list account lockouts by reason",
			logger.String("reason", reason),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list account lockouts by reason: %w", err)
	}

	return lockouts, total, nil
}

// ListRecent lists recent account lockouts since a specific time with pagination
func (r *accountLockoutRepository) ListRecent(ctx context.Context, since time.Time, limit, offset int) ([]*entities.AccountLockout, int64, error) {
	var lockouts []*entities.AccountLockout
	var total int64

	// Count total recent lockouts
	if err := r.db.WithContext(ctx).Model(&entities.AccountLockout{}).
		Where("created_at >= ?", since).Count(&total).Error; err != nil {
		logger.Error("Failed to count recent account lockouts",
			logger.String("since", since.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count recent account lockouts: %w", err)
	}

	// Get lockouts with pagination
	if err := r.db.WithContext(ctx).Where("created_at >= ?", since).
		Limit(limit).Offset(offset).Order("created_at DESC").Find(&lockouts).Error; err != nil {
		logger.Error("Failed to list recent account lockouts",
			logger.String("since", since.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list recent account lockouts: %w", err)
	}

	return lockouts, total, nil
}
