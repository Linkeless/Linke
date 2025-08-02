package implementations

import (
	"context"
	"fmt"
	"strings"

	"linke/internal/shared/logger"
	"linke/internal/domains/subscription/entities"

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

// CreateSubscriptionPlanRequest represents the request to create a subscription plan
type CreateSubscriptionPlanRequest struct {
	Name            string  `json:"name" binding:"required,min=1,max=100" example:"Premium Plan"`
	Code            string  `json:"code" binding:"required,min=1,max=50" example:"premium-monthly"`
	Description     string  `json:"description" binding:"max=1000" example:"Premium features with monthly billing"`
	Price           float64 `json:"price" binding:"required,min=0" example:"29.99"`
	Currency        string  `json:"currency" binding:"required,len=3" example:"USD"`
	BillingCycle    string  `json:"billing_cycle" binding:"required,oneof=monthly yearly lifetime" example:"monthly"`
	BillingInterval int     `json:"billing_interval" binding:"min=1,max=12" example:"1"`
	TrialPeriodDays int     `json:"trial_period_days" binding:"min=0,max=365" example:"7"`
	Features        string  `json:"features,omitempty" example:"{\"max_projects\": 10, \"storage_gb\": 100}"`
	Limits          string  `json:"limits,omitempty" example:"{\"api_calls_per_month\": 10000}"`
	IsVisible       *bool   `json:"is_visible,omitempty" example:"true"`
	SortOrder       int     `json:"sort_order,omitempty" example:"1"`
	IsPopular       *bool   `json:"is_popular,omitempty" example:"false"`
	IsRecommended   *bool   `json:"is_recommended,omitempty" example:"true"`
	SetupFee        float64 `json:"setup_fee,omitempty" example:"0"`
	CancellationFee float64 `json:"cancellation_fee,omitempty" example:"0"`
	
	// Traffic Configuration (Required)
	TrafficLimit      int64  `json:"traffic_limit" binding:"required,min=0" example:"107374182400"` // Traffic limit in bytes (0 = unlimited)
	TrafficResetCycle string `json:"traffic_reset_cycle" binding:"required,oneof=monthly never" example:"monthly"` // Traffic reset cycle
}

// UpdateSubscriptionPlanRequest represents the request to update a subscription plan
type UpdateSubscriptionPlanRequest struct {
	Name            *string  `json:"name,omitempty" binding:"omitempty,min=1,max=100" example:"Premium Plan Updated"`
	Description     *string  `json:"description,omitempty" binding:"omitempty,max=1000" example:"Updated description"`
	Price           *float64 `json:"price,omitempty" binding:"omitempty,min=0" example:"39.99"`
	TrialPeriodDays *int     `json:"trial_period_days,omitempty" binding:"omitempty,min=0,max=365" example:"14"`
	Features        *string  `json:"features,omitempty" example:"{\"max_projects\": 20}"`
	Limits          *string  `json:"limits,omitempty" example:"{\"api_calls_per_month\": 20000}"`
	Status          *string  `json:"status,omitempty" binding:"omitempty,oneof=active inactive archived" example:"active"`
	IsVisible       *bool    `json:"is_visible,omitempty" example:"true"`
	SortOrder       *int     `json:"sort_order,omitempty" example:"2"`
	IsPopular       *bool    `json:"is_popular,omitempty" example:"true"`
	IsRecommended   *bool    `json:"is_recommended,omitempty" example:"false"`
	SetupFee        *float64 `json:"setup_fee,omitempty" example:"10"`
	CancellationFee *float64 `json:"cancellation_fee,omitempty" example:"25"`
	
	// Traffic Configuration
	TrafficLimit      *int64  `json:"traffic_limit,omitempty" binding:"omitempty,min=0" example:"107374182400"` // Traffic limit in bytes
	TrafficResetCycle *string `json:"traffic_reset_cycle,omitempty" binding:"omitempty,oneof=monthly never" example:"monthly"` // Traffic reset cycle
}

// GetSubscriptionPlansRequest represents the request to get subscription plans
type GetSubscriptionPlansRequest struct {
	Status      string `form:"status" binding:"omitempty,oneof=active inactive archived" example:"active"`
	Currency    string `form:"currency" binding:"omitempty,len=3" example:"USD"`
	Visible     *bool  `form:"visible" example:"true"`
	Popular     *bool  `form:"popular" example:"false"`
	Recommended *bool  `form:"recommended" example:"true"`
	Limit       int    `form:"limit" binding:"omitempty,min=1,max=100" example:"10"`
	Offset      int    `form:"offset" binding:"omitempty,min=0" example:"0"`
}

// CreateSubscriptionPlan creates a new subscription plan
func (s *SubscriptionPlanService) CreateSubscriptionPlan(ctx context.Context, creatorID uint, req *CreateSubscriptionPlanRequest) (*entities.SubscriptionPlan, error) {
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
func (s *SubscriptionPlanService) GetSubscriptionPlans(ctx context.Context, req *GetSubscriptionPlansRequest) ([]*entities.SubscriptionPlan, int64, error) {
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

// GetPublicSubscriptionPlans gets visible and active subscription plans for public display
func (s *SubscriptionPlanService) GetPublicSubscriptionPlans(ctx context.Context, currency string) ([]*entities.SubscriptionPlan, error) {
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

// UpdateSubscriptionPlan updates a subscription plan
func (s *SubscriptionPlanService) UpdateSubscriptionPlan(ctx context.Context, planID uint, req *UpdateSubscriptionPlanRequest) (*entities.SubscriptionPlan, error) {
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