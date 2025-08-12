package implementations

import (
	"context"
	"fmt"
	"strings"

	"linke/internal/domains/subscription/constants"
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
		logger.Error("Failed to check existing plan", logger.ErrorField(err))
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
		Status:          constants.SubscriptionPlanStatusActive,
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

	// Set default server group IDs if provided
	if len(req.DefaultServerGroupIDs) > 0 {
		if err := plan.SetDefaultServerGroupIDs(req.DefaultServerGroupIDs); err != nil {
			return nil, fmt.Errorf("failed to set default server group IDs: %w", err)
		}
		logger.Info("Set default server group IDs for plan",
			logger.String("plan_code", code),
			logger.Any("server_group_ids", req.DefaultServerGroupIDs))
	}

	if err := s.db.WithContext(ctx).Create(plan).Error; err != nil {
		logger.Error("Failed to create subscription plan", logger.ErrorField(err))
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
		logger.Error("Failed to get subscription plan", logger.Uint("planID", uint(planID)))
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}

	// Fill server group name
	if groupID, err := plan.GetDefaultServerGroupID(); err == nil && groupID > 0 {
		var serverGroup struct {
			Name string `gorm:"column:name"`
		}
		if err := s.db.Table("server_groups").
			Select("name").
			Where("id = ?", groupID).
			First(&serverGroup).Error; err == nil {
			plan.DefaultServerGroupName = serverGroup.Name
		}
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
		logger.Error("Failed to get subscription plan by code", logger.String("code", code), logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}

	// Fill server group name
	if groupID, err := plan.GetDefaultServerGroupID(); err == nil && groupID > 0 {
		var serverGroup struct {
			Name string `gorm:"column:name"`
		}
		if err := s.db.Table("server_groups").
			Select("name").
			Where("id = ?", groupID).
			First(&serverGroup).Error; err == nil {
			plan.DefaultServerGroupName = serverGroup.Name
		}
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
		logger.Error("Failed to count subscription plans", logger.ErrorField(err))
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
		logger.Error("Failed to get subscription plans", logger.ErrorField(err))
		return nil, 0, fmt.Errorf("failed to get subscription plans: %w", err)
	}

	// Fill server group names for each plan
	for _, plan := range plans {
		if groupID, err := plan.GetDefaultServerGroupID(); err == nil && groupID > 0 {
			var serverGroup struct {
				Name string `gorm:"column:name"`
			}
			if err := s.db.Table("server_groups").
				Select("name").
				Where("id = ?", groupID).
				First(&serverGroup).Error; err == nil {
				plan.DefaultServerGroupName = serverGroup.Name
			}
		}
	}

	return plans, totalCount, nil
}

// GetVisibleSubscriptionPlans gets visible and active subscription plans for public display
func (s *SubscriptionPlanService) GetVisibleSubscriptionPlans(ctx context.Context, currency string) ([]*entities.SubscriptionPlan, error) {
	query := s.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where("status = ? AND is_visible = ?", constants.SubscriptionPlanStatusActive, true)

	if currency != "" {
		query = query.Where("currency = ?", strings.ToUpper(currency))
	}

	var plans []*entities.SubscriptionPlan
	if err := query.Order("sort_order ASC, created_at ASC").Find(&plans).Error; err != nil {
		logger.Error("Failed to get public subscription plans", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get public subscription plans: %w", err)
	}

	// Fill server group names for each plan
	for _, plan := range plans {
		if groupID, err := plan.GetDefaultServerGroupID(); err == nil && groupID > 0 {
			var serverGroup struct {
				Name string `gorm:"column:name"`
			}
			if err := s.db.Table("server_groups").
				Select("name").
				Where("id = ?", groupID).
				First(&serverGroup).Error; err == nil {
				plan.DefaultServerGroupName = serverGroup.Name
			}
		}
	}

	return plans, nil
}

// GetPopularSubscriptionPlans gets popular subscription plans
func (s *SubscriptionPlanService) GetPopularSubscriptionPlans(ctx context.Context, limit int) ([]*entities.SubscriptionPlan, error) {
	query := s.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where("status = ? AND is_visible = ? AND is_popular = ?", constants.SubscriptionPlanStatusActive, true, true)

	if limit > 0 {
		query = query.Limit(limit)
	}

	var plans []*entities.SubscriptionPlan
	if err := query.Order("sort_order ASC, created_at ASC").Find(&plans).Error; err != nil {
		logger.Error("Failed to get popular subscription plans", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get popular subscription plans: %w", err)
	}

	// Fill server group names for each plan
	for _, plan := range plans {
		if groupID, err := plan.GetDefaultServerGroupID(); err == nil && groupID > 0 {
			var serverGroup struct {
				Name string `gorm:"column:name"`
			}
			if err := s.db.Table("server_groups").
				Select("name").
				Where("id = ?", groupID).
				First(&serverGroup).Error; err == nil {
				plan.DefaultServerGroupName = serverGroup.Name
			}
		}
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
	updates := make(map[string]any)

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

	// Handle default server group IDs update
	if req.DefaultServerGroupIDs != nil {
		if err := plan.SetDefaultServerGroupIDs(*req.DefaultServerGroupIDs); err != nil {
			return nil, fmt.Errorf("failed to set default server group IDs: %w", err)
		}
		updates["default_server_groups"] = plan.DefaultServerGroupIDs
		logger.Info("Updated default server group IDs for plan",
			logger.Uint("plan_id", planID),
			logger.Any("server_group_ids", *req.DefaultServerGroupIDs))
	}

	// Update the plan
	if err := s.db.WithContext(ctx).Model(plan).Updates(updates).Error; err != nil {
		logger.Error("Failed to update subscription plan", logger.Uint("planID", uint(planID)))
		return nil, fmt.Errorf("failed to update subscription plan: %w", err)
	}

	// Reload the plan
	if err := s.db.WithContext(ctx).First(plan, planID).Error; err != nil {
		logger.Error("Failed to reload updated subscription plan", logger.Uint("planID", uint(planID)))
		return nil, fmt.Errorf("failed to reload updated subscription plan: %w", err)
	}

	// Fill server group name
	if groupID, err := plan.GetDefaultServerGroupID(); err == nil && groupID > 0 {
		var serverGroup struct {
			Name string `gorm:"column:name"`
		}
		if err := s.db.Table("server_groups").
			Select("name").
			Where("id = ?", groupID).
			First(&serverGroup).Error; err == nil {
			plan.DefaultServerGroupName = serverGroup.Name
		}
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
		Where("subscription_plan_id = ? AND status = ?", planID, constants.UserSubscriptionStatusActive).
		Count(&activeSubscriptionCount).Error; err != nil {
		logger.Error("Failed to check active subscriptions", logger.Uint("planID", uint(planID)))
		return fmt.Errorf("failed to check active subscriptions: %w", err)
	}

	if activeSubscriptionCount > 0 {
		return fmt.Errorf("cannot delete plan with %d active subscriptions", activeSubscriptionCount)
	}

	// Soft delete the plan
	if err := s.db.WithContext(ctx).Delete(plan).Error; err != nil {
		logger.Error("Failed to delete subscription plan", logger.Uint("planID", uint(planID)))
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
	case constants.SubscriptionPlanStatusActive:
		newStatus = constants.SubscriptionPlanStatusInactive
	case constants.SubscriptionPlanStatusInactive:
		newStatus = constants.SubscriptionPlanStatusActive
	default:
		return nil, fmt.Errorf("cannot toggle status of archived plan")
	}

	// Update the plan
	if err := s.db.WithContext(ctx).Model(plan).Update("status", newStatus).Error; err != nil {
		logger.Error("Failed to toggle subscription plan status", logger.Uint("planID", uint(planID)))
		return nil, fmt.Errorf("failed to toggle subscription plan status: %w", err)
	}

	// Reload the plan
	if err := s.db.WithContext(ctx).First(plan, planID).Error; err != nil {
		logger.Error("Failed to reload toggled subscription plan", logger.Uint("planID", uint(planID)))
		return nil, fmt.Errorf("failed to reload toggled subscription plan: %w", err)
	}

	// Fill server group name
	if groupID, err := plan.GetDefaultServerGroupID(); err == nil && groupID > 0 {
		var serverGroup struct {
			Name string `gorm:"column:name"`
		}
		if err := s.db.Table("server_groups").
			Select("name").
			Where("id = ?", groupID).
			First(&serverGroup).Error; err == nil {
			plan.DefaultServerGroupName = serverGroup.Name
		}
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
		Where("subscription_plan_id = ? AND status = ?", planID, constants.UserSubscriptionStatusActive).
		Count(&activeSubscriptionCount).Error; err != nil {
		logger.Error("Failed to check active subscriptions for archiving", logger.Uint("planID", uint(planID)))
		return fmt.Errorf("failed to check active subscriptions: %w", err)
	}

	if activeSubscriptionCount > 0 {
		return fmt.Errorf("cannot archive plan with %d active subscriptions", activeSubscriptionCount)
	}

	// Archive the plan
	if err := s.db.WithContext(ctx).Model(plan).Updates(map[string]any{
		"status":     constants.SubscriptionPlanStatusArchived,
		"is_visible": false,
	}).Error; err != nil {
		logger.Error("Failed to archive subscription plan", logger.Uint("planID", uint(planID)))
		return fmt.Errorf("failed to archive subscription plan: %w", err)
	}

	logger.Info("Subscription plan archived successfully", logger.Uint("plan_id", planID))

	return nil
}
