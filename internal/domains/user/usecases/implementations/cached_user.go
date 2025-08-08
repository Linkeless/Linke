package implementations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"linke/internal/domains/user/entities"
	"linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/cache"
	"linke/internal/shared/framework"
	"linke/internal/shared/logger"
)

// CachedUserService wraps UserService with comprehensive caching
type CachedUserService struct {
	baseService  *UserService
	cacheManager cache.CacheManager
	cacheKeys    *cache.AllCacheKeys
	logger       framework.Logger
}

// NewCachedUserService creates a new cached user service
func NewCachedUserService(
	baseService *UserService,
	cacheManager cache.CacheManager,
	cacheKeys *cache.AllCacheKeys,
	logger framework.Logger,
) *CachedUserService {
	return &CachedUserService{
		baseService:  baseService,
		cacheManager: cacheManager,
		cacheKeys:    cacheKeys,
		logger:       logger,
	}
}

// getUserFromCache attempts to get user from cache, returns nil if not found
func (s *CachedUserService) getUserFromCache(ctx context.Context, key string) (*entities.User, error) {
	data, err := s.cacheManager.GetCache().Get(ctx, key)
	if err != nil {
		s.logger.Debug("Cache get failed", logger.String("key", key), logger.ErrorField(err))
		return nil, nil // Cache miss or error, proceed to database
	}

	if data == nil {
		return nil, nil // Cache miss
	}

	var user entities.User
	if err := json.Unmarshal(data, &user); err != nil {
		s.logger.Warn("Failed to unmarshal cached user",
			logger.String("key", key),
			logger.ErrorField(err))
		// Delete corrupted cache entry
		s.cacheManager.GetCache().Delete(ctx, key)
		return nil, nil
	}

	return &user, nil
}

// setUserInCache stores user in cache with appropriate TTL
func (s *CachedUserService) setUserInCache(ctx context.Context, key string, user *entities.User, ttl time.Duration) {
	if user == nil {
		return
	}

	data, err := json.Marshal(user)
	if err != nil {
		s.logger.Error("Failed to marshal user for cache",
			logger.String("key", key),
			logger.ErrorField(err))
		return
	}

	if err := s.cacheManager.GetCache().Set(ctx, key, data, ttl); err != nil {
		s.logger.Error("Failed to set user in cache",
			logger.String("key", key),
			logger.ErrorField(err))
	}
}

// invalidateUserCache invalidates all cache entries for a user
func (s *CachedUserService) invalidateUserCache(ctx context.Context, user *entities.User) {
	if user == nil {
		return
	}

	// Invalidate cache by ID
	keyByID := s.cacheKeys.User.UserByID(user.ID)
	s.cacheManager.GetCache().Delete(ctx, keyByID)

	// Invalidate cache by email
	if user.Email != "" {
		keyByEmail := s.cacheKeys.User.UserByEmail(user.Email)
		s.cacheManager.GetCache().Delete(ctx, keyByEmail)
	}

	// Invalidate cache by username
	if user.Username != "" {
		keyByUsername := s.cacheKeys.User.UserByUsername(user.Username)
		s.cacheManager.GetCache().Delete(ctx, keyByUsername)
	}

	// Invalidate user profile cache
	profileKey := s.cacheKeys.User.UserProfile(user.ID)
	s.cacheManager.GetCache().Delete(ctx, profileKey)

	// Invalidate all user-related caches using pattern
	pattern := s.cacheKeys.User.UserPattern(user.ID)
	s.cacheManager.GetCache().DeleteByPattern(ctx, pattern)

	s.logger.Debug("User cache invalidated", logger.Uint("user_id", user.ID))
}

// invalidateUserCacheByID invalidates cache by user ID only
func (s *CachedUserService) invalidateUserCacheByID(ctx context.Context, userID uint) {
	// We need to fetch the user first to get email and username for cache invalidation
	user, err := s.baseService.GetUserByID(ctx, userID)
	if err == nil && user != nil {
		s.invalidateUserCache(ctx, user)
	} else {
		// Fallback: invalidate by pattern only
		pattern := s.cacheKeys.User.UserPattern(userID)
		s.cacheManager.GetCache().DeleteByPattern(ctx, pattern)
		s.logger.Debug("User cache invalidated by pattern only", logger.Uint("user_id", userID))
	}
}

// CreateUser creates a new user with write-through caching
func (s *CachedUserService) CreateUser(ctx context.Context, user *entities.User) error {
	// Create user in database first
	if err := s.baseService.CreateUser(ctx, user); err != nil {
		return err
	}

	// Cache the newly created user
	ttl := cache.MediumCacheTTL // 15 minutes

	// Cache by ID
	keyByID := s.cacheKeys.User.UserByID(user.ID)
	s.setUserInCache(ctx, keyByID, user, ttl)

	// Cache by email
	if user.Email != "" {
		keyByEmail := s.cacheKeys.User.UserByEmail(user.Email)
		s.setUserInCache(ctx, keyByEmail, user, ttl)
	}

	// Cache by username
	if user.Username != "" {
		keyByUsername := s.cacheKeys.User.UserByUsername(user.Username)
		s.setUserInCache(ctx, keyByUsername, user, ttl)
	}

	s.logger.Info("User created and cached",
		logger.Uint("user_id", user.ID),
		logger.String("email", user.Email))

	return nil
}

// GetUserByID retrieves a user by ID with cache-aside pattern
func (s *CachedUserService) GetUserByID(ctx context.Context, id uint) (*entities.User, error) {
	cacheKey := s.cacheKeys.User.UserByID(id)

	// Try cache first
	if user, err := s.getUserFromCache(ctx, cacheKey); err == nil && user != nil {
		return user, nil
	}

	// Cache miss, fetch from database
	user, err := s.baseService.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	s.setUserInCache(ctx, cacheKey, user, cache.MediumCacheTTL)

	return user, nil
}

// GetUserByEmail retrieves a user by email with cache-aside pattern
func (s *CachedUserService) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	cacheKey := s.cacheKeys.User.UserByEmail(email)

	// Try cache first
	if user, err := s.getUserFromCache(ctx, cacheKey); err == nil && user != nil {
		return user, nil
	}

	// Cache miss, fetch from database
	user, err := s.baseService.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Cache the result
	s.setUserInCache(ctx, cacheKey, user, cache.MediumCacheTTL)

	return user, nil
}

// GetUserByTelegramID retrieves a user by their Telegram ID with caching
func (s *CachedUserService) GetUserByTelegramID(ctx context.Context, telegramID string) (*entities.User, error) {
	cacheKey := cache.CachePrefixUser + "telegram:" + telegramID

	// Try cache first
	if user, err := s.getUserFromCache(ctx, cacheKey); err == nil && user != nil {
		return user, nil
	}

	// Cache miss, fetch from database
	user, err := s.baseService.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	s.setUserInCache(ctx, cacheKey, user, cache.MediumCacheTTL)

	return user, nil
}

// GetActiveUserByID retrieves an active user by ID with cache-aside pattern
func (s *CachedUserService) GetActiveUserByID(ctx context.Context, id uint) (*entities.User, error) {
	// For active users, we can still use the regular cache but verify active status
	user, err := s.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !user.IsActive() {
		return nil, fmt.Errorf("active user not found")
	}

	return user, nil
}

// GetActiveUserByEmail retrieves an active user by email with cache-aside pattern
func (s *CachedUserService) GetActiveUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	// For active users, we can still use the regular cache but verify active status
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if !user.IsActive() {
		return nil, fmt.Errorf("active user not found")
	}

	return user, nil
}

// GetUsersByIDs retrieves multiple users by IDs with caching optimization
func (s *CachedUserService) GetUsersByIDs(ctx context.Context, ids []uint) ([]*entities.User, error) {
	if len(ids) == 0 {
		return []*entities.User{}, nil
	}

	var cachedUsers []*entities.User
	var missedIDs []uint
	
	// Try to get users from cache first
	for _, id := range ids {
		cacheKey := s.cacheKeys.User.UserByID(id)
		if cachedUser, err := s.getUserFromCache(ctx, cacheKey); err == nil && cachedUser != nil {
			cachedUsers = append(cachedUsers, cachedUser)
		} else {
			missedIDs = append(missedIDs, id)
		}
	}

	// Fetch missed users from database in batch
	var missedUsers []*entities.User
	if len(missedIDs) > 0 {
		users, err := s.baseService.GetUsersByIDs(ctx, missedIDs)
		if err != nil {
			return nil, err
		}
		missedUsers = users
		
		// Cache the fetched users
		for _, user := range missedUsers {
			cacheKey := s.cacheKeys.User.UserByID(user.ID)
			s.setUserInCache(ctx, cacheKey, user, cache.MediumCacheTTL)
		}
	}

	// Combine cached and fetched users
	allUsers := make([]*entities.User, 0, len(cachedUsers)+len(missedUsers))
	allUsers = append(allUsers, cachedUsers...)
	allUsers = append(allUsers, missedUsers...)

	s.logger.Debug("Retrieved users by IDs with cache optimization",
		logger.Int("total_requested", len(ids)),
		logger.Int("cache_hits", len(cachedUsers)),
		logger.Int("cache_misses", len(missedIDs)),
		logger.Int("total_found", len(allUsers)))

	return allUsers, nil
}

// UpdateUser updates a user with write-through caching
func (s *CachedUserService) UpdateUser(ctx context.Context, user *entities.User) error {
	// Get the old user data for cache invalidation
	oldUser, _ := s.baseService.GetUserByID(ctx, user.ID)

	// Update in database first
	if err := s.baseService.UpdateUser(ctx, user); err != nil {
		return err
	}

	// Invalidate old cache entries (email/username might have changed)
	if oldUser != nil {
		s.invalidateUserCache(ctx, oldUser)
	}

	// Cache the updated user
	ttl := cache.MediumCacheTTL

	// Cache by ID
	keyByID := s.cacheKeys.User.UserByID(user.ID)
	s.setUserInCache(ctx, keyByID, user, ttl)

	// Cache by email
	if user.Email != "" {
		keyByEmail := s.cacheKeys.User.UserByEmail(user.Email)
		s.setUserInCache(ctx, keyByEmail, user, ttl)
	}

	// Cache by username
	if user.Username != "" {
		keyByUsername := s.cacheKeys.User.UserByUsername(user.Username)
		s.setUserInCache(ctx, keyByUsername, user, ttl)
	}

	s.logger.Info("User updated and cache refreshed",
		logger.Uint("user_id", user.ID))

	return nil
}

// SoftDeleteUser performs soft delete with cache invalidation
func (s *CachedUserService) SoftDeleteUser(ctx context.Context, id uint) error {
	// Get user data before deletion for cache invalidation
	user, _ := s.baseService.GetUserByID(ctx, id)

	// Perform soft delete
	if err := s.baseService.SoftDeleteUser(ctx, id); err != nil {
		return err
	}

	// Invalidate all cache entries for this user
	if user != nil {
		s.invalidateUserCache(ctx, user)
	} else {
		s.invalidateUserCacheByID(ctx, id)
	}

	s.logger.Info("User soft deleted and cache invalidated",
		logger.Uint("user_id", id))

	return nil
}

// RestoreUser restores a soft deleted user with cache invalidation
func (s *CachedUserService) RestoreUser(ctx context.Context, id uint) error {
	// Perform restore
	if err := s.baseService.RestoreUser(ctx, id); err != nil {
		return err
	}

	// Invalidate cache to ensure fresh data on next access
	s.invalidateUserCacheByID(ctx, id)

	s.logger.Info("User restored and cache invalidated",
		logger.Uint("user_id", id))

	return nil
}

// HardDeleteUser permanently deletes a user with cache invalidation
func (s *CachedUserService) HardDeleteUser(ctx context.Context, id uint) error {
	// Get user data before deletion for cache invalidation
	user, _ := s.baseService.GetUserByID(ctx, id)

	// Perform hard delete
	if err := s.baseService.HardDeleteUser(ctx, id); err != nil {
		return err
	}

	// Invalidate all cache entries for this user
	if user != nil {
		s.invalidateUserCache(ctx, user)
	} else {
		s.invalidateUserCacheByID(ctx, id)
	}

	s.logger.Warn("User hard deleted and cache invalidated",
		logger.Uint("user_id", id))

	return nil
}

// UpdateUserStatus updates a user's status with cache invalidation
func (s *CachedUserService) UpdateUserStatus(ctx context.Context, id uint, status string) (*entities.User, error) {
	// Update status in database
	user, err := s.baseService.UpdateUserStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}

	// Invalidate and refresh cache
	s.invalidateUserCache(ctx, user)

	// Cache the updated user
	ttl := cache.MediumCacheTTL
	keyByID := s.cacheKeys.User.UserByID(user.ID)
	s.setUserInCache(ctx, keyByID, user, ttl)

	if user.Email != "" {
		keyByEmail := s.cacheKeys.User.UserByEmail(user.Email)
		s.setUserInCache(ctx, keyByEmail, user, ttl)
	}

	if user.Username != "" {
		keyByUsername := s.cacheKeys.User.UserByUsername(user.Username)
		s.setUserInCache(ctx, keyByUsername, user, ttl)
	}

	s.logger.Info("User status updated and cache refreshed",
		logger.Uint("user_id", id),
		logger.String("new_status", status))

	return user, nil
}

// UpdateUserRole updates a user's role with cache invalidation
func (s *CachedUserService) UpdateUserRole(ctx context.Context, id uint, role string) (*entities.User, error) {
	// Update role in database
	user, err := s.baseService.UpdateUserRole(ctx, id, role)
	if err != nil {
		return nil, err
	}

	// Invalidate and refresh cache
	s.invalidateUserCache(ctx, user)

	// Cache the updated user
	ttl := cache.MediumCacheTTL
	keyByID := s.cacheKeys.User.UserByID(user.ID)
	s.setUserInCache(ctx, keyByID, user, ttl)

	if user.Email != "" {
		keyByEmail := s.cacheKeys.User.UserByEmail(user.Email)
		s.setUserInCache(ctx, keyByEmail, user, ttl)
	}

	if user.Username != "" {
		keyByUsername := s.cacheKeys.User.UserByUsername(user.Username)
		s.setUserInCache(ctx, keyByUsername, user, ttl)
	}

	s.logger.Info("User role updated and cache refreshed",
		logger.Uint("user_id", id),
		logger.String("new_role", role))

	return user, nil
}

// BatchDeleteUsers performs batch soft delete with cache invalidation
func (s *CachedUserService) BatchDeleteUsers(ctx context.Context, ids []uint) (*interfaces.BatchOperationResult, error) {
	// Get user data before deletion for cache invalidation using batch query
	var usersToInvalidate []*entities.User
	
	// Use batch query instead of N individual queries
	users, err := s.baseService.GetUsersByIDs(ctx, ids)
	if err != nil {
		s.logger.Error("Failed to batch fetch users for deletion", 
			logger.Error2("error", err),
			logger.Int("ids_count", len(ids)))
		// Continue with deletion even if we can't fetch users for cache invalidation
	} else {
		usersToInvalidate = users
	}

	// Perform batch delete
	result, err := s.baseService.BatchDeleteUsers(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Invalidate cache for all affected users
	for _, user := range usersToInvalidate {
		s.invalidateUserCache(ctx, user)
	}

	// Also invalidate by pattern for any users we couldn't fetch or all IDs as fallback
	for _, id := range ids {
		pattern := s.cacheKeys.User.UserPattern(id)
		s.cacheManager.GetCache().DeleteByPattern(ctx, pattern)
	}

	s.logger.Info("Batch delete completed and caches invalidated",
		logger.Int("deleted_count", result.DeletedCount),
		logger.Int("failed_count", len(result.FailedIDs)))

	return result, nil
}

// BatchRestoreUsers performs batch restore with cache invalidation
func (s *CachedUserService) BatchRestoreUsers(ctx context.Context, ids []uint) (*interfaces.BatchOperationResult, error) {
	// Perform batch restore
	result, err := s.baseService.BatchRestoreUsers(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Invalidate cache for all affected users to ensure fresh data
	for _, id := range ids {
		pattern := s.cacheKeys.User.UserPattern(id)
		s.cacheManager.GetCache().DeleteByPattern(ctx, pattern)
	}

	s.logger.Info("Batch restore completed and caches invalidated",
		logger.Int("restored_count", result.RestoredCount),
		logger.Int("failed_count", len(result.FailedIDs)))

	return result, nil
}

// Delegate non-cached methods to base service

// ListUsers delegates to base service (lists are not cached due to complexity)
func (s *CachedUserService) ListUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	return s.baseService.ListUsers(ctx, limit, offset)
}

// ListDeletedUsers delegates to base service
func (s *CachedUserService) ListDeletedUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	return s.baseService.ListDeletedUsers(ctx, limit, offset)
}

// ListUsersByProvider delegates to base service
func (s *CachedUserService) ListUsersByProvider(ctx context.Context, provider string, limit, offset int) ([]*entities.User, int64, error) {
	return s.baseService.ListUsersByProvider(ctx, provider, limit, offset)
}

// SearchUsers delegates to base service
func (s *CachedUserService) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, int64, error) {
	return s.baseService.SearchUsers(ctx, query, limit, offset)
}

// GetUserStats delegates to base service (statistics are not cached due to frequent changes)
func (s *CachedUserService) GetUserStats(ctx context.Context) (*interfaces.UserStats, error) {
	return s.baseService.GetUserStats(ctx)
}
