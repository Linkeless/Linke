package implementations

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/cache"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

type CachedUserSubscriptionService struct {
	*UserSubscriptionService
	cacheManager      cache.CacheManager
	cacheKeys         *cache.AllCacheKeys
	subscriptionCache *cache.CacheAside[entities.UserSubscription]
}

func NewUserSubscriptionServiceWithCache(
	db *gorm.DB,
	subscriptionPlanService interfaces.SubscriptionPlanService,
	cacheManager cache.CacheManager,
	cacheKeys *cache.AllCacheKeys,
) interfaces.UserSubscriptionService {
	baseService := NewUserSubscriptionService(db, subscriptionPlanService)

	subscriptionCache := cache.NewCacheAside[entities.UserSubscription](
		cacheManager.GetCache(),
		cache.CachePrefixSubscription,
		func(sub entities.UserSubscription) string {
			return fmt.Sprintf("id:%d", sub.ID)
		},
		cache.MediumCacheTTL, // 15 minutes TTL for subscriptions
	)

	return &CachedUserSubscriptionService{
		UserSubscriptionService: baseService,
		cacheManager:            cacheManager,
		cacheKeys:               cacheKeys,
		subscriptionCache:       subscriptionCache,
	}
}

// CreateUserSubscription creates a new user subscription with cache management
func (s *CachedUserSubscriptionService) CreateUserSubscription(ctx context.Context, req *interfaces.CreateSubscriptionRequest) (*entities.UserSubscription, error) {
	subscription, err := s.UserSubscriptionService.CreateUserSubscription(ctx, req)
	if err != nil {
		return nil, err
	}

	// Write-through: cache the new subscription
	if subscription != nil {
		if err := s.subscriptionCache.Set(ctx, subscription); err != nil {
			logger.Error("Failed to cache new subscription",
				logger.Uint("subscription_id", subscription.ID),
				logger.ErrorField(err))
		}

		// Invalidate user-related caches
		s.invalidateUserCaches(ctx, subscription.UserID)
	}

	return subscription, nil
}

// GetUserSubscription gets a user subscription by ID with caching
func (s *CachedUserSubscriptionService) GetUserSubscription(ctx context.Context, subscriptionID uint) (*entities.UserSubscription, error) {
	cacheKey := fmt.Sprintf("id:%d", subscriptionID)

	subscription, err := s.subscriptionCache.Get(ctx, cacheKey, func() (*entities.UserSubscription, error) {
		return s.UserSubscriptionService.GetUserSubscription(ctx, subscriptionID)
	})

	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// GetUserSubscriptionWithRelations gets a user subscription with related data (caching limited due to relations)
func (s *CachedUserSubscriptionService) GetUserSubscriptionWithRelations(ctx context.Context, subscriptionID uint) (*entities.UserSubscription, error) {
	// For relations, we use shorter cache to avoid stale data
	cacheKey := fmt.Sprintf("relations:id:%d", subscriptionID)

	subscription, err := s.subscriptionCache.Get(ctx, cacheKey, func() (*entities.UserSubscription, error) {
		return s.UserSubscriptionService.GetUserSubscriptionWithRelations(ctx, subscriptionID)
	})

	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// GetUserSubscriptions gets user subscriptions with caching for list results
func (s *CachedUserSubscriptionService) GetUserSubscriptions(ctx context.Context, req *interfaces.GetUserSubscriptionsRequest) ([]*entities.UserSubscription, int64, error) {
	cacheKey := s.buildSubscriptionListCacheKey(req)

	// Determine cache TTL based on query type
	cacheTTL := s.determineCacheTTL(req)

	// Skip cache for admin queries with no filters (likely admin dashboard)
	if s.shouldSkipCache(req) {
		return s.UserSubscriptionService.GetUserSubscriptions(ctx, req)
	}

	// Use cache decorator for list results
	cached, err := s.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var result struct {
			Subscriptions []*entities.UserSubscription `json:"subscriptions"`
			Total         int64                        `json:"total"`
		}
		if err := json.Unmarshal(cached, &result); err == nil {
			return result.Subscriptions, result.Total, nil
		}
	}

	// Cache miss - fetch from database
	subscriptions, total, err := s.UserSubscriptionService.GetUserSubscriptions(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	// Cache the result with appropriate TTL
	result := struct {
		Subscriptions []*entities.UserSubscription `json:"subscriptions"`
		Total         int64                        `json:"total"`
	}{
		Subscriptions: subscriptions,
		Total:         total,
	}

	if data, err := json.Marshal(result); err == nil {
		_ = s.cacheManager.GetCache().Set(ctx, cacheKey, data, cacheTTL)
	}

	return subscriptions, total, nil
}

// GetUserSubscriptionsWithRelations gets user subscriptions with relations (cached with shorter TTL)
func (s *CachedUserSubscriptionService) GetUserSubscriptionsWithRelations(ctx context.Context, req *interfaces.GetUserSubscriptionsRequest) ([]*entities.UserSubscription, int64, error) {
	cacheKey := s.buildSubscriptionListCacheKey(req) + ":relations"

	cached, err := s.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var result struct {
			Subscriptions []*entities.UserSubscription `json:"subscriptions"`
			Total         int64                        `json:"total"`
		}
		if err := json.Unmarshal(cached, &result); err == nil {
			return result.Subscriptions, result.Total, nil
		}
	}

	// Cache miss - fetch from database
	subscriptions, total, err := s.UserSubscriptionService.GetUserSubscriptionsWithRelations(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	// Cache the result with shorter TTL for relations
	result := struct {
		Subscriptions []*entities.UserSubscription `json:"subscriptions"`
		Total         int64                        `json:"total"`
	}{
		Subscriptions: subscriptions,
		Total:         total,
	}

	if data, err := json.Marshal(result); err == nil {
		_ = s.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.ShortCacheTTL)
	}

	return subscriptions, total, nil
}

// GetActiveUserSubscription gets active subscription with caching
func (s *CachedUserSubscriptionService) GetActiveUserSubscription(ctx context.Context, userID, planID uint) (*entities.UserSubscription, error) {
	cacheKey := fmt.Sprintf("active:user:%d:plan:%d", userID, planID)

	subscription, err := s.subscriptionCache.Get(ctx, cacheKey, func() (*entities.UserSubscription, error) {
		return s.UserSubscriptionService.GetActiveUserSubscription(ctx, userID, planID)
	})

	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// GetUserActiveSubscriptions gets all active subscriptions for a user with caching
func (s *CachedUserSubscriptionService) GetUserActiveSubscriptions(ctx context.Context, userID uint) ([]*entities.UserSubscription, error) {
	cacheKey := s.cacheKeys.Subscription.UserActiveSubscriptions(userID)

	cached, err := s.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var subscriptions []*entities.UserSubscription
		if err := json.Unmarshal(cached, &subscriptions); err == nil {
			return subscriptions, nil
		}
	}

	// Cache miss - fetch from database
	subscriptions, err := s.UserSubscriptionService.GetUserActiveSubscriptions(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if data, err := json.Marshal(subscriptions); err == nil {
		_ = s.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.MediumCacheTTL)
	}

	return subscriptions, nil
}

// UpdateUserSubscription updates a user subscription with cache invalidation
func (s *CachedUserSubscriptionService) UpdateUserSubscription(ctx context.Context, subscriptionID uint, req *interfaces.UpdateSubscriptionRequest) (*entities.UserSubscription, error) {
	subscription, err := s.UserSubscriptionService.UpdateUserSubscription(ctx, subscriptionID, req)
	if err != nil {
		return nil, err
	}

	// Invalidate caches for this subscription
	s.invalidateSubscriptionCaches(ctx, subscriptionID, subscription.UserID)

	// Write-through: cache the updated subscription
	if subscription != nil {
		if err := s.subscriptionCache.Set(ctx, subscription); err != nil {
			logger.Error("Failed to cache updated subscription",
				logger.Uint("subscription_id", subscription.ID),
				logger.ErrorField(err))
		}
	}

	return subscription, nil
}

// CancelUserSubscription cancels a subscription with cache invalidation
func (s *CachedUserSubscriptionService) CancelUserSubscription(ctx context.Context, subscriptionID uint, reason string, cancelAtPeriodEnd bool) error {
	// Get user ID before cancellation for cache invalidation
	subscription, err := s.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	err = s.UserSubscriptionService.CancelUserSubscription(ctx, subscriptionID, reason, cancelAtPeriodEnd)
	if err != nil {
		return err
	}

	// Invalidate caches
	s.invalidateSubscriptionCaches(ctx, subscriptionID, subscription.UserID)

	return nil
}

// RenewUserSubscription renews a subscription with cache invalidation
func (s *CachedUserSubscriptionService) RenewUserSubscription(ctx context.Context, subscriptionID uint) error {
	// Get user ID before renewal for cache invalidation
	subscription, err := s.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	err = s.UserSubscriptionService.RenewUserSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	// Invalidate caches
	s.invalidateSubscriptionCaches(ctx, subscriptionID, subscription.UserID)

	return nil
}

// DeleteUserSubscription deletes a subscription with cache invalidation
func (s *CachedUserSubscriptionService) DeleteUserSubscription(ctx context.Context, subscriptionID uint) error {
	// Get user ID before deletion for cache invalidation
	subscription, err := s.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	err = s.UserSubscriptionService.DeleteUserSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	// Invalidate all caches for this subscription
	s.invalidateSubscriptionCaches(ctx, subscriptionID, subscription.UserID)

	return nil
}

// UpdateLastUsed updates last used with selective cache invalidation
func (s *CachedUserSubscriptionService) UpdateLastUsed(ctx context.Context, subscriptionID uint) error {
	err := s.UserSubscriptionService.UpdateLastUsed(ctx, subscriptionID)
	if err != nil {
		return err
	}

	// Only invalidate the specific subscription cache, not all user caches
	// since this is a frequent operation
	cacheKey := fmt.Sprintf("id:%d", subscriptionID)
	if err := s.subscriptionCache.Invalidate(ctx, cacheKey); err != nil {
		logger.Error("Failed to invalidate subscription cache after last used update",
			logger.Uint("subscription_id", subscriptionID),
			logger.ErrorField(err))
	}

	return nil
}

// UpdateTrafficUsage updates traffic with selective cache invalidation
func (s *CachedUserSubscriptionService) UpdateTrafficUsage(ctx context.Context, subscriptionID uint, usedBytes int64) error {
	err := s.UserSubscriptionService.UpdateTrafficUsage(ctx, subscriptionID, usedBytes)
	if err != nil {
		return err
	}

	// Only invalidate the specific subscription cache for frequent updates
	cacheKey := fmt.Sprintf("id:%d", subscriptionID)
	if err := s.subscriptionCache.Invalidate(ctx, cacheKey); err != nil {
		logger.Error("Failed to invalidate subscription cache after traffic update",
			logger.Uint("subscription_id", subscriptionID),
			logger.ErrorField(err))
	}

	return nil
}

// Traffic and renewal methods with cache invalidation

// ResetTrafficUsage resets traffic usage with cache invalidation
func (s *CachedUserSubscriptionService) ResetTrafficUsage(ctx context.Context, subscriptionID uint, adminUserID uint) (*entities.UserSubscription, error) {
	subscription, err := s.UserSubscriptionService.ResetTrafficUsage(ctx, subscriptionID, adminUserID)
	if err != nil {
		return nil, err
	}

	// Invalidate related caches
	s.invalidateSubscriptionCaches(ctx, subscriptionID, subscription.UserID)

	// Cache the updated subscription
	if err := s.subscriptionCache.Set(ctx, subscription); err != nil {
		logger.Error("Failed to cache reset subscription",
			logger.Uint("subscription_id", subscription.ID),
			logger.ErrorField(err))
	}

	return subscription, nil
}

func (s *CachedUserSubscriptionService) ResetTrafficForSubscriptions(ctx context.Context, req *ResetTrafficRequest) (int, error) {
	count, err := s.UserSubscriptionService.ResetTrafficForSubscriptions(ctx, req)
	if err != nil {
		return count, err
	}

	// Invalidate broad patterns since we don't know which subscriptions were affected
	s.invalidateAllUserSubscriptionCaches(ctx)
	return count, nil
}

func (s *CachedUserSubscriptionService) UpdateSubscriptionTrafficLimit(ctx context.Context, subscriptionID uint, newLimit int64, resetCycle string) error {
	subscription, err := s.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	err = s.UserSubscriptionService.UpdateSubscriptionTrafficLimit(ctx, subscriptionID, newLimit, resetCycle)
	if err != nil {
		return err
	}

	s.invalidateSubscriptionCaches(ctx, subscriptionID, subscription.UserID)
	return nil
}

func (s *CachedUserSubscriptionService) ProcessAutoRenewals(ctx context.Context) (int, error) {
	count, err := s.UserSubscriptionService.ProcessAutoRenewals(ctx)
	if err != nil {
		return count, err
	}

	// Invalidate all subscription caches since multiple subscriptions may have been affected
	s.invalidateAllUserSubscriptionCaches(ctx)
	return count, nil
}

func (s *CachedUserSubscriptionService) ProcessSubscriptionAutoRenewal(ctx context.Context, subscriptionID uint) error {
	subscription, err := s.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	err = s.UserSubscriptionService.ProcessSubscriptionAutoRenewal(ctx, subscriptionID)
	if err != nil {
		return err
	}

	s.invalidateSubscriptionCaches(ctx, subscriptionID, subscription.UserID)
	return nil
}

func (s *CachedUserSubscriptionService) GetSubscriptionsForAutoRenewal(ctx context.Context) ([]*entities.UserSubscription, error) {
	// This is likely called from a background job, cache with short TTL
	cacheKey := "auto_renewal:eligible"

	cached, err := s.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var subscriptions []*entities.UserSubscription
		if err := json.Unmarshal(cached, &subscriptions); err == nil {
			return subscriptions, nil
		}
	}

	subscriptions, err := s.UserSubscriptionService.GetSubscriptionsForAutoRenewal(ctx)
	if err != nil {
		return nil, err
	}

	// Cache with very short TTL since this data changes frequently
	if data, err := json.Marshal(subscriptions); err == nil {
		_ = s.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.ShortCacheTTL)
	}

	return subscriptions, nil
}

func (s *CachedUserSubscriptionService) EnableAutoRenewal(ctx context.Context, subscriptionID uint) error {
	subscription, err := s.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	err = s.UserSubscriptionService.EnableAutoRenewal(ctx, subscriptionID)
	if err != nil {
		return err
	}

	s.invalidateSubscriptionCaches(ctx, subscriptionID, subscription.UserID)
	return nil
}

func (s *CachedUserSubscriptionService) DisableAutoRenewal(ctx context.Context, subscriptionID uint) error {
	subscription, err := s.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	err = s.UserSubscriptionService.DisableAutoRenewal(ctx, subscriptionID)
	if err != nil {
		return err
	}

	s.invalidateSubscriptionCaches(ctx, subscriptionID, subscription.UserID)
	return nil
}

// Utility and statistics methods

func (s *CachedUserSubscriptionService) GetSubscriptionTrafficStats(ctx context.Context, subscriptionID uint) (map[string]any, error) {
	cacheKey := fmt.Sprintf("stats:traffic:%d", subscriptionID)

	cached, err := s.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var stats map[string]any
		if err := json.Unmarshal(cached, &stats); err == nil {
			return stats, nil
		}
	}

	stats, err := s.UserSubscriptionService.GetSubscriptionTrafficStats(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	// Cache traffic stats with short TTL since they change frequently
	if data, err := json.Marshal(stats); err == nil {
		_ = s.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.ShortCacheTTL)
	}

	return stats, nil
}

func (s *CachedUserSubscriptionService) CheckAndProcessExpiredSubscriptions(ctx context.Context) error {
	err := s.UserSubscriptionService.CheckAndProcessExpiredSubscriptions(ctx)
	if err != nil {
		return err
	}

	// Invalidate all caches since we don't know which subscriptions were processed
	s.invalidateAllUserSubscriptionCaches(ctx)
	return nil
}

func (s *CachedUserSubscriptionService) ExtendSubscription(ctx context.Context, subscriptionID uint, extendByDays int, reason string) error {
	subscription, err := s.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	err = s.UserSubscriptionService.ExtendSubscription(ctx, subscriptionID, extendByDays, reason)
	if err != nil {
		return err
	}

	s.invalidateSubscriptionCaches(ctx, subscriptionID, subscription.UserID)
	return nil
}

func (s *CachedUserSubscriptionService) GetSubscriptionStatistics(ctx context.Context) (map[string]any, error) {
	cacheKey := "statistics:all"

	cached, err := s.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var stats map[string]any
		if err := json.Unmarshal(cached, &stats); err == nil {
			return stats, nil
		}
	}

	stats, err := s.UserSubscriptionService.GetSubscriptionStatistics(ctx)
	if err != nil {
		return nil, err
	}

	// Cache statistics with medium TTL
	if data, err := json.Marshal(stats); err == nil {
		_ = s.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.MediumCacheTTL)
	}

	return stats, nil
}

// Cache invalidation helper methods

func (s *CachedUserSubscriptionService) invalidateSubscriptionCaches(ctx context.Context, subscriptionID uint, userID uint) {
	// Invalidate specific subscription caches
	cacheKeys := []string{
		fmt.Sprintf("id:%d", subscriptionID),
		fmt.Sprintf("relations:id:%d", subscriptionID),
	}

	for _, key := range cacheKeys {
		if err := s.subscriptionCache.Invalidate(ctx, key); err != nil {
			logger.Error("Failed to invalidate subscription cache",
				logger.String("key", key),
				logger.ErrorField(err))
		}
	}

	// Invalidate user-related caches
	s.invalidateUserCaches(ctx, userID)
}

func (s *CachedUserSubscriptionService) invalidateUserCaches(ctx context.Context, userID uint) {
	// Invalidate user-specific caches
	userCacheKeys := []string{
		s.cacheKeys.Subscription.UserSubscription(userID),
		s.cacheKeys.Subscription.UserActiveSubscriptions(userID),
	}

	for _, key := range userCacheKeys {
		if err := s.cacheManager.GetCache().Delete(ctx, key); err != nil {
			logger.Error("Failed to invalidate user subscription cache",
				logger.String("key", key),
				logger.ErrorField(err))
		}
	}

	// Invalidate list caches for this user only (admin queries don't use cache)
	patterns := []string{
		fmt.Sprintf("list:*user:%d*", userID),
		fmt.Sprintf("active:user:%d*", userID),
	}

	for _, pattern := range patterns {
		fullPattern := cache.CachePrefixSubscription + pattern
		if err := s.cacheManager.GetCache().DeleteByPattern(ctx, fullPattern); err != nil {
			logger.Error("Failed to invalidate subscription cache pattern",
				logger.String("pattern", fullPattern),
				logger.ErrorField(err))
		}
	}
}

func (s *CachedUserSubscriptionService) invalidateAllUserSubscriptionCaches(ctx context.Context) {
	// Invalidate all subscription caches - use sparingly
	patterns := []string{
		cache.CachePrefixSubscription + "*",
	}

	for _, pattern := range patterns {
		if err := s.cacheManager.GetCache().DeleteByPattern(ctx, pattern); err != nil {
			logger.Error("Failed to invalidate all subscription caches",
				logger.String("pattern", pattern),
				logger.ErrorField(err))
		}
	}
}

func (s *CachedUserSubscriptionService) buildSubscriptionListCacheKey(req *interfaces.GetUserSubscriptionsRequest) string {
	var keyParts []string
	keyParts = append(keyParts, "list")

	// Distinguish between admin and user queries
	if req.UserID > 0 {
		keyParts = append(keyParts, "user", fmt.Sprintf("%d", req.UserID))
	} else {
		keyParts = append(keyParts, "admin", "all")
	}

	if req.Status != "" {
		keyParts = append(keyParts, "status", req.Status)
	}

	keyParts = append(keyParts, fmt.Sprintf("limit:%d", req.Limit))
	keyParts = append(keyParts, fmt.Sprintf("offset:%d", req.Offset))

	return cache.CachePrefixSubscription + strings.Join(keyParts, ":")
}

// shouldSkipCache determines if caching should be skipped for this request
func (s *CachedUserSubscriptionService) shouldSkipCache(req *interfaces.GetUserSubscriptionsRequest) bool {
	// ALWAYS skip cache for admin queries - admins need real-time data
	// Admin queries are identified by UserID == 0
	if req.UserID == 0 {
		return true
	}

	// Skip cache for large user queries (likely exports or bulk operations)
	if req.UserID > 0 && req.Limit >= 50 {
		return true
	}

	return false
}

// determineCacheTTL returns appropriate cache TTL based on request characteristics
func (s *CachedUserSubscriptionService) determineCacheTTL(req *interfaces.GetUserSubscriptionsRequest) time.Duration {
	// User-specific queries: medium TTL for balance of performance and freshness
	// Small limit queries (like pagination) use longer cache
	if req.UserID > 0 && req.Limit <= 20 {
		return cache.MediumCacheTTL // 15 minutes
	}

	// Larger user queries: shorter TTL for more recent data
	if req.UserID > 0 {
		return cache.ShortCacheTTL // 1 minute
	}

	// Fallback (should rarely be used since admin queries skip cache)
	return cache.ShortCacheTTL
}

// PauseUserSubscription pauses a user subscription with cache invalidation
func (s *CachedUserSubscriptionService) PauseUserSubscription(ctx context.Context, subscriptionID uint, req *interfaces.PauseSubscriptionRequest, adminUserID uint) (*entities.UserSubscription, error) {
	// Call base service method
	subscription, err := s.UserSubscriptionService.PauseUserSubscription(ctx, subscriptionID, req, adminUserID)
	if err != nil {
		return nil, err
	}

	// Invalidate related caches
	s.invalidateSubscriptionCaches(ctx, subscription.ID, subscription.UserID)

	logger.Info("User subscription paused and cache invalidated",
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID),
		logger.Uint("admin_user_id", adminUserID))

	return subscription, nil
}

// ResumeUserSubscription resumes a user subscription with cache invalidation
func (s *CachedUserSubscriptionService) ResumeUserSubscription(ctx context.Context, subscriptionID uint, req *interfaces.ResumeSubscriptionRequest, adminUserID uint) (*entities.UserSubscription, error) {
	// Call base service method
	subscription, err := s.UserSubscriptionService.ResumeUserSubscription(ctx, subscriptionID, req, adminUserID)
	if err != nil {
		return nil, err
	}

	// Invalidate related caches
	s.invalidateSubscriptionCaches(ctx, subscription.ID, subscription.UserID)

	logger.Info("User subscription resumed and cache invalidated",
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID),
		logger.Uint("admin_user_id", adminUserID))

	return subscription, nil
}

// UpgradeUserSubscription upgrades subscription with cache invalidation
func (s *CachedUserSubscriptionService) UpgradeUserSubscription(ctx context.Context, req *interfaces.UpgradeSubscriptionRequest) (*entities.UserSubscription, error) {
	subscription, err := s.UserSubscriptionService.UpgradeUserSubscription(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate related caches
	s.invalidateSubscriptionCaches(ctx, subscription.ID, subscription.UserID)

	// Cache the updated subscription
	if err := s.subscriptionCache.Set(ctx, subscription); err != nil {
		logger.Error("Failed to cache upgraded subscription",
			logger.Uint("subscription_id", subscription.ID),
			logger.ErrorField(err))
	}

	logger.Info("Subscription upgraded and cache invalidated",
		logger.Uint("subscription_id", subscription.ID),
		logger.Uint("user_id", subscription.UserID),
		logger.Uint("new_plan_id", req.NewSubscriptionPlanID))

	return subscription, nil
}

// DowngradeUserSubscription downgrades subscription with cache invalidation
func (s *CachedUserSubscriptionService) DowngradeUserSubscription(ctx context.Context, req *interfaces.DowngradeSubscriptionRequest) (*entities.UserSubscription, error) {
	subscription, err := s.UserSubscriptionService.DowngradeUserSubscription(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate related caches
	s.invalidateSubscriptionCaches(ctx, subscription.ID, subscription.UserID)

	// Cache the updated subscription
	if err := s.subscriptionCache.Set(ctx, subscription); err != nil {
		logger.Error("Failed to cache downgraded subscription",
			logger.Uint("subscription_id", subscription.ID),
			logger.ErrorField(err))
	}

	logger.Info("Subscription downgraded and cache invalidated",
		logger.Uint("subscription_id", subscription.ID),
		logger.Uint("user_id", subscription.UserID),
		logger.Uint("new_plan_id", req.NewSubscriptionPlanID))

	return subscription, nil
}

// ProcessSubscriptionChange processes subscription change with cache invalidation
func (s *CachedUserSubscriptionService) ProcessSubscriptionChange(ctx context.Context, subscriptionID uint, newPlanID uint, changeType string) (*entities.UserSubscription, error) {
	subscription, err := s.UserSubscriptionService.ProcessSubscriptionChange(ctx, subscriptionID, newPlanID, changeType)
	if err != nil {
		return nil, err
	}

	// Invalidate related caches
	s.invalidateSubscriptionCaches(ctx, subscriptionID, subscription.UserID)

	// Cache the updated subscription
	if err := s.subscriptionCache.Set(ctx, subscription); err != nil {
		logger.Error("Failed to cache changed subscription",
			logger.Uint("subscription_id", subscription.ID),
			logger.ErrorField(err))
	}

	logger.Info("Subscription change processed and cache invalidated",
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID),
		logger.String("change_type", changeType),
		logger.Uint("new_plan_id", newPlanID))

	return subscription, nil
}
