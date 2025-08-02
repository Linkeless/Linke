package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"linke/internal/domains/user/entities"
	"linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

// userRepository implements the UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository implementation
func NewUserRepository(db *gorm.DB) interfaces.UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create creates a new user in the database
func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		logger.Error("Failed to create user in repository",
			logger.String("email", user.Email),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to create user: %w", err)
	}

	logger.Debug("User created successfully in repository",
		logger.Uint("user_id", user.ID),
		logger.String("email", user.Email),
	)
	return nil
}

// GetByID retrieves a user by ID (excludes soft deleted)
func (r *userRepository) GetByID(ctx context.Context, id uint) (*entities.User, error) {
	var user entities.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		logger.Error("Failed to get user by ID",
			logger.Uint("user_id", id),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetByEmail retrieves a user by email (excludes soft deleted)
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
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
	if err := r.db.WithContext(ctx).Where("google_id = ?", googleID).First(&user).Error; err != nil {
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
	if err := r.db.WithContext(ctx).Where("github_id = ?", githubID).First(&user).Error; err != nil {
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
	if err := r.db.WithContext(ctx).Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
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
	if err := r.db.WithContext(ctx).Where("status = ?", entities.UserStatusActive).First(&user, id).Error; err != nil {
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
	if err := r.db.WithContext(ctx).Where("email = ? AND status = ?", email, entities.UserStatusActive).First(&user).Error; err != nil {
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

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *entities.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		logger.Error("Failed to update user in repository",
			logger.Uint("user_id", user.ID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update user: %w", err)
	}

	logger.Debug("User updated successfully in repository",
		logger.Uint("user_id", user.ID),
	)
	return nil
}

// Delete performs soft delete on a user
func (r *userRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entities.User{}, id)
	if result.Error != nil {
		logger.Error("Failed to soft delete user",
			logger.Uint("user_id", id),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	logger.Debug("User soft deleted successfully",
		logger.Uint("user_id", id),
	)
	return nil
}

// SoftDelete performs soft delete on a user (alias for Delete)
func (r *userRepository) SoftDelete(ctx context.Context, id uint) error {
	return r.Delete(ctx, id)
}

// Restore restores a soft deleted user
func (r *userRepository) Restore(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Unscoped().Model(&entities.User{}).Where("id = ?", id).Update("deleted_at", nil)
	if result.Error != nil {
		logger.Error("Failed to restore user",
			logger.Uint("user_id", id),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to restore user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	logger.Debug("User restored successfully",
		logger.Uint("user_id", id),
	)
	return nil
}

// HardDelete permanently deletes a user
func (r *userRepository) HardDelete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Unscoped().Delete(&entities.User{}, id)
	if result.Error != nil {
		logger.Error("Failed to hard delete user",
			logger.Uint("user_id", id),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to permanently delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	logger.Warn("User permanently deleted",
		logger.Uint("user_id", id),
	)
	return nil
}

// List lists all active users with pagination
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	// Count total active users
	if err := r.db.WithContext(ctx).Model(&entities.User{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count users", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users with pagination
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		logger.Error("Failed to list users", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// ListDeleted lists all soft deleted users with pagination
func (r *userRepository) ListDeleted(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	// Count total deleted users
	if err := r.db.WithContext(ctx).Unscoped().Model(&entities.User{}).Where("deleted_at IS NOT NULL").Count(&total).Error; err != nil {
		logger.Error("Failed to count deleted users", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count deleted users: %w", err)
	}

	// Get deleted users with pagination
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		logger.Error("Failed to list deleted users", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to list deleted users: %w", err)
	}

	return users, total, nil
}

// ListByProvider lists users filtered by OAuth provider
func (r *userRepository) ListByProvider(ctx context.Context, provider string, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	// Count total users for the provider
	if err := r.db.WithContext(ctx).Model(&entities.User{}).Where("provider = ?", provider).Count(&total).Error; err != nil {
		logger.Error("Failed to count users by provider",
			logger.String("provider", provider),
			logger.ErrorField(err),
		)
		return nil, 0, fmt.Errorf("failed to count users by provider: %w", err)
	}

	// Get users with pagination
	if err := r.db.WithContext(ctx).Where("provider = ?", provider).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		logger.Error("Failed to list users by provider",
			logger.String("provider", provider),
			logger.ErrorField(err),
		)
		return nil, 0, fmt.Errorf("failed to list users by provider: %w", err)
	}

	return users, total, nil
}

// ListByStatus lists users filtered by status
func (r *userRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	// Count total users for the status
	if err := r.db.WithContext(ctx).Model(&entities.User{}).Where("status = ?", status).Count(&total).Error; err != nil {
		logger.Error("Failed to count users by status",
			logger.String("status", status),
			logger.ErrorField(err),
		)
		return nil, 0, fmt.Errorf("failed to count users by status: %w", err)
	}

	// Get users with pagination
	if err := r.db.WithContext(ctx).Where("status = ?", status).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		logger.Error("Failed to list users by status",
			logger.String("status", status),
			logger.ErrorField(err),
		)
		return nil, 0, fmt.Errorf("failed to list users by status: %w", err)
	}

	return users, total, nil
}

// ListByRole lists users filtered by role
func (r *userRepository) ListByRole(ctx context.Context, role string, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	// Count total users for the role
	if err := r.db.WithContext(ctx).Model(&entities.User{}).Where("role = ?", role).Count(&total).Error; err != nil {
		logger.Error("Failed to count users by role",
			logger.String("role", role),
			logger.ErrorField(err),
		)
		return nil, 0, fmt.Errorf("failed to count users by role: %w", err)
	}

	// Get users with pagination
	if err := r.db.WithContext(ctx).Where("role = ?", role).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		logger.Error("Failed to list users by role",
			logger.String("role", role),
			logger.ErrorField(err),
		)
		return nil, 0, fmt.Errorf("failed to list users by role: %w", err)
	}

	return users, total, nil
}

// Search searches users by name, email, or username
func (r *userRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	// Prepare search query
	searchQuery := "%" + strings.ToLower(query) + "%"
	whereClause := "LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(username) LIKE ?"

	// Count total matching users
	if err := r.db.WithContext(ctx).Model(&entities.User{}).Where(whereClause, searchQuery, searchQuery, searchQuery).Count(&total).Error; err != nil {
		logger.Error("Failed to count search results",
			logger.String("query", query),
			logger.ErrorField(err),
		)
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Get matching users with pagination
	if err := r.db.WithContext(ctx).Where(whereClause, searchQuery, searchQuery, searchQuery).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		logger.Error("Failed to search users",
			logger.String("query", query),
			logger.ErrorField(err),
		)
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}

	return users, total, nil
}

// UpdateStatus updates a user's status
func (r *userRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	result := r.db.WithContext(ctx).Model(&entities.User{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		logger.Error("Failed to update user status",
			logger.Uint("user_id", id),
			logger.String("status", status),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to update user status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	logger.Debug("User status updated successfully",
		logger.Uint("user_id", id),
		logger.String("new_status", status),
	)
	return nil
}

// UpdateRole updates a user's role
func (r *userRepository) UpdateRole(ctx context.Context, id uint, role string) error {
	result := r.db.WithContext(ctx).Model(&entities.User{}).Where("id = ?", id).Update("role", role)
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

// CountTotal returns the total count of users (excluding soft deleted)
func (r *userRepository) CountTotal(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.User{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count total users: %w", err)
	}
	return count, nil
}

// CountByStatus returns the count of users by status
func (r *userRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.User{}).Where("status = ?", status).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users by status %s: %w", status, err)
	}
	return count, nil
}

// CountByProvider returns the count of users by provider
func (r *userRepository) CountByProvider(ctx context.Context, provider string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.User{}).Where("provider = ?", provider).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users for provider %s: %w", provider, err)
	}
	return count, nil
}

// CountDeleted returns the count of soft deleted users
func (r *userRepository) CountDeleted(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Unscoped().Model(&entities.User{}).Where("deleted_at IS NOT NULL").Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count deleted users: %w", err)
	}
	return count, nil
}

// CountRecentSignups returns the count of users registered in the last N days
func (r *userRepository) CountRecentSignups(ctx context.Context, days int) (int64, error) {
	var count int64
	since := time.Now().AddDate(0, 0, -days)
	if err := r.db.WithContext(ctx).Model(&entities.User{}).Where("created_at >= ?", since).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count recent signups: %w", err)
	}
	return count, nil
}

// BatchDelete performs batch soft delete on multiple users
func (r *userRepository) BatchDelete(ctx context.Context, ids []uint) (int, []uint, error) {
	var deletedCount int
	var failedIDs []uint

	// Validate that users exist and are not already deleted
	var existingUsers []entities.User
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&existingUsers).Error; err != nil {
		return 0, nil, fmt.Errorf("failed to validate users: %w", err)
	}

	// Create map of existing user IDs
	existingIDs := make(map[uint]bool)
	for _, user := range existingUsers {
		existingIDs[user.ID] = true
	}

	// Delete existing users and track failed IDs
	for _, id := range ids {
		if !existingIDs[id] {
			failedIDs = append(failedIDs, id)
			continue
		}

		deleteResult := r.db.WithContext(ctx).Delete(&entities.User{}, id)
		if deleteResult.Error != nil {
			logger.Error("Failed to delete user in batch",
				logger.Uint("user_id", id),
				logger.Error2("error", deleteResult.Error),
			)
			failedIDs = append(failedIDs, id)
			continue
		}

		if deleteResult.RowsAffected > 0 {
			deletedCount++
		} else {
			failedIDs = append(failedIDs, id)
		}
	}

	logger.Debug("Batch delete completed",
		logger.Int("deleted_count", deletedCount),
		logger.Int("failed_count", len(failedIDs)),
	)

	return deletedCount, failedIDs, nil
}

// BatchRestore performs batch restore on multiple soft deleted users
func (r *userRepository) BatchRestore(ctx context.Context, ids []uint) (int, []uint, error) {
	var restoredCount int
	var failedIDs []uint

	// Validate that users exist and are deleted
	var deletedUsers []entities.User
	if err := r.db.WithContext(ctx).Unscoped().Where("id IN ? AND deleted_at IS NOT NULL", ids).Find(&deletedUsers).Error; err != nil {
		return 0, nil, fmt.Errorf("failed to validate deleted users: %w", err)
	}

	// Create map of existing deleted user IDs
	deletedIDs := make(map[uint]bool)
	for _, user := range deletedUsers {
		deletedIDs[user.ID] = true
	}

	// Restore deleted users and track failed IDs
	for _, id := range ids {
		if !deletedIDs[id] {
			failedIDs = append(failedIDs, id)
			continue
		}

		restoreResult := r.db.WithContext(ctx).Unscoped().Model(&entities.User{}).Where("id = ?", id).Update("deleted_at", nil)
		if restoreResult.Error != nil {
			logger.Error("Failed to restore user in batch",
				logger.Uint("user_id", id),
				logger.Error2("error", restoreResult.Error),
			)
			failedIDs = append(failedIDs, id)
			continue
		}

		if restoreResult.RowsAffected > 0 {
			restoredCount++
		} else {
			failedIDs = append(failedIDs, id)
		}
	}

	logger.Debug("Batch restore completed",
		logger.Int("restored_count", restoredCount),
		logger.Int("failed_count", len(failedIDs)),
	)

	return restoredCount, failedIDs, nil
}

// ExistsByEmail checks if a user with the given email exists
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user existence by email: %w", err)
	}
	return count > 0, nil
}

// ExistsByID checks if a user with the given ID exists
func (r *userRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.User{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user existence by ID: %w", err)
	}
	return count > 0, nil
}

// GetByInviteCodeUsed retrieves users who used a specific invite code
func (r *userRepository) GetByInviteCodeUsed(ctx context.Context, inviteCode string) ([]*entities.User, error) {
	var users []*entities.User
	if err := r.db.WithContext(ctx).Where("invite_code_used = ?", inviteCode).Find(&users).Error; err != nil {
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
	if err := r.db.WithContext(ctx).Model(&entities.User{}).Where("invite_code_used = ?", inviteCode).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users by invite code %s: %w", inviteCode, err)
	}
	return count, nil
}
