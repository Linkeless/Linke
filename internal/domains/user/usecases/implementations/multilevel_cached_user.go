package implementations

import (
	"context"
	"encoding/json"
	"fmt"

	"linke/internal/domains/user/entities"
	"linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/cache"
	"linke/internal/shared/framework"
	"linke/internal/shared/logger"
)

// MultiLevelCachedUserService wraps UserService with multi-level caching
type MultiLevelCachedUserService struct {
	baseService  *UserService
	cacheManager cache.MultiLevelCacheManager
	cacheKeys    *cache.AllCacheKeys
	logger       framework.Logger

	// Cache configuration per operation type
	configs *UserCacheConfigs
}

// UserCacheConfigs defines cache configurations for different user operations
type UserCacheConfigs struct {
	// User lookup operations (highly cacheable)
	UserByID       *cache.CacheOptions
	UserByEmail    *cache.CacheOptions
	UserByUsername *cache.CacheOptions

	// Profile operations (medium cacheable)
	UserProfile *cache.CacheOptions

	// Dynamic operations (low cacheable)
	UserStats  *cache.CacheOptions
	UserSearch *cache.CacheOptions
}

// NewMultiLevelCachedUserService creates a new multi-level cached user service
func NewMultiLevelCachedUserService(
	baseService *UserService,
	cacheManager cache.MultiLevelCacheManager,
	cacheKeys *cache.AllCacheKeys,
	logger framework.Logger,
) *MultiLevelCachedUserService {
	configs := &UserCacheConfigs{
		// User lookups - cache aggressively in both L1 and L2
		UserByID: &cache.CacheOptions{
			TTL:                  cache.LongCacheTTL,
			Tags:                 []string{cache.CacheTagUser},
			SkipCompression:      false,
			RefreshOnMiss:        true,
			StaleWhileRevalidate: true,
		},
		UserByEmail: &cache.CacheOptions{
			TTL:                  cache.LongCacheTTL,
			Tags:                 []string{cache.CacheTagUser},
			SkipCompression:      false,
			RefreshOnMiss:        true,
			StaleWhileRevalidate: true,
		},
		UserByUsername: &cache.CacheOptions{
			TTL:                  cache.LongCacheTTL,
			Tags:                 []string{cache.CacheTagUser},
			SkipCompression:      false,
			RefreshOnMiss:        true,
			StaleWhileRevalidate: true,
		},

		// Profile data - medium caching
		UserProfile: &cache.CacheOptions{
			TTL:                  cache.MediumCacheTTL,
			Tags:                 []string{cache.CacheTagUser},
			SkipCompression:      false,
			RefreshOnMiss:        false,
			StaleWhileRevalidate: false,
		},

		// Dynamic data - minimal caching
		UserStats: &cache.CacheOptions{
			TTL:                  cache.ShortCacheTTL,
			Tags:                 []string{cache.CacheTagUser},
			SkipCompression:      true,
			RefreshOnMiss:        false,
			StaleWhileRevalidate: false,
		},
		UserSearch: &cache.CacheOptions{
			TTL:                  cache.ShortCacheTTL,
			Tags:                 []string{cache.CacheTagUser},
			SkipCompression:      true,
			RefreshOnMiss:        false,
			StaleWhileRevalidate: false,
		},
	}

	return &MultiLevelCachedUserService{
		baseService:  baseService,
		cacheManager: cacheManager,
		cacheKeys:    cacheKeys,
		logger:       logger,
		configs:      configs,
	}
}

// CreateUser creates a new user with intelligent caching strategy
func (s *MultiLevelCachedUserService) CreateUser(ctx context.Context, user *entities.User) error {
	// Create user in database first
	if err := s.baseService.CreateUser(ctx, user); err != nil {
		return err
	}

	// Use write-through strategy for new user data
	s.cacheNewUser(ctx, user)

	s.logger.Info("User created with multi-level caching",
		logger.Uint("user_id", user.ID),
		logger.String("email", user.Email))

	return nil
}

// GetUserByID retrieves a user by ID with optimized multi-level caching
func (s *MultiLevelCachedUserService) GetUserByID(ctx context.Context, id uint) (*entities.User, error) {
	cacheKey := s.cacheKeys.User.UserByID(id)

	// Try multi-level cache first
	if user, err := s.getUserFromCache(ctx, cacheKey); err == nil && user != nil {
		s.logger.Debug("User cache hit", logger.Uint("user_id", id), logger.String("key", cacheKey))
		return user, nil
	}

	// Cache miss - fetch from database
	user, err := s.baseService.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result with user-specific TTL
	if user != nil {
		s.cacheUser(ctx, user, s.configs.UserByID)
	}

	s.logger.Debug("User loaded from database and cached",
		logger.Uint("user_id", id))

	return user, nil
}

// GetUserByEmail retrieves a user by email with smart caching
func (s *MultiLevelCachedUserService) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	cacheKey := s.cacheKeys.User.UserByEmail(email)

	// Try cache first
	if user, err := s.getUserFromCache(ctx, cacheKey); err == nil && user != nil {
		return user, nil
	}

	// Cache miss - fetch from database
	user, err := s.baseService.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if user != nil {
		s.cacheUser(ctx, user, s.configs.UserByEmail)
		// Also cache by ID for better hit rate
		s.cacheUserByID(ctx, user, s.configs.UserByID)
	}

	return user, nil
}

// GetActiveUserByID with enhanced caching for frequently accessed active users
func (s *MultiLevelCachedUserService) GetActiveUserByID(ctx context.Context, id uint) (*entities.User, error) {
	// Use the same caching strategy but with additional validation
	user, err := s.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !user.IsActive() {
		// Remove from cache if user is no longer active
		s.invalidateUserCaches(ctx, user)
		return nil, fmt.Errorf("active user not found")
	}

	// Promote frequently accessed active users to L1 cache
	s.promoteToL1IfNeeded(ctx, user)

	return user, nil
}

// UpdateUser with intelligent cache invalidation and warming
func (s *MultiLevelCachedUserService) UpdateUser(ctx context.Context, user *entities.User) error {
	// Get the old user data for cache invalidation
	oldUser, _ := s.baseService.GetUserByID(ctx, user.ID)

	// Update in database first
	if err := s.baseService.UpdateUser(ctx, user); err != nil {
		return err
	}

	// Smart cache invalidation
	s.smartInvalidateAndUpdate(ctx, oldUser, user)

	s.logger.Info("User updated with smart cache management",
		logger.Uint("user_id", user.ID))

	return nil
}

// BatchDeleteUsers with efficient cache invalidation
func (s *MultiLevelCachedUserService) BatchDeleteUsers(ctx context.Context, ids []uint) (*interfaces.BatchOperationResult, error) {
	// Pre-fetch users for cache invalidation
	usersToInvalidate := make([]*entities.User, 0, len(ids))
	for _, id := range ids {
		if user, err := s.baseService.GetUserByID(ctx, id); err == nil && user != nil {
			usersToInvalidate = append(usersToInvalidate, user)
		}
	}

	// Perform batch delete
	result, err := s.baseService.BatchDeleteUsers(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Efficient batch cache invalidation
	s.batchInvalidateUsers(ctx, usersToInvalidate)

	s.logger.Info("Batch delete completed with cache invalidation",
		logger.Int("deleted_count", result.DeletedCount),
		logger.Int("failed_count", len(result.FailedIDs)))

	return result, nil
}

// Private helper methods

func (s *MultiLevelCachedUserService) getUserFromCache(ctx context.Context, key string) (*entities.User, error) {
	data, err := s.cacheManager.GetCache().Get(ctx, key)
	if err != nil {
		s.logger.Debug("Cache get failed", logger.String("key", key), logger.ErrorField(err))
		return nil, nil
	}

	if data == nil {
		return nil, nil
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

func (s *MultiLevelCachedUserService) cacheUser(ctx context.Context, user *entities.User, config *cache.CacheOptions) {
	if user == nil {
		return
	}

	data, err := json.Marshal(user)
	if err != nil {
		s.logger.Error("Failed to marshal user for cache",
			logger.Uint("user_id", user.ID),
			logger.ErrorField(err))
		return
	}

	// Cache by ID
	keyByID := s.cacheKeys.User.UserByID(user.ID)
	if err := s.cacheManager.GetCache().Set(ctx, keyByID, data, config.TTL); err != nil {
		s.logger.Error("Failed to cache user by ID",
			logger.String("key", keyByID),
			logger.ErrorField(err))
	}

	// Cache by email if present
	if user.Email != "" {
		keyByEmail := s.cacheKeys.User.UserByEmail(user.Email)
		if err := s.cacheManager.GetCache().Set(ctx, keyByEmail, data, config.TTL); err != nil {
			s.logger.Error("Failed to cache user by email",
				logger.String("key", keyByEmail),
				logger.ErrorField(err))
		}
	}

	// Cache by username if present
	if user.Username != "" {
		keyByUsername := s.cacheKeys.User.UserByUsername(user.Username)
		if err := s.cacheManager.GetCache().Set(ctx, keyByUsername, data, config.TTL); err != nil {
			s.logger.Error("Failed to cache user by username",
				logger.String("key", keyByUsername),
				logger.ErrorField(err))
		}
	}
}

func (s *MultiLevelCachedUserService) cacheNewUser(ctx context.Context, user *entities.User) {
	// For new users, cache immediately in both L1 and L2 with longer TTL
	// This uses write-through strategy
	s.cacheUser(ctx, user, s.configs.UserByID)
}

func (s *MultiLevelCachedUserService) cacheUserByID(ctx context.Context, user *entities.User, config *cache.CacheOptions) {
	if user == nil {
		return
	}

	data, err := json.Marshal(user)
	if err != nil {
		return
	}

	keyByID := s.cacheKeys.User.UserByID(user.ID)
	s.cacheManager.GetCache().Set(ctx, keyByID, data, config.TTL)
}

func (s *MultiLevelCachedUserService) invalidateUserCaches(ctx context.Context, user *entities.User) {
	if user == nil {
		return
	}

	// Invalidate all cache entries for this user
	patterns := []string{
		s.cacheKeys.User.UserByID(user.ID),
		s.cacheKeys.User.UserPattern(user.ID),
	}

	if user.Email != "" {
		patterns = append(patterns, s.cacheKeys.User.UserByEmail(user.Email))
	}

	if user.Username != "" {
		patterns = append(patterns, s.cacheKeys.User.UserByUsername(user.Username))
	}

	for _, pattern := range patterns {
		s.cacheManager.GetCache().Delete(ctx, pattern)
	}

	s.logger.Debug("User caches invalidated", logger.Uint("user_id", user.ID))
}

func (s *MultiLevelCachedUserService) smartInvalidateAndUpdate(ctx context.Context, oldUser, newUser *entities.User) {
	// Invalidate old cache entries first
	if oldUser != nil {
		s.invalidateUserCaches(ctx, oldUser)
	}

	// Cache the updated user with appropriate strategy
	// If it's a frequently accessed user (determined by some criteria), cache in L1
	if s.isFrequentlyAccessedUser(newUser) {
		s.cacheUser(ctx, newUser, s.configs.UserByID)
		s.promoteToL1IfNeeded(ctx, newUser)
	} else {
		// Cache only in L2 for less frequently accessed users
		s.cacheUser(ctx, newUser, s.configs.UserProfile)
	}
}

func (s *MultiLevelCachedUserService) batchInvalidateUsers(ctx context.Context, users []*entities.User) {
	// Efficiently invalidate multiple users
	patterns := make([]string, 0, len(users)*3)

	for _, user := range users {
		if user != nil {
			patterns = append(patterns, s.cacheKeys.User.UserPattern(user.ID))
		}
	}

	// Batch invalidate
	for _, pattern := range patterns {
		s.cacheManager.GetCache().DeleteByPattern(ctx, pattern)
	}

	s.logger.Debug("Batch invalidated user caches", logger.Int("user_count", len(users)))
}

func (s *MultiLevelCachedUserService) promoteToL1IfNeeded(ctx context.Context, user *entities.User) {
	// Logic to promote frequently accessed users to L1 cache
	// This could be based on access frequency, user role, etc.

	if user.Role == "admin" || user.Role == "premium" {
		// Promote important users to L1 cache
		if l1Cache := s.cacheManager.GetL1Cache(); l1Cache != nil {
			data, _ := json.Marshal(user)
			keyByID := s.cacheKeys.User.UserByID(user.ID)
			l1Cache.Set(ctx, keyByID, data, cache.LongCacheTTL)
		}
	}
}

func (s *MultiLevelCachedUserService) isFrequentlyAccessedUser(user *entities.User) bool {
	// Determine if user should be cached more aggressively
	// This could be based on:
	// - User role (admin, premium users)
	// - Recent activity
	// - Access patterns

	return user.Role == "admin" ||
		user.Role == "premium" ||
		user.Status == "active"
}

// Delegate non-cached methods to base service
func (s *MultiLevelCachedUserService) ListUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	return s.baseService.ListUsers(ctx, limit, offset)
}

func (s *MultiLevelCachedUserService) ListDeletedUsers(ctx context.Context, limit, offset int) ([]*entities.User, int64, error) {
	return s.baseService.ListDeletedUsers(ctx, limit, offset)
}

func (s *MultiLevelCachedUserService) ListUsersByProvider(ctx context.Context, provider string, limit, offset int) ([]*entities.User, int64, error) {
	return s.baseService.ListUsersByProvider(ctx, provider, limit, offset)
}

func (s *MultiLevelCachedUserService) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, int64, error) {
	// Search results are cached briefly
	return s.baseService.SearchUsers(ctx, query, limit, offset)
}

func (s *MultiLevelCachedUserService) GetUserStats(ctx context.Context) (*interfaces.UserStats, error) {
	// Statistics are cached for a short time only
	return s.baseService.GetUserStats(ctx)
}

func (s *MultiLevelCachedUserService) SoftDeleteUser(ctx context.Context, id uint) error {
	// Get user data before deletion for cache invalidation
	user, _ := s.baseService.GetUserByID(ctx, id)

	// Perform soft delete
	if err := s.baseService.SoftDeleteUser(ctx, id); err != nil {
		return err
	}

	// Invalidate all cache entries for this user
	if user != nil {
		s.invalidateUserCaches(ctx, user)
	}

	return nil
}

func (s *MultiLevelCachedUserService) RestoreUser(ctx context.Context, id uint) error {
	// Perform restore
	if err := s.baseService.RestoreUser(ctx, id); err != nil {
		return err
	}

	// Invalidate cache to ensure fresh data on next access
	s.cacheManager.InvalidateCache(ctx, s.cacheKeys.User.UserPattern(id))

	return nil
}

func (s *MultiLevelCachedUserService) HardDeleteUser(ctx context.Context, id uint) error {
	// Get user data before deletion for cache invalidation
	user, _ := s.baseService.GetUserByID(ctx, id)

	// Perform hard delete
	if err := s.baseService.HardDeleteUser(ctx, id); err != nil {
		return err
	}

	// Invalidate all cache entries for this user
	if user != nil {
		s.invalidateUserCaches(ctx, user)
	}

	return nil
}

func (s *MultiLevelCachedUserService) UpdateUserStatus(ctx context.Context, id uint, status string) (*entities.User, error) {
	// Update status in database
	user, err := s.baseService.UpdateUserStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}

	// Smart cache management based on new status
	if status == "active" {
		// Promote active users to better caching
		s.cacheUser(ctx, user, s.configs.UserByID)
		s.promoteToL1IfNeeded(ctx, user)
	} else {
		// Invalidate cache for inactive users
		s.invalidateUserCaches(ctx, user)
	}

	return user, nil
}

func (s *MultiLevelCachedUserService) UpdateUserRole(ctx context.Context, id uint, role string) (*entities.User, error) {
	// Update role in database
	user, err := s.baseService.UpdateUserRole(ctx, id, role)
	if err != nil {
		return nil, err
	}

	// Invalidate and refresh cache
	s.invalidateUserCaches(ctx, user)
	s.cacheUser(ctx, user, s.configs.UserByID)

	// Promote important roles to L1
	if role == "admin" || role == "premium" {
		s.promoteToL1IfNeeded(ctx, user)
	}

	return user, nil
}

func (s *MultiLevelCachedUserService) BatchRestoreUsers(ctx context.Context, ids []uint) (*interfaces.BatchOperationResult, error) {
	// Perform batch restore
	result, err := s.baseService.BatchRestoreUsers(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Invalidate cache for all affected users to ensure fresh data
	for _, id := range ids {
		s.cacheManager.InvalidateCache(ctx, s.cacheKeys.User.UserPattern(id))
	}

	return result, nil
}
