package implementations

import (
	"context"
	"fmt"
	"strings"
	"time"

	"linke/internal/domains/user/entities"
	"linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/framework"
	"linke/internal/shared/logger"

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

// GetUserByTelegramID retrieves a user by their Telegram ID
func (s *UserService) GetUserByTelegramID(ctx context.Context, telegramID string) (*entities.User, error) {
	var user entities.User
	if err := s.db.WithContext(ctx).Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		s.logger.Error("Failed to get user by telegram ID",
			logger.String("telegram_id", telegramID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetUsersByIDs retrieves multiple users by their IDs with optimized batch processing
func (s *UserService) GetUsersByIDs(ctx context.Context, ids []uint) ([]*entities.User, error) {
	if len(ids) == 0 {
		return []*entities.User{}, nil
	}

	// Remove duplicates and optimize ID list
	uniqueIDs := s.deduplicateIDs(ids)
	if len(uniqueIDs) == 0 {
		return []*entities.User{}, nil
	}

	// Use chunked processing for large batch queries to prevent query size limits
	const maxChunkSize = 1000 // Prevent hitting database query limits
	var allUsers []*entities.User

	for i := 0; i < len(uniqueIDs); i += maxChunkSize {
		end := i + maxChunkSize
		if end > len(uniqueIDs) {
			end = len(uniqueIDs)
		}
		
		chunk := uniqueIDs[i:end]
		var chunkUsers []*entities.User
		
		// Optimized query with minimal field selection for better performance
		if err := s.db.WithContext(ctx).
			Select("id, email, username, name, status, role, provider, created_at, updated_at").
			Where("id IN (?)", chunk).
			Find(&chunkUsers).Error; err != nil {
			s.logger.Error("Failed to get users chunk by IDs",
				logger.Int("chunk_start", i),
				logger.Int("chunk_size", len(chunk)),
				logger.Int("total_chunks", (len(uniqueIDs)+maxChunkSize-1)/maxChunkSize),
				logger.ErrorField(err),
			)
			return nil, fmt.Errorf("failed to get users by IDs (chunk %d-%d): %w", i, end-1, err)
		}
		
		allUsers = append(allUsers, chunkUsers...)
	}

	s.logger.Debug("Retrieved users by IDs with chunked processing",
		logger.Int("requested_count", len(ids)),
		logger.Int("unique_count", len(uniqueIDs)),
		logger.Int("found_count", len(allUsers)),
		logger.Int("chunks_processed", (len(uniqueIDs)+maxChunkSize-1)/maxChunkSize))

	return allUsers, nil
}

// deduplicateIDs removes duplicate IDs and maintains order for better cache locality
func (s *UserService) deduplicateIDs(ids []uint) []uint {
	if len(ids) == 0 {
		return ids
	}

	seen := make(map[uint]bool, len(ids))
	result := make([]uint, 0, len(ids))
	
	for _, id := range ids {
		if id > 0 && !seen[id] { // Skip zero IDs and duplicates
			seen[id] = true
			result = append(result, id)
		}
	}
	
	return result
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

// GetUserStats returns user statistics with optimized batch queries
func (s *UserService) GetUserStats(ctx context.Context) (*interfaces.UserStats, error) {
	stats := &interfaces.UserStats{
		ByProvider: make(map[string]int64),
	}

	// Use single query to get status counts (more efficient than separate queries)
	statusCounts := []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}{}
	
	if err := s.db.WithContext(ctx).
		Model(&entities.User{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}

	// Process status counts
	for _, sc := range statusCounts {
		switch sc.Status {
		case entities.UserStatusActive:
			stats.ActiveUsers = sc.Count
		case entities.UserStatusInactive:
			stats.InactiveUsers = sc.Count
		case entities.UserStatusBanned:
			stats.BannedUsers = sc.Count
		}
		stats.TotalUsers += sc.Count
	}

	// Use single query to get provider counts
	providerCounts := []struct {
		Provider string `json:"provider"`
		Count    int64  `json:"count"`
	}{}
	
	if err := s.db.WithContext(ctx).
		Model(&entities.User{}).
		Select("provider, COUNT(*) as count").
		Group("provider").
		Find(&providerCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get provider counts: %w", err)
	}

	// Process provider counts
	for _, pc := range providerCounts {
		stats.ByProvider[pc.Provider] = pc.Count
	}

	// Get additional stats in a single batch query with subqueries
	var additionalStats struct {
		DeletedUsers  int64 `json:"deleted_users"`
		RecentSignups int64 `json:"recent_signups"`
	}

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	
	// Use raw SQL for better performance with complex aggregations
	query := `
		SELECT 
			(SELECT COUNT(*) FROM users WHERE deleted_at IS NOT NULL) as deleted_users,
			(SELECT COUNT(*) FROM users WHERE created_at >= ? AND deleted_at IS NULL) as recent_signups
	`
	
	if err := s.db.WithContext(ctx).Raw(query, thirtyDaysAgo).Scan(&additionalStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get additional stats: %w", err)
	}

	stats.DeletedUsers = additionalStats.DeletedUsers
	stats.RecentSignups = additionalStats.RecentSignups

	s.logger.Debug("Retrieved user statistics with optimized queries",
		logger.Int64("total_users", stats.TotalUsers),
		logger.Int64("active_users", stats.ActiveUsers),
		logger.Int64("deleted_users", stats.DeletedUsers),
		logger.Int64("recent_signups", stats.RecentSignups))

	return stats, nil
}

// BatchDeleteUsers performs optimized batch soft delete on multiple users
func (s *UserService) BatchDeleteUsers(ctx context.Context, ids []uint) (*interfaces.BatchOperationResult, error) {
	if len(ids) == 0 {
		return &interfaces.BatchOperationResult{}, nil
	}

	result := &interfaces.BatchOperationResult{}

	// Remove duplicates for better performance
	uniqueIDs := s.deduplicateIDs(ids)
	
	// Use single optimized batch delete query instead of individual deletes
	deleteResult := s.db.WithContext(ctx).
		Where("id IN (?)", uniqueIDs).
		Delete(&entities.User{})
	
	if deleteResult.Error != nil {
		s.logger.Error("Failed to perform batch delete",
			logger.Int("ids_count", len(uniqueIDs)),
			logger.ErrorField(deleteResult.Error),
		)
		return nil, fmt.Errorf("failed to perform batch delete: %w", deleteResult.Error)
	}

	result.DeletedCount = int(deleteResult.RowsAffected)

	// Identify failed IDs by checking which users still exist (weren't deleted)
	if result.DeletedCount < len(uniqueIDs) {
		var stillExistingUsers []entities.User
		if err := s.db.WithContext(ctx).
			Select("id").
			Where("id IN (?)", uniqueIDs).
			Find(&stillExistingUsers).Error; err == nil {
			
			existingIDsMap := make(map[uint]bool)
			for _, user := range stillExistingUsers {
				existingIDsMap[user.ID] = true
			}

			// IDs that still exist are failed deletes
			for _, id := range uniqueIDs {
				if existingIDsMap[id] {
					result.FailedIDs = append(result.FailedIDs, id)
				}
			}
		}
	}

	s.logger.Info("Batch delete completed with optimized query",
		logger.Int("requested_count", len(ids)),
		logger.Int("unique_count", len(uniqueIDs)),
		logger.Int("deleted_count", result.DeletedCount),
		logger.Int("failed_count", len(result.FailedIDs)),
	)

	return result, nil
}

// BatchRestoreUsers performs optimized batch restore on multiple soft deleted users
func (s *UserService) BatchRestoreUsers(ctx context.Context, ids []uint) (*interfaces.BatchOperationResult, error) {
	if len(ids) == 0 {
		return &interfaces.BatchOperationResult{}, nil
	}

	result := &interfaces.BatchOperationResult{}

	// Remove duplicates for better performance
	uniqueIDs := s.deduplicateIDs(ids)

	// Use single optimized batch restore query
	restoreResult := s.db.WithContext(ctx).
		Unscoped().
		Model(&entities.User{}).
		Where("id IN (?) AND deleted_at IS NOT NULL", uniqueIDs).
		Update("deleted_at", nil)

	if restoreResult.Error != nil {
		s.logger.Error("Failed to perform batch restore",
			logger.Int("ids_count", len(uniqueIDs)),
			logger.ErrorField(restoreResult.Error),
		)
		return nil, fmt.Errorf("failed to perform batch restore: %w", restoreResult.Error)
	}

	result.RestoredCount = int(restoreResult.RowsAffected)

	// Identify failed IDs by checking which users are still deleted
	if result.RestoredCount < len(uniqueIDs) {
		var stillDeletedUsers []entities.User
		if err := s.db.WithContext(ctx).
			Unscoped().
			Select("id").
			Where("id IN (?) AND deleted_at IS NOT NULL", uniqueIDs).
			Find(&stillDeletedUsers).Error; err == nil {

			stillDeletedIDsMap := make(map[uint]bool)
			for _, user := range stillDeletedUsers {
				stillDeletedIDsMap[user.ID] = true
			}

			// IDs that are still deleted are failed restores
			for _, id := range uniqueIDs {
				if stillDeletedIDsMap[id] {
					result.FailedIDs = append(result.FailedIDs, id)
				}
			}
		}

		// Also check for IDs that don't exist at all
		var allExistingUsers []entities.User
		if err := s.db.WithContext(ctx).
			Unscoped().
			Select("id").
			Where("id IN (?)", uniqueIDs).
			Find(&allExistingUsers).Error; err == nil {

			existingIDsMap := make(map[uint]bool)
			for _, user := range allExistingUsers {
				existingIDsMap[user.ID] = true
			}

			// IDs that don't exist at all are also failed restores
			for _, id := range uniqueIDs {
				if !existingIDsMap[id] {
					// Only add if not already in failed list
					found := false
					for _, failedID := range result.FailedIDs {
						if failedID == id {
							found = true
							break
						}
					}
					if !found {
						result.FailedIDs = append(result.FailedIDs, id)
					}
				}
			}
		}
	}

	s.logger.Info("Batch restore completed with optimized query",
		logger.Int("requested_count", len(ids)),
		logger.Int("unique_count", len(uniqueIDs)),
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
