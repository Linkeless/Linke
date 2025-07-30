package service

import (
	"context"
	"fmt"
	"time"

	"linke/internal/logger"
	"linke/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserSubscriptionService struct {
	db                      *gorm.DB
	subscriptionPlanService *SubscriptionPlanService
}

func NewUserSubscriptionService(db *gorm.DB, subscriptionPlanService *SubscriptionPlanService) *UserSubscriptionService {
	return &UserSubscriptionService{
		db:                      db,
		subscriptionPlanService: subscriptionPlanService,
	}
}

// CreateSubscriptionRequest represents the request to create a user subscription
type CreateSubscriptionRequest struct {
	UserID             uint   `json:"user_id" binding:"required" example:"1"`
	SubscriptionPlanID uint   `json:"subscription_plan_id" binding:"required" example:"1"`
	StartDate          string `json:"start_date,omitempty" binding:"omitempty" example:"2024-01-01T00:00:00Z"`
	UseTrial           bool   `json:"use_trial,omitempty" example:"true"`
	ServerGroupIDs     []uint `json:"server_group_ids,omitempty"`
	
	// Custom Traffic Configuration (optional, overrides plan defaults)
	CustomTrafficLimit      *int64  `json:"custom_traffic_limit,omitempty" example:"107374182400"`     // Custom traffic limit in bytes
	CustomTrafficResetCycle *string `json:"custom_traffic_reset_cycle,omitempty" example:"monthly"`   // Custom reset cycle
	DisableTrafficLimit     *bool   `json:"disable_traffic_limit,omitempty" example:"false"`          // Disable traffic limit for this subscription
}

// UpdateSubscriptionRequest represents the request to update a user subscription
type UpdateSubscriptionRequest struct {
	Status             *string    `json:"status,omitempty" binding:"omitempty,oneof=active paused cancelled expired trial" example:"active"`
	EndDate            *time.Time `json:"end_date,omitempty" example:"2024-12-31T23:59:59Z"`
	CancellationReason *string    `json:"cancellation_reason,omitempty" binding:"omitempty,max=255" example:"User request"`
	CancelAtPeriodEnd  *bool      `json:"cancel_at_period_end,omitempty" example:"true"`
	AutoRenew          *bool      `json:"auto_renew,omitempty" example:"true"`
	Notes              *string    `json:"notes,omitempty" binding:"omitempty,max=1000" example:"Customer feedback notes"`
	ServerGroupIDs     *[]uint    `json:"server_group_ids,omitempty"`
}

// GetUserSubscriptionsRequest represents the request to get user subscriptions
type GetUserSubscriptionsRequest struct {
	UserID uint   `form:"user_id,omitempty" example:"1"`
	Status string `form:"status,omitempty" binding:"omitempty,oneof=active paused cancelled expired trial" example:"active"`
	Limit  int    `form:"limit,omitempty" binding:"omitempty,min=1,max=100" example:"10"`
	Offset int    `form:"offset,omitempty" binding:"omitempty,min=0" example:"0"`
}

// CreateUserSubscription creates a new user subscription
func (s *UserSubscriptionService) CreateUserSubscription(ctx context.Context, req *CreateSubscriptionRequest) (*model.UserSubscription, error) {
	// Get the subscription plan
	plan, err := s.subscriptionPlanService.GetSubscriptionPlan(ctx, req.SubscriptionPlanID)
	if err != nil {
		return nil, fmt.Errorf("invalid subscription plan: %w", err)
	}

	if !plan.IsAvailableForPurchase() {
		return nil, fmt.Errorf("subscription plan is not available for purchase")
	}

	// Check if user already has an active subscription to this plan
	var existingSubscription model.UserSubscription
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND subscription_plan_id = ? AND status IN (?)", 
			req.UserID, req.SubscriptionPlanID, []string{model.UserSubscriptionStatusActive, model.UserSubscriptionStatusTrial}).
		First(&existingSubscription).Error; err == nil {
		return nil, fmt.Errorf("user already has an active subscription to this plan")
	} else if err != gorm.ErrRecordNotFound {
		logger.Error("Failed to check existing subscription", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to check existing subscription: %w", err)
	}

	// Parse start date or use current time
	var startDate time.Time
	if req.StartDate != "" {
		if parsedDate, err := time.Parse(time.RFC3339, req.StartDate); err == nil {
			startDate = parsedDate
		} else {
			return nil, fmt.Errorf("invalid start date format: %w", err)
		}
	} else {
		startDate = time.Now()
	}

	// Calculate end date and billing dates
	var endDate *time.Time
	var currentPeriodStart, currentPeriodEnd, nextBillingDate *time.Time
	var trialEndDate *time.Time

	// Set trial end date if applicable
	if req.UseTrial && plan.TrialPeriodDays > 0 {
		trialEnd := startDate.AddDate(0, 0, plan.TrialPeriodDays)
		trialEndDate = &trialEnd
	}

	// Calculate billing dates based on plan
	if plan.BillingCycle != model.BillingCycleLifetime {
		periodStart := startDate
		if trialEndDate != nil {
			periodStart = *trialEndDate
		}
		currentPeriodStart = &periodStart

		// Calculate period end based on billing cycle and interval
		var periodEnd time.Time
		switch plan.BillingCycle {
		case model.BillingCycleMonthly:
			periodEnd = periodStart.AddDate(0, plan.BillingInterval, 0)
		case model.BillingCycleYearly:
			periodEnd = periodStart.AddDate(plan.BillingInterval, 0, 0)
		}
		currentPeriodEnd = &periodEnd
		nextBillingDate = &periodEnd
	}
	// For lifetime subscriptions, endDate remains nil

	// Determine initial status
	status := model.UserSubscriptionStatusActive
	if trialEndDate != nil {
		status = model.UserSubscriptionStatusTrial
	}

	// Generate UUID for server access
	subscriptionUUID := uuid.New().String()

	// Determine traffic configuration (custom overrides plan defaults)
	var trafficLimit int64
	var trafficResetCycle string
	var trafficResetDate *time.Time
	
	// Check if traffic limit is disabled for this subscription
	if req.DisableTrafficLimit != nil && *req.DisableTrafficLimit {
		trafficLimit = 0 // Unlimited
		trafficResetCycle = model.TrafficResetCycleNever
	} else {
		// Use custom traffic limit if provided, otherwise use plan default
		if req.CustomTrafficLimit != nil {
			trafficLimit = *req.CustomTrafficLimit
		} else {
			trafficLimit = plan.TrafficLimit
		}
		
		// Use custom reset cycle if provided, otherwise use plan default
		if req.CustomTrafficResetCycle != nil {
			trafficResetCycle = *req.CustomTrafficResetCycle
		} else {
			trafficResetCycle = plan.TrafficResetCycle
		}
		
		// Calculate traffic reset date based on configuration
		if trafficResetCycle != model.TrafficResetCycleNever && trafficLimit > 0 {
			var resetDate time.Time
			switch trafficResetCycle {
			case model.TrafficResetCycleMonthly:
				// Set traffic reset to the next month's start date
				resetDate = startDate.AddDate(0, 1, 0)
				// Reset to the first day of the month
				resetDate = time.Date(resetDate.Year(), resetDate.Month(), 1, 0, 0, 0, 0, resetDate.Location())
			default:
				// For other cycles, use the billing period end
				if currentPeriodEnd != nil {
					resetDate = *currentPeriodEnd
				}
			}
			if !resetDate.IsZero() {
				trafficResetDate = &resetDate
			}
		}
	}

	// Create the subscription
	subscription := &model.UserSubscription{
		UserID:             req.UserID,
		SubscriptionPlanID: req.SubscriptionPlanID,
		UUID:               subscriptionUUID,
		Status:             status,
		StartDate:          startDate,
		EndDate:            endDate,
		TrialEndDate:       trialEndDate,
		CurrentPeriodStart: currentPeriodStart,
		CurrentPeriodEnd:   currentPeriodEnd,
		NextBillingDate:    nextBillingDate,
		Price:              plan.Price,
		Currency:           plan.Currency,
		BillingCycle:       plan.BillingCycle,
		BillingInterval:    plan.BillingInterval,
		CancelAtPeriodEnd:  false,
		
		// Apply calculated traffic configuration
		TrafficLimit:      trafficLimit,
		TrafficUsed:       0, // Start with zero usage
		TrafficResetDate:  trafficResetDate,
		TrafficResetCycle: trafficResetCycle,
		TrafficSuspended:  false, // Start as not suspended
	}
	
	// Set server group IDs if provided
	if len(req.ServerGroupIDs) > 0 {
		if err := subscription.SetServerGroupIDs(req.ServerGroupIDs); err != nil {
			return nil, fmt.Errorf("failed to set server group IDs: %w", err)
		}
	}

	if err := s.db.WithContext(ctx).Create(subscription).Error; err != nil {
		logger.Error("Failed to create user subscription", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to create user subscription: %w", err)
	}

	logger.Info("User subscription created successfully", 
		logger.Uint("subscription_id", subscription.ID),
		logger.Uint("user_id", req.UserID),
		logger.Uint("plan_id", req.SubscriptionPlanID),
		logger.String("uuid", subscription.UUID))

	return subscription, nil
}

// GetUserSubscription gets a user subscription by ID
func (s *UserSubscriptionService) GetUserSubscription(ctx context.Context, subscriptionID uint) (*model.UserSubscription, error) {
	var subscription model.UserSubscription
	if err := s.db.WithContext(ctx).First(&subscription, subscriptionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription not found")
		}
		logger.Error("Failed to get user subscription", logger.Error2("error", err), logger.Uint("subscription_id", subscriptionID))
		return nil, fmt.Errorf("failed to get user subscription: %w", err)
	}

	return &subscription, nil
}

// GetUserSubscriptionWithRelations gets a user subscription with related data
func (s *UserSubscriptionService) GetUserSubscriptionWithRelations(ctx context.Context, subscriptionID uint) (*model.UserSubscription, error) {
	subscription, err := s.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	// Load related user
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, subscription.UserID).Error; err == nil {
		subscription.User = &user
	}

	// Load related subscription plan
	var plan model.SubscriptionPlan
	if err := s.db.WithContext(ctx).First(&plan, subscription.SubscriptionPlanID).Error; err == nil {
		subscription.SubscriptionPlan = &plan
	}

	return subscription, nil
}

// GetUserSubscriptions gets user subscriptions with filtering and pagination
func (s *UserSubscriptionService) GetUserSubscriptions(ctx context.Context, req *GetUserSubscriptionsRequest) ([]*model.UserSubscription, int64, error) {
	query := s.db.WithContext(ctx).Model(&model.UserSubscription{})

	// Apply filters
	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		logger.Error("Failed to count user subscriptions", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count user subscriptions: %w", err)
	}

	// Apply pagination and ordering
	query = query.Order("created_at DESC")

	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	var subscriptions []*model.UserSubscription
	if err := query.Find(&subscriptions).Error; err != nil {
		logger.Error("Failed to get user subscriptions", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to get user subscriptions: %w", err)
	}

	return subscriptions, totalCount, nil
}

// GetUserSubscriptionsWithRelations gets user subscriptions with related user and plan data using single JOIN query
func (s *UserSubscriptionService) GetUserSubscriptionsWithRelations(ctx context.Context, req *GetUserSubscriptionsRequest) ([]*model.UserSubscription, int64, error) {
	// Build base query for counting
	countQuery := s.db.WithContext(ctx).Model(&model.UserSubscription{})

	// Apply filters for count query
	if req.UserID > 0 {
		countQuery = countQuery.Where("user_subscriptions.user_id = ?", req.UserID)
	}

	if req.Status != "" {
		countQuery = countQuery.Where("user_subscriptions.status = ?", req.Status)
	}

	// Get total count
	var totalCount int64
	if err := countQuery.Count(&totalCount).Error; err != nil {
		logger.Error("Failed to count user subscriptions", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count user subscriptions: %w", err)
	}

	// Build single JOIN query to get all data at once
	query := s.db.WithContext(ctx).
		Select(`
			user_subscriptions.*,
			users.id as user_id_ref,
			users.email as user_email,
			users.username as user_username,
			users.status as user_status,
			users.role as user_role,
			users.created_at as user_created_at,
			users.updated_at as user_updated_at,
			subscription_plans.id as plan_id_ref,
			subscription_plans.name as plan_name,
			subscription_plans.description as plan_description,
			subscription_plans.price as plan_price,
			subscription_plans.currency as plan_currency,
			subscription_plans.billing_cycle as plan_billing_cycle,
			subscription_plans.billing_interval as plan_billing_interval,
			subscription_plans.status as plan_status,
			subscription_plans.is_recommended as plan_is_recommended,
			subscription_plans.created_at as plan_created_at,
			subscription_plans.updated_at as plan_updated_at
		`).
		Joins("LEFT JOIN users ON users.id = user_subscriptions.user_id AND users.deleted_at IS NULL").
		Joins("LEFT JOIN subscription_plans ON subscription_plans.id = user_subscriptions.subscription_plan_id AND subscription_plans.deleted_at IS NULL").
		Model(&model.UserSubscription{})

	// Apply same filters for main query
	if req.UserID > 0 {
		query = query.Where("user_subscriptions.user_id = ?", req.UserID)
	}

	if req.Status != "" {
		query = query.Where("user_subscriptions.status = ?", req.Status)
	}

	// Apply pagination and ordering
	query = query.Order("user_subscriptions.created_at DESC")

	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	// Custom struct to capture all fields from JOIN
	type SubscriptionWithRelations struct {
		model.UserSubscription
		// User fields
		UserIDRef       *uint      `gorm:"column:user_id_ref"`
		UserEmail       *string    `gorm:"column:user_email"`
		UserUsername    *string    `gorm:"column:user_username"`
		UserStatus      *string    `gorm:"column:user_status"`
		UserRole        *string    `gorm:"column:user_role"`
		UserCreatedAt   *time.Time `gorm:"column:user_created_at"`
		UserUpdatedAt   *time.Time `gorm:"column:user_updated_at"`
		// Plan fields
		PlanIDRef           *uint      `gorm:"column:plan_id_ref"`
		PlanName            *string    `gorm:"column:plan_name"`
		PlanDescription     *string    `gorm:"column:plan_description"`
		PlanPrice           *float64   `gorm:"column:plan_price"`
		PlanCurrency        *string    `gorm:"column:plan_currency"`
		PlanBillingCycle    *string    `gorm:"column:plan_billing_cycle"`
		PlanBillingInterval *int       `gorm:"column:plan_billing_interval"`
		PlanStatus          *string    `gorm:"column:plan_status"`
		PlanIsRecommended   *bool      `gorm:"column:plan_is_recommended"`
		PlanCreatedAt       *time.Time `gorm:"column:plan_created_at"`
		PlanUpdatedAt       *time.Time `gorm:"column:plan_updated_at"`
	}

	var results []SubscriptionWithRelations
	if err := query.Find(&results).Error; err != nil {
		logger.Error("Failed to get user subscriptions with relations", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to get user subscriptions with relations: %w", err)
	}

	// Convert results to UserSubscription with populated relations
	subscriptions := make([]*model.UserSubscription, len(results))
	for i, result := range results {
		subscription := &result.UserSubscription
		
		// Populate User if data exists
		if result.UserIDRef != nil {
			subscription.User = &model.User{
				ID:        *result.UserIDRef,
				Email:     *result.UserEmail,
				Username:  getStringValue(result.UserUsername),
				Status:    getStringValue(result.UserStatus),
				Role:      getStringValue(result.UserRole),
				CreatedAt: *result.UserCreatedAt,
				UpdatedAt: *result.UserUpdatedAt,
			}
		}

		// Populate SubscriptionPlan if data exists
		if result.PlanIDRef != nil {
			subscription.SubscriptionPlan = &model.SubscriptionPlan{
				ID:              *result.PlanIDRef,
				Name:            *result.PlanName,
				Description:     getStringValue(result.PlanDescription),
				Price:           *result.PlanPrice,
				Currency:        *result.PlanCurrency,
				BillingCycle:    *result.PlanBillingCycle,
				BillingInterval: *result.PlanBillingInterval,
				Status:          *result.PlanStatus,
				IsRecommended:   getBoolValue(result.PlanIsRecommended),
				CreatedAt:       *result.PlanCreatedAt,
				UpdatedAt:       *result.PlanUpdatedAt,
			}
		}

		subscriptions[i] = subscription
	}

	logger.Info("User subscriptions with relations retrieved successfully", 
		logger.Int("count", len(subscriptions)),
		logger.Int64("total", totalCount))

	return subscriptions, totalCount, nil
}

// Helper functions to safely get values from pointers
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func getBoolValue(ptr *bool) bool {
	if ptr == nil {
		return false
	}
	return *ptr
}

// GetActiveUserSubscription gets the active subscription for a user and plan
func (s *UserSubscriptionService) GetActiveUserSubscription(ctx context.Context, userID, planID uint) (*model.UserSubscription, error) {
	var subscription model.UserSubscription
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND subscription_plan_id = ? AND status IN (?)", 
			userID, planID, []string{model.UserSubscriptionStatusActive, model.UserSubscriptionStatusTrial}).
		First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no active subscription found")
		}
		logger.Error("Failed to get active user subscription", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get active user subscription: %w", err)
	}

	return &subscription, nil
}

// UpdateUserSubscription updates a user subscription
func (s *UserSubscriptionService) UpdateUserSubscription(ctx context.Context, subscriptionID uint, req *UpdateSubscriptionRequest) (*model.UserSubscription, error) {
	// Get existing subscription with row lock to prevent race conditions
	var subscription model.UserSubscription
	if err := s.db.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").
		First(&subscription, subscriptionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	// SECURITY: Validate status transitions to prevent unauthorized changes
	if req.Status != nil {
		newStatus := *req.Status
		currentStatus := subscription.Status
		
		// Define allowed status transitions
		allowedTransitions := map[string][]string{
			model.UserSubscriptionStatusActive: {
				model.UserSubscriptionStatusPaused,
				model.UserSubscriptionStatusCancelled,
			},
			model.UserSubscriptionStatusPaused: {
				model.UserSubscriptionStatusActive,
				model.UserSubscriptionStatusCancelled,
			},
			model.UserSubscriptionStatusTrial: {
				model.UserSubscriptionStatusActive,
				model.UserSubscriptionStatusCancelled,
				model.UserSubscriptionStatusExpired,
			},
			model.UserSubscriptionStatusCancelled: {}, // No transitions allowed from cancelled
			model.UserSubscriptionStatusExpired: {
				model.UserSubscriptionStatusActive, // Only allow reactivation
			},
		}
		
		// Check if transition is allowed
		if allowedStatuses, exists := allowedTransitions[currentStatus]; exists {
			isAllowed := false
			for _, allowedStatus := range allowedStatuses {
				if newStatus == allowedStatus {
					isAllowed = true
					break
				}
			}
			if !isAllowed {
				return nil, fmt.Errorf("invalid status transition from %s to %s", currentStatus, newStatus)
			}
		} else {
			return nil, fmt.Errorf("unknown current status: %s", currentStatus)
		}
	}

	// Prepare updates
	updates := make(map[string]interface{})

	if req.Status != nil {
		updates["status"] = *req.Status

		// If cancelling, set cancellation date
		if *req.Status == model.UserSubscriptionStatusCancelled && subscription.CancelledAt == nil {
			now := time.Now()
			updates["cancelled_at"] = &now
		}
		
		// If reactivating from expired, clear cancellation data
		if *req.Status == model.UserSubscriptionStatusActive && subscription.Status == model.UserSubscriptionStatusExpired {
			updates["cancelled_at"] = nil
			updates["cancellation_reason"] = ""
		}
	}

	if req.EndDate != nil {
		updates["end_date"] = *req.EndDate
	}

	if req.CancellationReason != nil {
		updates["cancellation_reason"] = *req.CancellationReason
	}

	if req.CancelAtPeriodEnd != nil {
		updates["cancel_at_period_end"] = *req.CancelAtPeriodEnd
	}

	if req.AutoRenew != nil {
		updates["auto_renew"] = *req.AutoRenew
		// Reset renewal attempts when auto-renew is re-enabled
		if *req.AutoRenew {
			updates["renewal_attempts"] = 0
			updates["last_renewal_failed"] = nil
			updates["renewal_fail_reason"] = ""
		}
	}

	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	// Handle server group IDs update
	if req.ServerGroupIDs != nil {
		if err := subscription.SetServerGroupIDs(*req.ServerGroupIDs); err != nil {
			return nil, fmt.Errorf("failed to set server group IDs: %w", err)
		}
		updates["server_group_ids"] = subscription.ServerGroupIDs
	}

	// Update the subscription
	if err := s.db.WithContext(ctx).Model(subscription).Updates(updates).Error; err != nil {
		logger.Error("Failed to update user subscription", logger.Error2("error", err), logger.Uint("subscription_id", subscriptionID))
		return nil, fmt.Errorf("failed to update user subscription: %w", err)
	}

	// Reload the subscription
	if err := s.db.WithContext(ctx).First(&subscription, subscriptionID).Error; err != nil {
		logger.Error("Failed to reload updated user subscription", logger.Error2("error", err), logger.Uint("subscription_id", subscriptionID))
		return nil, fmt.Errorf("failed to reload updated user subscription: %w", err)
	}

	logger.Info("User subscription updated successfully", logger.Uint("subscription_id", subscription.ID))

	return &subscription, nil
}

// CancelUserSubscription cancels a user subscription
func (s *UserSubscriptionService) CancelUserSubscription(ctx context.Context, subscriptionID uint, reason string, immediately bool) (*model.UserSubscription, error) {
	subscription, err := s.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	if subscription.Status == model.UserSubscriptionStatusCancelled {
		return nil, fmt.Errorf("subscription is already cancelled")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"cancelled_at":        &now,
		"cancellation_reason": reason,
	}

	if immediately {
		updates["status"] = model.UserSubscriptionStatusCancelled
		updates["cancel_at_period_end"] = false
	} else {
		updates["cancel_at_period_end"] = true
	}

	if err := s.db.WithContext(ctx).Model(subscription).Updates(updates).Error; err != nil {
		logger.Error("Failed to cancel user subscription", logger.Error2("error", err), logger.Uint("subscription_id", subscriptionID))
		return nil, fmt.Errorf("failed to cancel user subscription: %w", err)
	}

	// Reload the subscription
	if err := s.db.WithContext(ctx).First(subscription, subscriptionID).Error; err != nil {
		logger.Error("Failed to reload cancelled user subscription", logger.Error2("error", err), logger.Uint("subscription_id", subscriptionID))
		return nil, fmt.Errorf("failed to reload cancelled user subscription: %w", err)
	}

	logger.Info("User subscription cancelled successfully", 
		logger.Uint("subscription_id", subscription.ID),
		logger.Any("immediately", immediately))

	return subscription, nil
}

// RenewUserSubscription renews a user subscription for the next billing period
func (s *UserSubscriptionService) RenewUserSubscription(ctx context.Context, subscriptionID uint) (*model.UserSubscription, error) {
	subscription, err := s.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	if !subscription.ShouldRenew() {
		return nil, fmt.Errorf("subscription does not need renewal")
	}

	// Calculate new billing period
	var newPeriodStart, newPeriodEnd, newNextBillingDate time.Time

	if subscription.CurrentPeriodEnd != nil {
		newPeriodStart = *subscription.CurrentPeriodEnd
	} else {
		newPeriodStart = time.Now()
	}

	// Get the subscription plan to determine billing cycle
	plan, err := s.subscriptionPlanService.GetSubscriptionPlan(ctx, subscription.SubscriptionPlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}

	switch plan.BillingCycle {
	case model.BillingCycleMonthly:
		newPeriodEnd = newPeriodStart.AddDate(0, plan.BillingInterval, 0)
	case model.BillingCycleYearly:
		newPeriodEnd = newPeriodStart.AddDate(plan.BillingInterval, 0, 0)
	default:
		return nil, fmt.Errorf("cannot renew lifetime subscription")
	}

	newNextBillingDate = newPeriodEnd

	updates := map[string]interface{}{
		"current_period_start": &newPeriodStart,
		"current_period_end":   &newPeriodEnd,
		"next_billing_date":    &newNextBillingDate,
		"status":               model.UserSubscriptionStatusActive,
	}

	if err := s.db.WithContext(ctx).Model(subscription).Updates(updates).Error; err != nil {
		logger.Error("Failed to renew user subscription", logger.Error2("error", err), logger.Uint("subscription_id", subscriptionID))
		return nil, fmt.Errorf("failed to renew user subscription: %w", err)
	}

	// Reload the subscription
	if err := s.db.WithContext(ctx).First(subscription, subscriptionID).Error; err != nil {
		logger.Error("Failed to reload renewed user subscription", logger.Error2("error", err), logger.Uint("subscription_id", subscriptionID))
		return nil, fmt.Errorf("failed to reload renewed user subscription: %w", err)
	}

	logger.Info("User subscription renewed successfully", logger.Uint("subscription_id", subscription.ID))

	return subscription, nil
}

// DeleteUserSubscription soft deletes a user subscription
func (s *UserSubscriptionService) DeleteUserSubscription(ctx context.Context, subscriptionID uint) error {
	subscription, err := s.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	if subscription.IsActive() {
		return fmt.Errorf("cannot delete active subscription - cancel it first")
	}

	if err := s.db.WithContext(ctx).Delete(subscription).Error; err != nil {
		logger.Error("Failed to delete user subscription", logger.Error2("error", err), logger.Uint("subscription_id", subscriptionID))
		return fmt.Errorf("failed to delete user subscription: %w", err)
	}

	logger.Info("User subscription deleted successfully", logger.Uint("subscription_id", subscriptionID))

	return nil
}

// UpdateLastUsed updates the last used timestamp for a subscription
func (s *UserSubscriptionService) UpdateLastUsed(ctx context.Context, subscriptionID uint) error {
	now := time.Now()
	if err := s.db.WithContext(ctx).Model(&model.UserSubscription{}).
		Where("id = ?", subscriptionID).
		Update("last_used_at", &now).Error; err != nil {
		logger.Error("Failed to update subscription last used", logger.Error2("error", err), logger.Uint("subscription_id", subscriptionID))
		return fmt.Errorf("failed to update subscription last used: %w", err)
	}

	return nil
}

// ResetTrafficRequest represents the request to reset traffic for subscriptions
type ResetTrafficRequest struct {
	SubscriptionIDs []uint `json:"subscription_ids,omitempty"`     // Specific subscription IDs (optional)
	UserID          *uint  `json:"user_id,omitempty"`              // Reset for specific user (optional)
	ResetAll        bool   `json:"reset_all,omitempty"`            // Reset all eligible subscriptions
}

// ResetSubscriptionTraffic resets traffic usage for a specific subscription
func (s *UserSubscriptionService) ResetSubscriptionTraffic(ctx context.Context, subscriptionID uint) error {
	subscription, err := s.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	// Only reset if subscription has traffic limits and reset cycle is enabled
	if subscription.TrafficLimit == 0 || subscription.TrafficResetCycle == model.TrafficResetCycleNever {
		return fmt.Errorf("traffic reset is not enabled for this subscription")
	}

	// Calculate next reset date
	var nextResetDate *time.Time
	now := time.Now()
	
	switch subscription.TrafficResetCycle {
	case model.TrafficResetCycleMonthly:
		// Reset to the next month's start date
		nextReset := now.AddDate(0, 1, 0)
		nextReset = time.Date(nextReset.Year(), nextReset.Month(), 1, 0, 0, 0, 0, nextReset.Location())
		nextResetDate = &nextReset
	default:
		// For other cycles, calculate based on billing cycle
		if subscription.CurrentPeriodEnd != nil {
			nextReset := *subscription.CurrentPeriodEnd
			nextResetDate = &nextReset
		}
	}

	// Reset traffic usage
	updates := map[string]interface{}{
		"traffic_used":      0,
		"traffic_suspended": false,
		"updated_at":        now,
	}
	
	if nextResetDate != nil {
		updates["traffic_reset_date"] = nextResetDate
	}

	if err := s.db.WithContext(ctx).Model(subscription).Updates(updates).Error; err != nil {
		logger.Error("Failed to reset subscription traffic", 
			logger.Uint("subscription_id", subscriptionID),
			logger.Error2("error", err))
		return fmt.Errorf("failed to reset subscription traffic: %w", err)
	}

	logger.Info("Subscription traffic reset successfully", 
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID))

	return nil
}

// ResetTrafficForSubscriptions resets traffic for multiple subscriptions based on criteria
func (s *UserSubscriptionService) ResetTrafficForSubscriptions(ctx context.Context, req *ResetTrafficRequest) (int, error) {
	var resetCount int

	if req.ResetAll {
		// Reset all subscriptions that need traffic reset
		var subscriptions []model.UserSubscription
		if err := s.db.WithContext(ctx).
			Where("traffic_reset_cycle != ? AND traffic_limit > 0 AND (traffic_reset_date IS NULL OR traffic_reset_date <= ?)", 
				model.TrafficResetCycleNever, time.Now()).
			Find(&subscriptions).Error; err != nil {
			logger.Error("Failed to find subscriptions for traffic reset", logger.Error2("error", err))
			return 0, fmt.Errorf("failed to find subscriptions for traffic reset: %w", err)
		}

		for _, subscription := range subscriptions {
			if err := s.ResetSubscriptionTraffic(ctx, subscription.ID); err != nil {
				logger.Error("Failed to reset traffic for subscription", 
					logger.Uint("subscription_id", subscription.ID),
					logger.Error2("error", err))
				continue
			}
			resetCount++
		}
	} else if req.UserID != nil {
		// Reset for specific user
		var subscriptions []model.UserSubscription
		if err := s.db.WithContext(ctx).
			Where("user_id = ? AND traffic_reset_cycle != ? AND traffic_limit > 0", 
				*req.UserID, model.TrafficResetCycleNever).
			Find(&subscriptions).Error; err != nil {
			logger.Error("Failed to find user subscriptions for traffic reset", logger.Error2("error", err))
			return 0, fmt.Errorf("failed to find user subscriptions for traffic reset: %w", err)
		}

		for _, subscription := range subscriptions {
			if err := s.ResetSubscriptionTraffic(ctx, subscription.ID); err != nil {
				logger.Error("Failed to reset traffic for user subscription", 
					logger.Uint("subscription_id", subscription.ID),
					logger.Error2("error", err))
				continue
			}
			resetCount++
		}
	} else if len(req.SubscriptionIDs) > 0 {
		// Reset specific subscriptions
		for _, subscriptionID := range req.SubscriptionIDs {
			if err := s.ResetSubscriptionTraffic(ctx, subscriptionID); err != nil {
				logger.Error("Failed to reset traffic for specific subscription", 
					logger.Uint("subscription_id", subscriptionID),
					logger.Error2("error", err))
				continue
			}
			resetCount++
		}
	} else {
		return 0, fmt.Errorf("no reset criteria specified")
	}

	logger.Info("Traffic reset completed", logger.Int("reset_count", resetCount))
	return resetCount, nil
}

// UpdateSubscriptionTrafficLimit updates the traffic limit for a subscription
func (s *UserSubscriptionService) UpdateSubscriptionTrafficLimit(ctx context.Context, subscriptionID uint, newLimit int64, resetCycle string) error {
	subscription, err := s.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	// Calculate new reset date if cycle is changing
	var nextResetDate *time.Time
	if resetCycle != model.TrafficResetCycleNever && newLimit > 0 {
		now := time.Now()
		switch resetCycle {
		case model.TrafficResetCycleMonthly:
			nextReset := now.AddDate(0, 1, 0)
			nextReset = time.Date(nextReset.Year(), nextReset.Month(), 1, 0, 0, 0, 0, nextReset.Location())
			nextResetDate = &nextReset
		}
	}

	updates := map[string]interface{}{
		"traffic_limit":      newLimit,
		"traffic_reset_cycle": resetCycle,
		"updated_at":         time.Now(),
	}

	if nextResetDate != nil {
		updates["traffic_reset_date"] = nextResetDate
	} else {
		updates["traffic_reset_date"] = nil
	}

	// If increasing limit, check if we should unsuspend
	if newLimit > subscription.TrafficLimit && subscription.TrafficSuspended {
		if newLimit == 0 || subscription.TrafficUsed < newLimit {
			updates["traffic_suspended"] = false
		}
	}

	// If decreasing limit, check if we should suspend
	if newLimit > 0 && newLimit < subscription.TrafficUsed {
		updates["traffic_suspended"] = true
	}

	if err := s.db.WithContext(ctx).Model(subscription).Updates(updates).Error; err != nil {
		logger.Error("Failed to update subscription traffic limit", 
			logger.Uint("subscription_id", subscriptionID),
			logger.Error2("error", err))
		return fmt.Errorf("failed to update subscription traffic limit: %w", err)
	}

	logger.Info("Subscription traffic limit updated successfully", 
		logger.Uint("subscription_id", subscriptionID),
		logger.Int64("new_limit", newLimit),
		logger.String("reset_cycle", resetCycle))

	return nil
}

// ProcessAutoRenewals processes all eligible subscriptions for auto-renewal
func (s *UserSubscriptionService) ProcessAutoRenewals(ctx context.Context) (int, error) {
	// Find subscriptions that need auto-renewal
	var subscriptions []model.UserSubscription
	if err := s.db.WithContext(ctx).
		Where("auto_renew = ? AND status IN (?) AND next_billing_date IS NOT NULL AND next_billing_date <= ? AND renewal_attempts < 3", 
			true, []string{model.UserSubscriptionStatusActive, model.UserSubscriptionStatusTrial}, time.Now().Add(24*time.Hour)).
		Find(&subscriptions).Error; err != nil {
		logger.Error("Failed to find subscriptions for auto-renewal", logger.Error2("error", err))
		return 0, fmt.Errorf("failed to find subscriptions for auto-renewal: %w", err)
	}

	var processedCount int
	for _, subscription := range subscriptions {
		if !subscription.ShouldAutoRenew() || !subscription.CanAttemptRenewal() {
			continue
		}

		if err := s.ProcessSubscriptionAutoRenewal(ctx, subscription.ID); err != nil {
			logger.Error("Failed to process auto-renewal", 
				logger.Uint("subscription_id", subscription.ID),
				logger.Error2("error", err))
			continue
		}
		processedCount++
	}

	logger.Info("Auto-renewal processing completed", logger.Int("processed_count", processedCount))
	return processedCount, nil
}

// ProcessSubscriptionAutoRenewal processes auto-renewal for a specific subscription
func (s *UserSubscriptionService) ProcessSubscriptionAutoRenewal(ctx context.Context, subscriptionID uint) error {
	// Get subscription with row lock to prevent race conditions
	var subscription model.UserSubscription
	if err := s.db.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").
		First(&subscription, subscriptionID).Error; err != nil {
		return fmt.Errorf("failed to get subscription for auto-renewal: %w", err)
	}

	// Validate subscription is eligible for auto-renewal
	if !subscription.ShouldAutoRenew() || !subscription.CanAttemptRenewal() {
		return fmt.Errorf("subscription is not eligible for auto-renewal")
	}

	// Increment renewal attempts
	now := time.Now()
	updates := map[string]interface{}{
		"renewal_attempts": subscription.RenewalAttempts + 1,
	}

	// Try to renew the subscription
	if _, err := s.RenewUserSubscription(ctx, subscriptionID); err != nil {
		// Record renewal failure
		updates["last_renewal_failed"] = &now
		updates["renewal_fail_reason"] = err.Error()

		// If this was the final attempt, disable auto-renewal
		if subscription.RenewalAttempts+1 >= 3 {
			updates["auto_renew"] = false
			logger.Error("Auto-renewal permanently failed after 3 attempts", 
				logger.Uint("subscription_id", subscriptionID),
				logger.Error2("error", err))
		}

		// Update the subscription with failure info
		if updateErr := s.db.WithContext(ctx).Model(&subscription).Updates(updates).Error; updateErr != nil {
			logger.Error("Failed to update subscription renewal failure", 
				logger.Uint("subscription_id", subscriptionID),
				logger.Error2("error", updateErr))
		}

		return fmt.Errorf("auto-renewal failed: %w", err)
	}

	// Renewal successful - reset failure tracking
	updates["last_renewal_failed"] = nil
	updates["renewal_fail_reason"] = ""

	if err := s.db.WithContext(ctx).Model(&subscription).Updates(updates).Error; err != nil {
		logger.Error("Failed to update subscription after successful renewal", 
			logger.Uint("subscription_id", subscriptionID),
			logger.Error2("error", err))
	}

	logger.Info("Subscription auto-renewed successfully", 
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID))

	return nil
}

// GetSubscriptionsForAutoRenewal gets subscriptions that are eligible for auto-renewal
func (s *UserSubscriptionService) GetSubscriptionsForAutoRenewal(ctx context.Context) ([]*model.UserSubscription, error) {
	var subscriptions []*model.UserSubscription
	if err := s.db.WithContext(ctx).
		Where("auto_renew = ? AND status IN (?) AND next_billing_date IS NOT NULL AND next_billing_date <= ? AND renewal_attempts < 3", 
			true, []string{model.UserSubscriptionStatusActive, model.UserSubscriptionStatusTrial}, time.Now().Add(24*time.Hour)).
		Preload("User").
		Preload("SubscriptionPlan").
		Find(&subscriptions).Error; err != nil {
		logger.Error("Failed to get subscriptions for auto-renewal", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get subscriptions for auto-renewal: %w", err)
	}

	return subscriptions, nil
}

// EnableAutoRenewal enables auto-renewal for a subscription
func (s *UserSubscriptionService) EnableAutoRenewal(ctx context.Context, subscriptionID uint) error {
	updates := map[string]interface{}{
		"auto_renew":          true,
		"renewal_attempts":    0,
		"last_renewal_failed": nil,
		"renewal_fail_reason": "",
	}

	if err := s.db.WithContext(ctx).Model(&model.UserSubscription{}).
		Where("id = ?", subscriptionID).
		Updates(updates).Error; err != nil {
		logger.Error("Failed to enable auto-renewal", 
			logger.Uint("subscription_id", subscriptionID),
			logger.Error2("error", err))
		return fmt.Errorf("failed to enable auto-renewal: %w", err)
	}

	logger.Info("Auto-renewal enabled", logger.Uint("subscription_id", subscriptionID))
	return nil
}

// DisableAutoRenewal disables auto-renewal for a subscription
func (s *UserSubscriptionService) DisableAutoRenewal(ctx context.Context, subscriptionID uint) error {
	if err := s.db.WithContext(ctx).Model(&model.UserSubscription{}).
		Where("id = ?", subscriptionID).
		Update("auto_renew", false).Error; err != nil {
		logger.Error("Failed to disable auto-renewal", 
			logger.Uint("subscription_id", subscriptionID),
			logger.Error2("error", err))
		return fmt.Errorf("failed to disable auto-renewal: %w", err)
	}

	logger.Info("Auto-renewal disabled", logger.Uint("subscription_id", subscriptionID))
	return nil
}

