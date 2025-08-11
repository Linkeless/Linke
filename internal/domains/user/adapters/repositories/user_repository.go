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


// 注意：具体的 GetByEmail、GetByGoogleID 等方法已被移除
// 现在统一使用通用的 GetByField 和 GetActiveByField 方法
// 例如：
// - GetByEmail → GetByField(ctx, "email", email)
// - GetByGoogleID → GetByField(ctx, "google_id", googleID)
// - GetActiveByEmail → GetActiveByField(ctx, "email", email)


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


// 注意：CountByProvider 已被移除，使用通用方法 CountByField(ctx, "provider", provider) 替代


// CountRecentSignups returns the count of users registered in the last N days
func (r *userRepository) CountRecentSignups(ctx context.Context, days int) (int64, error) {
	var count int64
	since := time.Now().AddDate(0, 0, -days)
	if err := r.GetDB().WithContext(ctx).Model(&entities.User{}).Where("created_at >= ?", since).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count recent signups: %w", err)
	}
	return count, nil
}


// 注意：ExistsByEmail 已被移除，使用通用方法 ExistsByField(ctx, "email", email) 替代


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

// 注意：CountByInviteCodeUsed 已被移除，使用通用方法 CountByField(ctx, "invite_code_used", inviteCode) 替代

// Generic field-based methods

// GetByField retrieves a user by a specific field value
func (r *userRepository) GetByField(ctx context.Context, field string, value interface{}) (*entities.User, error) {
	var user entities.User
	if err := r.GetDB().WithContext(ctx).Where(field+" = ?", value).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by %s: %w", field, err)
	}
	return &user, nil
}

// GetActiveByField retrieves an active user by a specific field value
func (r *userRepository) GetActiveByField(ctx context.Context, field string, value interface{}) (*entities.User, error) {
	var user entities.User
	if err := r.GetDB().WithContext(ctx).Where(field+" = ? AND status = 'active'", value).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("active user not found")
		}
		return nil, fmt.Errorf("failed to get active user by %s: %w", field, err)
	}
	return &user, nil
}

// ExistsByField checks if a user exists with a specific field value
func (r *userRepository) ExistsByField(ctx context.Context, field string, value interface{}) (bool, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.User{}).Where(field+" = ?", value).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user existence by %s: %w", field, err)
	}
	return count > 0, nil
}

// ListByField lists users matching a specific field value
func (r *userRepository) ListByField(ctx context.Context, field string, value interface{}, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64
	
	query := r.GetDB().WithContext(ctx).Model(&entities.User{}).Where(field+" = ?", value)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users by %s: %w", field, err)
	}
	
	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users by %s: %w", field, err)
	}
	
	return users, total, nil
}

// CountByField returns the count of users matching a specific field value
func (r *userRepository) CountByField(ctx context.Context, field string, value interface{}) (int64, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.User{}).Where(field+" = ?", value).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users by field %s: %w", field, err)
	}
	return count, nil
}

// GetMultipleByField retrieves multiple users by field values
func (r *userRepository) GetMultipleByField(ctx context.Context, field string, values []interface{}) ([]*entities.User, error) {
	var users []*entities.User
	if err := r.GetDB().WithContext(ctx).Where(field+" IN ?", values).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get multiple users by %s: %w", field, err)
	}
	return users, nil
}
