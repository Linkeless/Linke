package repositories

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/user/entities"
	"linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/framework"
	"linke/internal/shared/logger"
	"linke/internal/shared/repository"

	"gorm.io/gorm"
)

// userRepository implements the UserRepository interface
type userRepository struct {
	*repository.BaseRepositoryImpl[entities.User, uint]
}

// NewUserRepository creates a new UserRepository implementation
func NewUserRepository(db *gorm.DB, frameworkLogger framework.Logger) interfaces.UserRepository {
	return &userRepository{
		BaseRepositoryImpl: repository.NewBaseRepository[entities.User, uint](db, frameworkLogger),
	}
}


// GetByEmail retrieves a user by email (excludes soft deleted)
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	if err := r.GetDB().WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		logger.Error("Failed to get user by email",
			logger.String("email", email),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetByGoogleID retrieves a user by Google ID
func (r *userRepository) GetByGoogleID(ctx context.Context, googleID string) (*entities.User, error) {
	var user entities.User
	if err := r.GetDB().WithContext(ctx).Where("google_id = ?", googleID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		logger.Error("Failed to get user by Google ID",
			logger.String("google_id", googleID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetByGitHubID retrieves a user by GitHub ID
func (r *userRepository) GetByGitHubID(ctx context.Context, githubID string) (*entities.User, error) {
	var user entities.User
	if err := r.GetDB().WithContext(ctx).Where("github_id = ?", githubID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		logger.Error("Failed to get user by GitHub ID",
			logger.String("github_id", githubID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetByTelegramID retrieves a user by Telegram ID
func (r *userRepository) GetByTelegramID(ctx context.Context, telegramID string) (*entities.User, error) {
	var user entities.User
	if err := r.GetDB().WithContext(ctx).Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		logger.Error("Failed to get user by Telegram ID",
			logger.String("telegram_id", telegramID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetActiveByID retrieves an active user by ID (excludes soft deleted and inactive users)
func (r *userRepository) GetActiveByID(ctx context.Context, id uint) (*entities.User, error) {
	var user entities.User
	if err := r.GetDB().WithContext(ctx).Where("status = ?", entities.UserStatusActive).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("active user not found")
		}
		logger.Error("Failed to get active user by ID",
			logger.Uint("user_id", id),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get active user: %w", err)
	}
	return &user, nil
}

// GetActiveByEmail retrieves an active user by email (excludes soft deleted and inactive users)
func (r *userRepository) GetActiveByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	if err := r.GetDB().WithContext(ctx).Where("email = ? AND status = ?", email, entities.UserStatusActive).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("active user not found")
		}
		logger.Error("Failed to get active user by email",
			logger.String("email", email),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get active user: %w", err)
	}
	return &user, nil
}


// ListByProvider lists users filtered by OAuth provider
func (r *userRepository) ListByProvider(ctx context.Context, provider string, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	// Count total users for the provider
	if err := r.GetDB().WithContext(ctx).Model(&entities.User{}).Where("provider = ?", provider).Count(&total).Error; err != nil {
		logger.Error("Failed to count users by provider",
			logger.String("provider", provider),
			logger.ErrorField(err),
		)
		return nil, 0, fmt.Errorf("failed to count users by provider: %w", err)
	}

	// Get users with pagination
	if err := r.GetDB().WithContext(ctx).Where("provider = ?", provider).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		logger.Error("Failed to list users by provider",
			logger.String("provider", provider),
			logger.ErrorField(err),
		)
		return nil, 0, fmt.Errorf("failed to list users by provider: %w", err)
	}

	return users, total, nil
}


// ListByRole lists users filtered by role
func (r *userRepository) ListByRole(ctx context.Context, role string, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	// Count total users for the role
	if err := r.GetDB().WithContext(ctx).Model(&entities.User{}).Where("role = ?", role).Count(&total).Error; err != nil {
		logger.Error("Failed to count users by role",
			logger.String("role", role),
			logger.ErrorField(err),
		)
		return nil, 0, fmt.Errorf("failed to count users by role: %w", err)
	}

	// Get users with pagination
	if err := r.GetDB().WithContext(ctx).Where("role = ?", role).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		logger.Error("Failed to list users by role",
			logger.String("role", role),
			logger.ErrorField(err),
		)
		return nil, 0, fmt.Errorf("failed to list users by role: %w", err)
	}

	return users, total, nil
}


// UpdateRole updates a user's role
func (r *userRepository) UpdateRole(ctx context.Context, id uint, role string) error {
	result := r.GetDB().WithContext(ctx).Model(&entities.User{}).Where("id = ?", id).Update("role", role)
	if result.Error != nil {
		logger.Error("Failed to update user role",
			logger.Uint("user_id", id),
			logger.String("role", role),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to update user role: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	logger.Debug("User role updated successfully",
		logger.Uint("user_id", id),
		logger.String("new_role", role),
	)
	return nil
}


// CountByProvider returns the count of users by provider
func (r *userRepository) CountByProvider(ctx context.Context, provider string) (int64, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.User{}).Where("provider = ?", provider).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users for provider %s: %w", provider, err)
	}
	return count, nil
}


// CountRecentSignups returns the count of users registered in the last N days
func (r *userRepository) CountRecentSignups(ctx context.Context, days int) (int64, error) {
	var count int64
	since := time.Now().AddDate(0, 0, -days)
	if err := r.GetDB().WithContext(ctx).Model(&entities.User{}).Where("created_at >= ?", since).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count recent signups: %w", err)
	}
	return count, nil
}


// ExistsByEmail checks if a user with the given email exists
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user existence by email: %w", err)
	}
	return count > 0, nil
}


// GetByInviteCodeUsed retrieves users who used a specific invite code
func (r *userRepository) GetByInviteCodeUsed(ctx context.Context, inviteCode string) ([]*entities.User, error) {
	var users []*entities.User
	if err := r.GetDB().WithContext(ctx).Where("invite_code_used = ?", inviteCode).Find(&users).Error; err != nil {
		logger.Error("Failed to get users by invite code used",
			logger.String("invite_code", inviteCode),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get users by invite code: %w", err)
	}
	return users, nil
}

// CountByInviteCodeUsed returns the count of users who used a specific invite code
func (r *userRepository) CountByInviteCodeUsed(ctx context.Context, inviteCode string) (int64, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.User{}).Where("invite_code_used = ?", inviteCode).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users by invite code %s: %w", inviteCode, err)
	}
	return count, nil
}
