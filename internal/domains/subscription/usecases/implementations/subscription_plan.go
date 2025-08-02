package implementations

import (
	"context"
	"fmt"
	"strings"

	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

type SubscriptionPlanService struct {
	db *gorm.DB
}

func NewSubscriptionPlanService(db *gorm.DB) *SubscriptionPlanService {
	return &SubscriptionPlanService{
		db: db,
	}
}


// CreateSubscriptionPlan creates a new subscription plan
func (s *SubscriptionPlanService) CreateSubscriptionPlan(ctx context.Context, creatorID uint, req *interfaces.CreateSubscriptionPlanRequest) (*entities.SubscriptionPlan, error) {
	// Validate and normalize code
	code := strings.TrimSpace(strings.ToLower(req.Code))
	if code == "" {
		return nil, fmt.Errorf("plan code cannot be empty")
	}

	// Check if code already exists
	var existingPlan entities.SubscriptionPlan
	if err := s.db.Where("code = ?", code).First(&existingPlan).Error; err == nil {
		return nil, fmt.Errorf("plan with code '%s' already exists", code)
	} else if err != gorm.ErrRecordNotFound {
		logger.Error("Failed to check existing plan", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to check existing plan: %w", err)
	}

	// Set defaults
	isVisible := true
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}

	isPopular := false
	if req.IsPopular != nil {
		isPopular = *req.IsPopular
	}

	isRecommended := false
	if req.IsRecommended != nil {
		isRecommended = *req.IsRecommended
	}

	// Create the plan
	plan := &entities.SubscriptionPlan{
		Name:            req.Name,
		Code:            code,
		Description:     req.Description,
		Price:           req.Price,
		Currency:        strings.ToUpper(req.Currency),
		BillingCycle:    req.BillingCycle,
		BillingInterval: req.BillingInterval,
		TrialPeriodDays: req.TrialPeriodDays,
		Features:        req.Features,
		Limits:          req.Limits,
		Status:          entities.SubscriptionPlanStatusActive,
		IsVisible:       isVisible,
		SortOrder:       req.SortOrder,
		IsPopular:       isPopular,
		IsRecommended:   isRecommended,
		SetupFee:        req.SetupFee,
		CancellationFee: req.CancellationFee,

		// Traffic Configuration (Required for all plans)
		TrafficLimit:      req.TrafficLimit,
		TrafficResetCycle: req.TrafficResetCycle,
	}

	if err := s.db.WithContext(ctx).Create(plan).Error; err != nil {
		logger.Error("Failed to create subscription plan", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to create subscription plan: %w", err)
	}

	logger.Info("Subscription plan created successfully",
		logger.Uint("plan_id", plan.ID),
		logger.String("plan_code", plan.Code),
		logger.Uint("creator_id", creatorID))

	return plan, nil
}

// GetSubscriptionPlan gets a subscription plan by ID
func (s *SubscriptionPlanService) GetSubscriptionPlan(ctx context.Context, planID uint) (*entities.SubscriptionPlan, error) {
	var plan entities.SubscriptionPlan
	if err := s.db.WithContext(ctx).First(&plan, planID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription plan not found")
		}
		logger.Error("Failed to get subscription plan", logger.Error2("error", err), logger.Uint("plan_id", planID))
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}

	return &plan, nil
}

// GetSubscriptionPlanByCode gets a subscription plan by code
func (s *SubscriptionPlanService) GetSubscriptionPlanByCode(ctx context.Context, code string) (*entities.SubscriptionPlan, error) {
	var plan entities.SubscriptionPlan
	if err := s.db.WithContext(ctx).Where("code = ?", code).First(&plan).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription plan not found")
		}
		logger.Error("Failed to get subscription plan by code", logger.Error2("error", err), logger.String("code", code))
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}

	return &plan, nil
}

// GetSubscriptionPlans gets subscription plans with filtering and pagination
func (s *SubscriptionPlanService) GetSubscriptionPlans(ctx context.Context, req *interfaces.GetSubscriptionPlansRequest) ([]*entities.SubscriptionPlan, int64, error) {
	query := s.db.WithContext(ctx).Model(&entities.SubscriptionPlan{})

	// Apply filters
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.Currency != "" {
		query = query.Where("currency = ?", strings.ToUpper(req.Currency))
	}

	if req.Visible != nil {
		query = query.Where("is_visible = ?", *req.Visible)
	}

	if req.Popular != nil {
		query = query.Where("is_popular = ?", *req.Popular)
	}

	if req.Recommended != nil {
		query = query.Where("is_recommended = ?", *req.Recommended)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		logger.Error("Failed to count subscription plans", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count subscription plans: %w", err)
	}

	// Apply pagination and ordering
	query = query.Order("sort_order ASC, created_at DESC")

	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	var plans []*entities.SubscriptionPlan
	if err := query.Find(&plans).Error; err != nil {
		logger.Error("Failed to get subscription plans", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to get subscription plans: %w", err)
	}

	return plans, totalCount, nil
}

// GetVisibleSubscriptionPlans gets visible and active subscription plans for public display
func (s *SubscriptionPlanService) GetVisibleSubscriptionPlans(ctx context.Context, currency string) ([]*entities.SubscriptionPlan, error) {
	query := s.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where("status = ? AND is_visible = ?", entities.SubscriptionPlanStatusActive, true)

	if currency != "" {
		query = query.Where("currency = ?", strings.ToUpper(currency))
	}

	var plans []*entities.SubscriptionPlan
	if err := query.Order("sort_order ASC, created_at ASC").Find(&plans).Error; err != nil {
		logger.Error("Failed to get public subscription plans", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get public subscription plans: %w", err)
	}

	return plans, nil
}

// GetPopularSubscriptionPlans gets popular subscription plans
func (s *SubscriptionPlanService) GetPopularSubscriptionPlans(ctx context.Context, limit int) ([]*entities.SubscriptionPlan, error) {
	query := s.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where("status = ? AND is_visible = ? AND is_popular = ?", entities.SubscriptionPlanStatusActive, true, true)

	if limit > 0 {
		query = query.Limit(limit)
	}

	var plans []*entities.SubscriptionPlan
	if err := query.Order("sort_order ASC, created_at ASC").Find(&plans).Error; err != nil {
		logger.Error("Failed to get popular subscription plans", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get popular subscription plans: %w", err)
	}

	return plans, nil
}

// UpdateSubscriptionPlan updates a subscription plan
func (s *SubscriptionPlanService) UpdateSubscriptionPlan(ctx context.Context, planID uint, req *interfaces.UpdateSubscriptionPlanRequest) (*entities.SubscriptionPlan, error) {
	// Get existing plan
	plan, err := s.GetSubscriptionPlan(ctx, planID)
	if err != nil {
		return nil, err
	}

	// Prepare updates
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if req.Price != nil {
		updates["price"] = *req.Price
	}

	if req.TrialPeriodDays != nil {
		updates["trial_period_days"] = *req.TrialPeriodDays
	}

	if req.Features != nil {
		updates["features"] = *req.Features
	}

	if req.Limits != nil {
		updates["limits"] = *req.Limits
	}

	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if req.IsVisible != nil {
		updates["is_visible"] = *req.IsVisible
	}

	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}

	if req.IsPopular != nil {
		updates["is_popular"] = *req.IsPopular
	}

	if req.IsRecommended != nil {
		updates["is_recommended"] = *req.IsRecommended
	}

	if req.SetupFee != nil {
		updates["setup_fee"] = *req.SetupFee
	}

	if req.CancellationFee != nil {
		updates["cancellation_fee"] = *req.CancellationFee
	}

	// Traffic Configuration
	if req.TrafficLimit != nil {
		updates["traffic_limit"] = *req.TrafficLimit
	}

	if req.TrafficResetCycle != nil {
		updates["traffic_reset_cycle"] = *req.TrafficResetCycle
	}

	// Update the plan
	if err := s.db.WithContext(ctx).Model(plan).Updates(updates).Error; err != nil {
		logger.Error("Failed to update subscription plan", logger.Error2("error", err), logger.Uint("plan_id", planID))
		return nil, fmt.Errorf("failed to update subscription plan: %w", err)
	}

	// Reload the plan
	if err := s.db.WithContext(ctx).First(plan, planID).Error; err != nil {
		logger.Error("Failed to reload updated subscription plan", logger.Error2("error", err), logger.Uint("plan_id", planID))
		return nil, fmt.Errorf("failed to reload updated subscription plan: %w", err)
	}

	logger.Info("Subscription plan updated successfully", logger.Uint("plan_id", plan.ID))

	return plan, nil
}

// DeleteSubscriptionPlan soft deletes a subscription plan
func (s *SubscriptionPlanService) DeleteSubscriptionPlan(ctx context.Context, planID uint) error {
	// Check if plan exists
	plan, err := s.GetSubscriptionPlan(ctx, planID)
	if err != nil {
		return err
	}

	// Check if plan has active subscriptions
	var activeSubscriptionCount int64
	if err := s.db.WithContext(ctx).Model(&entities.UserSubscription{}).
		Where("subscription_plan_id = ? AND status = ?", planID, entities.UserSubscriptionStatusActive).
		Count(&activeSubscriptionCount).Error; err != nil {
		logger.Error("Failed to check active subscriptions", logger.Error2("error", err), logger.Uint("plan_id", planID))
		return fmt.Errorf("failed to check active subscriptions: %w", err)
	}

	if activeSubscriptionCount > 0 {
		return fmt.Errorf("cannot delete plan with %d active subscriptions", activeSubscriptionCount)
	}

	// Soft delete the plan
	if err := s.db.WithContext(ctx).Delete(plan).Error; err != nil {
		logger.Error("Failed to delete subscription plan", logger.Error2("error", err), logger.Uint("plan_id", planID))
		return fmt.Errorf("failed to delete subscription plan: %w", err)
	}

	logger.Info("Subscription plan deleted successfully", logger.Uint("plan_id", planID))

	return nil
}

// ToggleSubscriptionPlanStatus toggles a subscription plan's status between active and inactive
func (s *SubscriptionPlanService) ToggleSubscriptionPlanStatus(ctx context.Context, planID uint) (*entities.SubscriptionPlan, error) {
	// Get existing plan
	plan, err := s.GetSubscriptionPlan(ctx, planID)
	if err != nil {
		return nil, err
	}

	// Toggle status
	var newStatus string
	switch plan.Status {
	case entities.SubscriptionPlanStatusActive:
		newStatus = entities.SubscriptionPlanStatusInactive
	case entities.SubscriptionPlanStatusInactive:
		newStatus = entities.SubscriptionPlanStatusActive
	default:
		return nil, fmt.Errorf("cannot toggle status of archived plan")
	}

	// Update the plan
	if err := s.db.WithContext(ctx).Model(plan).Update("status", newStatus).Error; err != nil {
		logger.Error("Failed to toggle subscription plan status", logger.Error2("error", err), logger.Uint("plan_id", planID))
		return nil, fmt.Errorf("failed to toggle subscription plan status: %w", err)
	}

	// Reload the plan
	if err := s.db.WithContext(ctx).First(plan, planID).Error; err != nil {
		logger.Error("Failed to reload toggled subscription plan", logger.Error2("error", err), logger.Uint("plan_id", planID))
		return nil, fmt.Errorf("failed to reload toggled subscription plan: %w", err)
	}

	logger.Info("Subscription plan status toggled successfully", 
		logger.Uint("plan_id", plan.ID),
		logger.String("new_status", plan.Status))

	return plan, nil
}

// ArchiveSubscriptionPlan archives a subscription plan
func (s *SubscriptionPlanService) ArchiveSubscriptionPlan(ctx context.Context, planID uint) error {
	// Get existing plan
	plan, err := s.GetSubscriptionPlan(ctx, planID)
	if err != nil {
		return err
	}

	// Check if plan has active subscriptions
	var activeSubscriptionCount int64
	if err := s.db.WithContext(ctx).Model(&entities.UserSubscription{}).
		Where("subscription_plan_id = ? AND status = ?", planID, entities.UserSubscriptionStatusActive).
		Count(&activeSubscriptionCount).Error; err != nil {
		logger.Error("Failed to check active subscriptions for archiving", logger.Error2("error", err), logger.Uint("plan_id", planID))
		return fmt.Errorf("failed to check active subscriptions: %w", err)
	}

	if activeSubscriptionCount > 0 {
		return fmt.Errorf("cannot archive plan with %d active subscriptions", activeSubscriptionCount)
	}

	// Archive the plan
	if err := s.db.WithContext(ctx).Model(plan).Updates(map[string]interface{}{
		"status":     entities.SubscriptionPlanStatusArchived,
		"is_visible": false,
	}).Error; err != nil {
		logger.Error("Failed to archive subscription plan", logger.Error2("error", err), logger.Uint("plan_id", planID))
		return fmt.Errorf("failed to archive subscription plan: %w", err)
	}

	logger.Info("Subscription plan archived successfully", logger.Uint("plan_id", planID))

	return nil
}
