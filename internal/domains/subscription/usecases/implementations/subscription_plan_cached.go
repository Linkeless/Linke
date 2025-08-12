package implementations

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/cache"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

type CachedSubscriptionPlanService struct {
	*SubscriptionPlanService
	cacheManager cache.CacheManager
	cacheKeys    *cache.AllCacheKeys
	planCache    *cache.CacheAside[entities.SubscriptionPlan]
}

func NewSubscriptionPlanServiceWithCache(
	db *gorm.DB,
	cacheManager cache.CacheManager,
	cacheKeys *cache.AllCacheKeys,
) interfaces.SubscriptionPlanService {
	baseService := NewSubscriptionPlanService(db)

	planCache := cache.NewCacheAside[entities.SubscriptionPlan](
		cacheManager.GetCache(),
		cache.CachePrefixPlan,
		func(plan entities.SubscriptionPlan) string {
			return fmt.Sprintf("id:%d", plan.ID)
		},
		cache.LongCacheTTL, // 1 hour TTL for plans
	)

	return &CachedSubscriptionPlanService{
		SubscriptionPlanService: baseService,
		cacheManager:            cacheManager,
		cacheKeys:               cacheKeys,
		planCache:               planCache,
	}
}

// CreateSubscriptionPlan creates a new subscription plan with cache invalidation
func (s *CachedSubscriptionPlanService) CreateSubscriptionPlan(ctx context.Context, creatorID uint, req *interfaces.CreateSubscriptionPlanRequest) (*entities.SubscriptionPlan, error) {
	plan, err := s.SubscriptionPlanService.CreateSubscriptionPlan(ctx, creatorID, req)
	if err != nil {
		return nil, err
	}

	// Write-through: cache the new plan
	if plan != nil {
		if err := s.planCache.Set(ctx, plan); err != nil {
			logger.Error("Failed to cache new plan",
				logger.Uint("plan_id", plan.ID),
				logger.ErrorField(err))
		}

		// Invalidate list caches
		s.invalidateListCaches(ctx)
	}

	return plan, nil
}

// GetSubscriptionPlan gets a subscription plan by ID with caching
func (s *CachedSubscriptionPlanService) GetSubscriptionPlan(ctx context.Context, planID uint) (*entities.SubscriptionPlan, error) {
	cacheKey := s.cacheKeys.Subscription.PlanByID(planID)

	plan, err := s.planCache.Get(ctx, cacheKey, func() (*entities.SubscriptionPlan, error) {
		return s.SubscriptionPlanService.GetSubscriptionPlan(ctx, planID)
	})

	if err != nil {
		return nil, err
	}

	return plan, nil
}

// GetSubscriptionPlanByCode gets a subscription plan by code with caching
func (s *CachedSubscriptionPlanService) GetSubscriptionPlanByCode(ctx context.Context, code string) (*entities.SubscriptionPlan, error) {
	cacheKey := fmt.Sprintf("code:%s", code)

	plan, err := s.planCache.Get(ctx, cacheKey, func() (*entities.SubscriptionPlan, error) {
		return s.SubscriptionPlanService.GetSubscriptionPlanByCode(ctx, code)
	})

	if err != nil {
		return nil, err
	}

	return plan, nil
}

// GetSubscriptionPlans gets subscription plans with caching for list results
func (s *CachedSubscriptionPlanService) GetSubscriptionPlans(ctx context.Context, req *interfaces.GetSubscriptionPlansRequest) ([]*entities.SubscriptionPlan, int64, error) {
	// For complex queries with filters, we'll cache based on the query parameters
	cacheKey := s.buildListCacheKey(req)

	// Use cache decorator for list results
	cached, err := s.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var result struct {
			Plans []*entities.SubscriptionPlan `json:"plans"`
			Total int64                        `json:"total"`
		}
		if err := json.Unmarshal(cached, &result); err == nil {
			return result.Plans, result.Total, nil
		}
	}

	// Cache miss - fetch from database
	plans, total, err := s.SubscriptionPlanService.GetSubscriptionPlans(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	// Cache the result
	result := struct {
		Plans []*entities.SubscriptionPlan `json:"plans"`
		Total int64                        `json:"total"`
	}{
		Plans: plans,
		Total: total,
	}

	if data, err := json.Marshal(result); err == nil {
		_ = s.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.MediumCacheTTL)
	}

	return plans, total, nil
}

// GetVisibleSubscriptionPlans gets visible plans with caching
func (s *CachedSubscriptionPlanService) GetVisibleSubscriptionPlans(ctx context.Context, currency string) ([]*entities.SubscriptionPlan, error) {
	cacheKey := s.cacheKeys.Subscription.ActivePlans()
	if currency != "" {
		cacheKey = fmt.Sprintf("%s:currency:%s", cacheKey, currency)
	}

	cached, err := s.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var plans []*entities.SubscriptionPlan
		if err := json.Unmarshal(cached, &plans); err == nil {
			return plans, nil
		}
	}

	// Cache miss - fetch from database
	plans, err := s.SubscriptionPlanService.GetVisibleSubscriptionPlans(ctx, currency)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if data, err := json.Marshal(plans); err == nil {
		_ = s.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.LongCacheTTL)
	}

	return plans, nil
}

// GetPopularSubscriptionPlans gets popular plans with caching
func (s *CachedSubscriptionPlanService) GetPopularSubscriptionPlans(ctx context.Context, limit int) ([]*entities.SubscriptionPlan, error) {
	cacheKey := fmt.Sprintf("popular:limit:%d", limit)

	cached, err := s.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var plans []*entities.SubscriptionPlan
		if err := json.Unmarshal(cached, &plans); err == nil {
			return plans, nil
		}
	}

	// Cache miss - fetch from database
	plans, err := s.SubscriptionPlanService.GetPopularSubscriptionPlans(ctx, limit)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if data, err := json.Marshal(plans); err == nil {
		_ = s.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.LongCacheTTL)
	}

	return plans, nil
}

// UpdateSubscriptionPlan updates a subscription plan with cache invalidation
func (s *CachedSubscriptionPlanService) UpdateSubscriptionPlan(ctx context.Context, planID uint, req *interfaces.UpdateSubscriptionPlanRequest) (*entities.SubscriptionPlan, error) {
	plan, err := s.SubscriptionPlanService.UpdateSubscriptionPlan(ctx, planID, req)
	if err != nil {
		return nil, err
	}

	// Invalidate caches for this plan
	s.invalidatePlanCaches(ctx, planID)

	// Write-through: cache the updated plan
	if plan != nil {
		if err := s.planCache.Set(ctx, plan); err != nil {
			logger.Error("Failed to cache updated plan",
				logger.Uint("plan_id", plan.ID),
				logger.ErrorField(err))
		}
	}

	return plan, nil
}

// DeleteSubscriptionPlan deletes a subscription plan with cache invalidation
func (s *CachedSubscriptionPlanService) DeleteSubscriptionPlan(ctx context.Context, planID uint) error {
	err := s.SubscriptionPlanService.DeleteSubscriptionPlan(ctx, planID)
	if err != nil {
		return err
	}

	// Invalidate all caches for this plan
	s.invalidatePlanCaches(ctx, planID)

	return nil
}

// ToggleSubscriptionPlanStatus toggles status with cache invalidation
func (s *CachedSubscriptionPlanService) ToggleSubscriptionPlanStatus(ctx context.Context, planID uint) (*entities.SubscriptionPlan, error) {
	plan, err := s.SubscriptionPlanService.ToggleSubscriptionPlanStatus(ctx, planID)
	if err != nil {
		return nil, err
	}

	// Invalidate caches for this plan
	s.invalidatePlanCaches(ctx, planID)

	// Write-through: cache the updated plan
	if plan != nil {
		if err := s.planCache.Set(ctx, plan); err != nil {
			logger.Error("Failed to cache plan after status toggle",
				logger.Uint("plan_id", plan.ID),
				logger.ErrorField(err))
		}
	}

	return plan, nil
}

// ArchiveSubscriptionPlan archives a plan with cache invalidation
func (s *CachedSubscriptionPlanService) ArchiveSubscriptionPlan(ctx context.Context, planID uint) error {
	err := s.SubscriptionPlanService.ArchiveSubscriptionPlan(ctx, planID)
	if err != nil {
		return err
	}

	// Invalidate all caches for this plan
	s.invalidatePlanCaches(ctx, planID)

	return nil
}

// Cache invalidation helper methods

func (s *CachedSubscriptionPlanService) invalidatePlanCaches(ctx context.Context, planID uint) {
	// Invalidate specific plan caches
	planByIDKey := s.cacheKeys.Subscription.PlanByID(planID)
	if err := s.planCache.Invalidate(ctx, planByIDKey); err != nil {
		logger.Error("Failed to invalidate plan by ID cache",
			logger.Uint("plan_id", planID),
			logger.ErrorField(err))
	}

	// Also try to get the plan's code to invalidate code cache
	if plan, err := s.SubscriptionPlanService.GetSubscriptionPlan(ctx, planID); err == nil && plan != nil {
		codeKey := fmt.Sprintf("code:%s", plan.Code)
		if err := s.planCache.Invalidate(ctx, codeKey); err != nil {
			logger.Error("Failed to invalidate plan by code cache",
				logger.String("code", plan.Code),
				logger.ErrorField(err))
		}
	}

	// Invalidate list caches
	s.invalidateListCaches(ctx)
}

func (s *CachedSubscriptionPlanService) invalidateListCaches(ctx context.Context) {
	// Invalidate all list caches
	patterns := []string{
		"list:*",
		"active:*",
		"popular:*",
		"visible:*",
	}

	for _, pattern := range patterns {
		fullPattern := cache.CachePrefixPlan + pattern
		if err := s.cacheManager.GetCache().DeleteByPattern(ctx, fullPattern); err != nil {
			logger.Error("Failed to invalidate plan list cache pattern",
				logger.String("pattern", fullPattern),
				logger.ErrorField(err))
		}
	}
}

func (s *CachedSubscriptionPlanService) buildListCacheKey(req *interfaces.GetSubscriptionPlansRequest) string {
	var keyParts []string
	keyParts = append(keyParts, "list")

	if req.Status != "" {
		keyParts = append(keyParts, "status", req.Status)
	}
	if req.Currency != "" {
		keyParts = append(keyParts, "currency", req.Currency)
	}
	if req.Visible != nil {
		keyParts = append(keyParts, "visible", fmt.Sprintf("%t", *req.Visible))
	}
	if req.Popular != nil {
		keyParts = append(keyParts, "popular", fmt.Sprintf("%t", *req.Popular))
	}
	if req.Recommended != nil {
		keyParts = append(keyParts, "recommended", fmt.Sprintf("%t", *req.Recommended))
	}

	keyParts = append(keyParts, fmt.Sprintf("limit:%d", req.Limit))
	keyParts = append(keyParts, fmt.Sprintf("offset:%d", req.Offset))

	return cache.CachePrefixPlan + strings.Join(keyParts, ":")
}
