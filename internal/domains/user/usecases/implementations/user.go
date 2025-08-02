package implementations

import (
	"context"
	"fmt"
	"strings"
	"time"

	"linke/internal/shared/framework"
	"linke/internal/shared/logger"
	"linke/internal/domains/user/entities"

	"gorm.io/gorm"
)

type UserService struct {
	db     *gorm.DB
	logger framework.Logger
}

func NewUserService(db *gorm.DB, logger framework.Logger) *UserService {
	return &UserService{
		db:     db,
		logger: logger,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, user *entities.User) error {
	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		s.logger.Error("Failed to create user",
			logger.String("email", user.Email),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("User created successfully",
		logger.Uint("user_id", user.ID),
		logger.String("email", user.Email),
	)
	return nil
}

// GetUserByID retrieves a user by ID (excludes soft deleted)
func (s *UserService) GetUserByID(ctx context.Context, id uint) (*entities.User, error) {
	var user entities.User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		s.logger.Error("Failed to get user by ID",
			logger.Uint("user_id", id),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email (excludes soft deleted)
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		s.logger.Error("Failed to get user by email",
			logger.String("email", email),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetActiveUserByID retrieves an active user by ID (excludes soft deleted and inactive users)
func (s *UserService) GetActiveUserByID(ctx context.Context, id uint) (*entities.User, error) {
	var user entities.User
	if err := s.db.WithContext(ctx).Where("status = ?", entities.UserStatusActive).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("active user not found")
		}
		s.logger.Error("Failed to get active user by ID",
			logger.Uint("user_id", id),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get active user: %w", err)
	}
	return &user, nil
}

// GetActiveUserByEmail retrieves an active user by email (excludes soft deleted and inactive users)
func (s *UserService) GetActiveUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	if err := s.db.WithContext(ctx).Where("email = ? AND status = ?", email, entities.UserStatusActive).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("active user not found")
		}
		s.logger.Error("Failed to get active user by email",
			logger.String("email", email),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get active user: %w", err)
	}
	return &user, nil
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(ctx context.Context, user *entities.User) error {
	if err := s.db.WithContext(ctx).Save(user).Error; err != nil {
		s.logger.Error("Failed to update user",
			logger.Uint("user_id", user.ID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.Info("User updated successfully",
		logger.Uint("user_id", user.ID),
	)
	return nil
}

// SoftDeleteUser performs soft delete on a user
func (s *UserService) SoftDeleteUser(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&entities.User{}, id)
	if result.Error != nil {
		s.logger.Error("Failed to soft delete user",
			logger.Uint("user_id", id),
			logger.ErrorField(result.Error),
		)
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	s.logger.Info("User soft deleted successfully",
		logger.Uint("user_id", id),
	)
	return nil
}

// RestoreUser restores a soft deleted user
func (s *UserService) RestoreUser(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Unscoped().Model(&entities.User{}).Where("id = ?", id).Update("deleted_at", nil)
	if result.Error != nil {
		s.logger.Error("Failed to restore user",
			logger.Uint("user_id", id),
			logger.ErrorField(result.Error),
		)
		return fmt.Errorf("failed to restore user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	s.logger.Info("User restored successfully",
		logger.Uint("user_id", id),
	)
	return nil
}

// ListUsers lists all active users with pagination
func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	// Count total active users
	if err := s.db.WithContext(ctx).Model(&entities.User{}).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count users", logger.ErrorField(err))
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users with pagination
	if err := s.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		s.logger.Error("Failed to list users", logger.ErrorField(err))
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// ListDeletedUsers lists all soft deleted users with pagination
func (s *UserService) ListDeletedUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	// Count total deleted users
	if err := s.db.WithContext(ctx).Unscoped().Model(&entities.User{}).Where("deleted_at IS NOT NULL").Count(&total).Error; err != nil {
		s.logger.Error("Failed to count deleted users", logger.ErrorField(err))
		return nil, 0, fmt.Errorf("failed to count deleted users: %w", err)
	}

	// Get deleted users with pagination
	if err := s.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		s.logger.Error("Failed to list deleted users", logger.ErrorField(err))
		return nil, 0, fmt.Errorf("failed to list deleted users: %w", err)
	}

	return users, total, nil
}

// HardDeleteUser permanently deletes a user
func (s *UserService) HardDeleteUser(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Unscoped().Delete(&entities.User{}, id)
	if result.Error != nil {
		s.logger.Error("Failed to hard delete user",
			logger.Uint("user_id", id),
			logger.ErrorField(result.Error),
		)
		return fmt.Errorf("failed to permanently delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	s.logger.Warn("User permanently deleted",
		logger.Uint("user_id", id),
	)
	return nil
}

// SearchUsers searches users by name, email, or username
func (s *UserService) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	// Prepare search query
	searchQuery := "%" + strings.ToLower(query) + "%"
	whereClause := "LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(username) LIKE ?"

	// Count total matching users
	if err := s.db.WithContext(ctx).Model(&entities.User{}).Where(whereClause, searchQuery, searchQuery, searchQuery).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count search results", 
			logger.String("query", query),
			logger.ErrorField(err),
		)
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Get matching users with pagination
	if err := s.db.WithContext(ctx).Where(whereClause, searchQuery, searchQuery, searchQuery).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		s.logger.Error("Failed to search users", 
			logger.String("query", query),
			logger.ErrorField(err),
		)
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}

	return users, total, nil
}

// UpdateUserStatus updates a user's status
func (s *UserService) UpdateUserStatus(ctx context.Context, id uint, status string) (*entities.User, error) {
	var user entities.User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Status = status
	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		s.logger.Error("Failed to update user status",
			logger.Uint("user_id", id),
			logger.String("status", status),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	s.logger.Info("User status updated successfully",
		logger.Uint("user_id", id),
		logger.String("new_status", status),
	)
	return &user, nil
}

// UpdateUserRole updates a user's role
func (s *UserService) UpdateUserRole(ctx context.Context, id uint, role string) (*entities.User, error) {
	var user entities.User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Role = role
	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		s.logger.Error("Failed to update user role",
			logger.Uint("user_id", id),
			logger.String("role", role),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to update user role: %w", err)
	}

	s.logger.Info("User role updated successfully",
		logger.Uint("user_id", id),
		logger.String("new_role", role),
	)
	return &user, nil
}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers    int64            `json:"total_users"`
	ActiveUsers   int64            `json:"active_users"`
	InactiveUsers int64            `json:"inactive_users"`
	BannedUsers   int64            `json:"banned_users"`
	DeletedUsers  int64            `json:"deleted_users"`
	ByProvider    map[string]int64 `json:"by_provider"`
	RecentSignups int64            `json:"recent_signups"`
}

// GetUserStats returns user statistics
func (s *UserService) GetUserStats(ctx context.Context) (*UserStats, error) {
	stats := &UserStats{
		ByProvider: make(map[string]int64),
	}

	// Total users (excluding deleted)
	if err := s.db.WithContext(ctx).Model(&entities.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}

	// Users by status
	statusQueries := map[string]*int64{
		entities.UserStatusActive:   &stats.ActiveUsers,
		entities.UserStatusInactive: &stats.InactiveUsers,
		entities.UserStatusBanned:   &stats.BannedUsers,
	}

	for status, count := range statusQueries {
		if err := s.db.WithContext(ctx).Model(&entities.User{}).Where("status = ?", status).Count(count).Error; err != nil {
			return nil, fmt.Errorf("failed to count %s users: %w", status, err)
		}
	}

	// Deleted users
	if err := s.db.WithContext(ctx).Unscoped().Model(&entities.User{}).Where("deleted_at IS NOT NULL").Count(&stats.DeletedUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count deleted users: %w", err)
	}

	// Users by provider
	providers := []string{entities.ProviderLocal, entities.ProviderGoogle, entities.ProviderGitHub, entities.ProviderTelegram}
	for _, provider := range providers {
		var count int64
		if err := s.db.WithContext(ctx).Model(&entities.User{}).Where("provider = ?", provider).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("failed to count users for provider %s: %w", provider, err)
		}
		stats.ByProvider[provider] = count
	}

	// Recent signups (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	if err := s.db.WithContext(ctx).Model(&entities.User{}).Where("created_at >= ?", thirtyDaysAgo).Count(&stats.RecentSignups).Error; err != nil {
		return nil, fmt.Errorf("failed to count recent signups: %w", err)
	}

	return stats, nil
}

// BatchOperationResult represents the result of batch operations
type BatchOperationResult struct {
	DeletedCount  int    `json:"deleted_count,omitempty"`
	RestoredCount int    `json:"restored_count,omitempty"`
	FailedIDs     []uint `json:"failed_ids,omitempty"`
}

// BatchDeleteUsers performs batch soft delete on multiple users
func (s *UserService) BatchDeleteUsers(ctx context.Context, ids []uint) (*BatchOperationResult, error) {
	result := &BatchOperationResult{}

	// Validate that users exist and are not already deleted
	var existingUsers []entities.User
	if err := s.db.WithContext(ctx).Where("id IN ?", ids).Find(&existingUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to validate users: %w", err)
	}

	// Create map of existing user IDs
	existingIDs := make(map[uint]bool)
	for _, user := range existingUsers {
		existingIDs[user.ID] = true
	}

	// Delete existing users and track failed IDs
	for _, id := range ids {
		if !existingIDs[id] {
			result.FailedIDs = append(result.FailedIDs, id)
			continue
		}

		deleteResult := s.db.WithContext(ctx).Delete(&entities.User{}, id)
		if deleteResult.Error != nil {
			s.logger.Error("Failed to delete user in batch",
				logger.Uint("user_id", id),
				logger.ErrorField(deleteResult.Error),
			)
			result.FailedIDs = append(result.FailedIDs, id)
			continue
		}

		if deleteResult.RowsAffected > 0 {
			result.DeletedCount++
		} else {
			result.FailedIDs = append(result.FailedIDs, id)
		}
	}

	s.logger.Info("Batch delete completed",
		logger.Int("deleted_count", result.DeletedCount),
		logger.Int("failed_count", len(result.FailedIDs)),
	)

	return result, nil
}

// BatchRestoreUsers performs batch restore on multiple soft deleted users
func (s *UserService) BatchRestoreUsers(ctx context.Context, ids []uint) (*BatchOperationResult, error) {
	result := &BatchOperationResult{}

	// Validate that users exist and are deleted
	var deletedUsers []entities.User
	if err := s.db.WithContext(ctx).Unscoped().Where("id IN ? AND deleted_at IS NOT NULL", ids).Find(&deletedUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to validate deleted users: %w", err)
	}

	// Create map of existing deleted user IDs
	deletedIDs := make(map[uint]bool)
	for _, user := range deletedUsers {
		deletedIDs[user.ID] = true
	}

	// Restore deleted users and track failed IDs
	for _, id := range ids {
		if !deletedIDs[id] {
			result.FailedIDs = append(result.FailedIDs, id)
			continue
		}

		restoreResult := s.db.WithContext(ctx).Unscoped().Model(&entities.User{}).Where("id = ?", id).Update("deleted_at", nil)
		if restoreResult.Error != nil {
			s.logger.Error("Failed to restore user in batch",
				logger.Uint("user_id", id),
				logger.ErrorField(restoreResult.Error),
			)
			result.FailedIDs = append(result.FailedIDs, id)
			continue
		}

		if restoreResult.RowsAffected > 0 {
			result.RestoredCount++
		} else {
			result.FailedIDs = append(result.FailedIDs, id)
		}
	}

	s.logger.Info("Batch restore completed",
		logger.Int("restored_count", result.RestoredCount),
		logger.Int("failed_count", len(result.FailedIDs)),
	)

	return result, nil
}

// ListUsersByProvider lists users filtered by OAuth provider
func (s *UserService) ListUsersByProvider(ctx context.Context, provider string, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	// Count total users for the provider
	if err := s.db.WithContext(ctx).Model(&entities.User{}).Where("provider = ?", provider).Count(&total).Error; err != nil {
		s.logger.Error("Failed to count users by provider", 
			logger.String("provider", provider),
			logger.ErrorField(err),
		)
		return nil, 0, fmt.Errorf("failed to count users by provider: %w", err)
	}

	// Get users with pagination
	if err := s.db.WithContext(ctx).Where("provider = ?", provider).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		s.logger.Error("Failed to list users by provider", 
			logger.String("provider", provider),
			logger.ErrorField(err),
		)
		return nil, 0, fmt.Errorf("failed to list users by provider: %w", err)
	}

	return users, total, nil
}