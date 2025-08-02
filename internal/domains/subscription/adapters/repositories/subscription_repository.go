package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

// subscriptionPlanRepository implements the SubscriptionPlanRepository interface
type subscriptionPlanRepository struct {
	db *gorm.DB
}

// subscriptionOrderRepository implements the SubscriptionOrderRepository interface
type subscriptionOrderRepository struct {
	db *gorm.DB
}

// userSubscriptionRepository implements the UserSubscriptionRepository interface
type userSubscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionPlanRepository creates a new SubscriptionPlanRepository implementation
func NewSubscriptionPlanRepository(db *gorm.DB) interfaces.SubscriptionPlanRepository {
	return &subscriptionPlanRepository{
		db: db,
	}
}

// NewSubscriptionOrderRepository creates a new SubscriptionOrderRepository implementation
func NewSubscriptionOrderRepository(db *gorm.DB) interfaces.SubscriptionOrderRepository {
	return &subscriptionOrderRepository{
		db: db,
	}
}

// NewUserSubscriptionRepository creates a new UserSubscriptionRepository implementation
func NewUserSubscriptionRepository(db *gorm.DB) interfaces.UserSubscriptionRepository {
	return &userSubscriptionRepository{
		db: db,
	}
}

// === SubscriptionPlanRepository Implementation ===

// Create creates a new subscription plan
func (r *subscriptionPlanRepository) Create(ctx context.Context, plan *entities.SubscriptionPlan) error {
	if err := r.db.WithContext(ctx).Create(plan).Error; err != nil {
		logger.Error("Failed to create subscription plan",
			logger.String("plan_code", plan.Code),
			logger.Error2("error", err),
		)
		return fmt.Errorf("failed to create subscription plan: %w", err)
	}

	logger.Debug("Subscription plan created successfully",
		logger.Uint("plan_id", plan.ID),
		logger.String("plan_code", plan.Code),
	)
	return nil
}

// GetByID retrieves a subscription plan by ID
func (r *subscriptionPlanRepository) GetByID(ctx context.Context, id uint) (*entities.SubscriptionPlan, error) {
	var plan entities.SubscriptionPlan
	if err := r.db.WithContext(ctx).First(&plan, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription plan not found")
		}
		logger.Error("Failed to get subscription plan by ID",
			logger.Uint("plan_id", id),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}
	return &plan, nil
}

// GetByCode retrieves a subscription plan by code
func (r *subscriptionPlanRepository) GetByCode(ctx context.Context, code string) (*entities.SubscriptionPlan, error) {
	var plan entities.SubscriptionPlan
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&plan).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription plan not found")
		}
		logger.Error("Failed to get subscription plan by code",
			logger.String("plan_code", code),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}
	return &plan, nil
}

// Update updates a subscription plan
func (r *subscriptionPlanRepository) Update(ctx context.Context, plan *entities.SubscriptionPlan) error {
	if err := r.db.WithContext(ctx).Save(plan).Error; err != nil {
		logger.Error("Failed to update subscription plan",
			logger.Uint("plan_id", plan.ID),
			logger.Error2("error", err),
		)
		return fmt.Errorf("failed to update subscription plan: %w", err)
	}

	logger.Debug("Subscription plan updated successfully",
		logger.Uint("plan_id", plan.ID),
		logger.String("plan_code", plan.Code),
	)
	return nil
}

// Delete performs soft delete on a subscription plan
func (r *subscriptionPlanRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entities.SubscriptionPlan{}, id)
	if result.Error != nil {
		logger.Error("Failed to delete subscription plan",
			logger.Uint("plan_id", id),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to delete subscription plan: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription plan not found")
	}

	logger.Debug("Subscription plan deleted successfully",
		logger.Uint("plan_id", id),
	)
	return nil
}

// SoftDelete performs soft delete (alias for Delete)
func (r *subscriptionPlanRepository) SoftDelete(ctx context.Context, id uint) error {
	return r.Delete(ctx, id)
}

// Restore restores a soft deleted subscription plan
func (r *subscriptionPlanRepository) Restore(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Unscoped().Model(&entities.SubscriptionPlan{}).Where("id = ?", id).Update("deleted_at", nil)
	if result.Error != nil {
		logger.Error("Failed to restore subscription plan",
			logger.Uint("plan_id", id),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to restore subscription plan: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription plan not found")
	}

	logger.Debug("Subscription plan restored successfully",
		logger.Uint("plan_id", id),
	)
	return nil
}

// HardDelete permanently deletes a subscription plan
func (r *subscriptionPlanRepository) HardDelete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Unscoped().Delete(&entities.SubscriptionPlan{}, id)
	if result.Error != nil {
		logger.Error("Failed to hard delete subscription plan",
			logger.Uint("plan_id", id),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to permanently delete subscription plan: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription plan not found")
	}

	logger.Warn("Subscription plan permanently deleted",
		logger.Uint("plan_id", id),
	)
	return nil
}

// List lists all subscription plans with pagination
func (r *subscriptionPlanRepository) List(ctx context.Context, limit, offset int) ([]*entities.SubscriptionPlan, int64, error) {
	var plans []*entities.SubscriptionPlan
	var total int64

	// Count total plans
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count subscription plans", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count subscription plans: %w", err)
	}

	// Get plans with pagination
	if err := r.db.WithContext(ctx).Order("sort_order ASC, created_at DESC").
		Limit(limit).Offset(offset).Find(&plans).Error; err != nil {
		logger.Error("Failed to list subscription plans", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to list subscription plans: %w", err)
	}

	return plans, total, nil
}

// ListActive lists active subscription plans
func (r *subscriptionPlanRepository) ListActive(ctx context.Context, limit, offset int) ([]*entities.SubscriptionPlan, int64, error) {
	var plans []*entities.SubscriptionPlan
	var total int64

	condition := "status = ?"
	args := []interface{}{entities.SubscriptionPlanStatusActive}

	// Count total active plans
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where(condition, args...).Count(&total).Error; err != nil {
		logger.Error("Failed to count active subscription plans", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count active subscription plans: %w", err)
	}

	// Get active plans with pagination
	if err := r.db.WithContext(ctx).Where(condition, args...).
		Order("sort_order ASC, created_at DESC").
		Limit(limit).Offset(offset).Find(&plans).Error; err != nil {
		logger.Error("Failed to list active subscription plans", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to list active subscription plans: %w", err)
	}

	return plans, total, nil
}

// ListVisible lists visible subscription plans
func (r *subscriptionPlanRepository) ListVisible(ctx context.Context, limit, offset int) ([]*entities.SubscriptionPlan, int64, error) {
	var plans []*entities.SubscriptionPlan
	var total int64

	condition := "is_visible = ?"
	args := []interface{}{true}

	// Count total visible plans
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where(condition, args...).Count(&total).Error; err != nil {
		logger.Error("Failed to count visible subscription plans", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count visible subscription plans: %w", err)
	}

	// Get visible plans with pagination
	if err := r.db.WithContext(ctx).Where(condition, args...).
		Order("sort_order ASC, created_at DESC").
		Limit(limit).Offset(offset).Find(&plans).Error; err != nil {
		logger.Error("Failed to list visible subscription plans", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to list visible subscription plans: %w", err)
	}

	return plans, total, nil
}

// ListDeleted lists soft deleted subscription plans
func (r *subscriptionPlanRepository) ListDeleted(ctx context.Context, limit, offset int) ([]*entities.SubscriptionPlan, int64, error) {
	var plans []*entities.SubscriptionPlan
	var total int64

	// Count total deleted plans
	if err := r.db.WithContext(ctx).Unscoped().Model(&entities.SubscriptionPlan{}).
		Where("deleted_at IS NOT NULL").Count(&total).Error; err != nil {
		logger.Error("Failed to count deleted subscription plans", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count deleted subscription plans: %w", err)
	}

	// Get deleted plans with pagination
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").
		Order("deleted_at DESC").Limit(limit).Offset(offset).Find(&plans).Error; err != nil {
		logger.Error("Failed to list deleted subscription plans", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to list deleted subscription plans: %w", err)
	}

	return plans, total, nil
}

// ListByStatus lists subscription plans by status
func (r *subscriptionPlanRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.SubscriptionPlan, int64, error) {
	var plans []*entities.SubscriptionPlan
	var total int64

	// Count total plans by status
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where("status = ?", status).Count(&total).Error; err != nil {
		logger.Error("Failed to count subscription plans by status",
			logger.String("status", status),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count subscription plans by status: %w", err)
	}

	// Get plans with pagination
	if err := r.db.WithContext(ctx).Where("status = ?", status).
		Order("sort_order ASC, created_at DESC").
		Limit(limit).Offset(offset).Find(&plans).Error; err != nil {
		logger.Error("Failed to list subscription plans by status",
			logger.String("status", status),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list subscription plans by status: %w", err)
	}

	return plans, total, nil
}

// ListByCurrency lists subscription plans by currency
func (r *subscriptionPlanRepository) ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.SubscriptionPlan, int64, error) {
	var plans []*entities.SubscriptionPlan
	var total int64

	// Count total plans by currency
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where("currency = ?", currency).Count(&total).Error; err != nil {
		logger.Error("Failed to count subscription plans by currency",
			logger.String("currency", currency),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count subscription plans by currency: %w", err)
	}

	// Get plans with pagination
	if err := r.db.WithContext(ctx).Where("currency = ?", currency).
		Order("sort_order ASC, price ASC").
		Limit(limit).Offset(offset).Find(&plans).Error; err != nil {
		logger.Error("Failed to list subscription plans by currency",
			logger.String("currency", currency),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list subscription plans by currency: %w", err)
	}

	return plans, total, nil
}

// ListByBillingCycle lists subscription plans by billing cycle
func (r *subscriptionPlanRepository) ListByBillingCycle(ctx context.Context, billingCycle string, limit, offset int) ([]*entities.SubscriptionPlan, int64, error) {
	var plans []*entities.SubscriptionPlan
	var total int64

	// Count total plans by billing cycle
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where("billing_cycle = ?", billingCycle).Count(&total).Error; err != nil {
		logger.Error("Failed to count subscription plans by billing cycle",
			logger.String("billing_cycle", billingCycle),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count subscription plans by billing cycle: %w", err)
	}

	// Get plans with pagination
	if err := r.db.WithContext(ctx).Where("billing_cycle = ?", billingCycle).
		Order("sort_order ASC, price ASC").
		Limit(limit).Offset(offset).Find(&plans).Error; err != nil {
		logger.Error("Failed to list subscription plans by billing cycle",
			logger.String("billing_cycle", billingCycle),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list subscription plans by billing cycle: %w", err)
	}

	return plans, total, nil
}

// ListPopular lists popular subscription plans
func (r *subscriptionPlanRepository) ListPopular(ctx context.Context, limit, offset int) ([]*entities.SubscriptionPlan, int64, error) {
	var plans []*entities.SubscriptionPlan
	var total int64

	condition := "is_popular = ? AND status = ? AND is_visible = ?"
	args := []interface{}{true, entities.SubscriptionPlanStatusActive, true}

	// Count total popular plans
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where(condition, args...).Count(&total).Error; err != nil {
		logger.Error("Failed to count popular subscription plans", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count popular subscription plans: %w", err)
	}

	// Get popular plans with pagination
	if err := r.db.WithContext(ctx).Where(condition, args...).
		Order("sort_order ASC, price ASC").
		Limit(limit).Offset(offset).Find(&plans).Error; err != nil {
		logger.Error("Failed to list popular subscription plans", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to list popular subscription plans: %w", err)
	}

	return plans, total, nil
}

// ListRecommended lists recommended subscription plans
func (r *subscriptionPlanRepository) ListRecommended(ctx context.Context, limit, offset int) ([]*entities.SubscriptionPlan, int64, error) {
	var plans []*entities.SubscriptionPlan
	var total int64

	condition := "is_recommended = ? AND status = ? AND is_visible = ?"
	args := []interface{}{true, entities.SubscriptionPlanStatusActive, true}

	// Count total recommended plans
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where(condition, args...).Count(&total).Error; err != nil {
		logger.Error("Failed to count recommended subscription plans", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count recommended subscription plans: %w", err)
	}

	// Get recommended plans with pagination
	if err := r.db.WithContext(ctx).Where(condition, args...).
		Order("sort_order ASC, price ASC").
		Limit(limit).Offset(offset).Find(&plans).Error; err != nil {
		logger.Error("Failed to list recommended subscription plans", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to list recommended subscription plans: %w", err)
	}

	return plans, total, nil
}

// ListPublic lists public visible active subscription plans (for non-authenticated users)
func (r *subscriptionPlanRepository) ListPublic(ctx context.Context, limit, offset int) ([]*entities.SubscriptionPlan, int64, error) {
	var plans []*entities.SubscriptionPlan
	var total int64

	condition := "status = ? AND is_visible = ?"
	args := []interface{}{entities.SubscriptionPlanStatusActive, true}

	// Count total public plans
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where(condition, args...).Count(&total).Error; err != nil {
		logger.Error("Failed to count public subscription plans", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count public subscription plans: %w", err)
	}

	// Get public plans with pagination
	if err := r.db.WithContext(ctx).Where(condition, args...).
		Order("sort_order ASC, price ASC").
		Limit(limit).Offset(offset).Find(&plans).Error; err != nil {
		logger.Error("Failed to list public subscription plans", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to list public subscription plans: %w", err)
	}

	return plans, total, nil
}

// ListPublicByCurrency lists public subscription plans by currency
func (r *subscriptionPlanRepository) ListPublicByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.SubscriptionPlan, int64, error) {
	var plans []*entities.SubscriptionPlan
	var total int64

	condition := "status = ? AND is_visible = ? AND currency = ?"
	args := []interface{}{entities.SubscriptionPlanStatusActive, true, currency}

	// Count total public plans by currency
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where(condition, args...).Count(&total).Error; err != nil {
		logger.Error("Failed to count public subscription plans by currency",
			logger.String("currency", currency),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count public subscription plans by currency: %w", err)
	}

	// Get public plans with pagination
	if err := r.db.WithContext(ctx).Where(condition, args...).
		Order("sort_order ASC, price ASC").
		Limit(limit).Offset(offset).Find(&plans).Error; err != nil {
		logger.Error("Failed to list public subscription plans by currency",
			logger.String("currency", currency),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list public subscription plans by currency: %w", err)
	}

	return plans, total, nil
}

// ListByPriceRange lists subscription plans within a price range
func (r *subscriptionPlanRepository) ListByPriceRange(ctx context.Context, minPrice, maxPrice float64, currency string, limit, offset int) ([]*entities.SubscriptionPlan, int64, error) {
	var plans []*entities.SubscriptionPlan
	var total int64

	condition := "price BETWEEN ? AND ? AND currency = ?"
	args := []interface{}{minPrice, maxPrice, currency}

	// Count total plans in price range
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where(condition, args...).Count(&total).Error; err != nil {
		logger.Error("Failed to count subscription plans by price range",
			logger.String("min_price", fmt.Sprintf("%.2f", minPrice)),
			logger.String("max_price", fmt.Sprintf("%.2f", maxPrice)),
			logger.String("currency", currency),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count subscription plans by price range: %w", err)
	}

	// Get plans with pagination
	if err := r.db.WithContext(ctx).Where(condition, args...).
		Order("price ASC, sort_order ASC").
		Limit(limit).Offset(offset).Find(&plans).Error; err != nil {
		logger.Error("Failed to list subscription plans by price range",
			logger.String("min_price", fmt.Sprintf("%.2f", minPrice)),
			logger.String("max_price", fmt.Sprintf("%.2f", maxPrice)),
			logger.String("currency", currency),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list subscription plans by price range: %w", err)
	}

	return plans, total, nil
}

// GetCheapest gets the cheapest plan for a currency
func (r *subscriptionPlanRepository) GetCheapest(ctx context.Context, currency string) (*entities.SubscriptionPlan, error) {
	var plan entities.SubscriptionPlan
	if err := r.db.WithContext(ctx).Where("currency = ? AND status = ?", currency, entities.SubscriptionPlanStatusActive).
		Order("price ASC").First(&plan).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no plans found for currency %s", currency)
		}
		logger.Error("Failed to get cheapest subscription plan",
			logger.String("currency", currency),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get cheapest subscription plan: %w", err)
	}
	return &plan, nil
}

// GetMostExpensive gets the most expensive plan for a currency
func (r *subscriptionPlanRepository) GetMostExpensive(ctx context.Context, currency string) (*entities.SubscriptionPlan, error) {
	var plan entities.SubscriptionPlan
	if err := r.db.WithContext(ctx).Where("currency = ? AND status = ?", currency, entities.SubscriptionPlanStatusActive).
		Order("price DESC").First(&plan).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no plans found for currency %s", currency)
		}
		logger.Error("Failed to get most expensive subscription plan",
			logger.String("currency", currency),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get most expensive subscription plan: %w", err)
	}
	return &plan, nil
}

// Search searches subscription plans by name, code, or description
func (r *subscriptionPlanRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.SubscriptionPlan, int64, error) {
	var plans []*entities.SubscriptionPlan
	var total int64

	// Prepare search query
	searchQuery := "%" + strings.ToLower(query) + "%"
	whereClause := "LOWER(name) LIKE ? OR LOWER(code) LIKE ? OR LOWER(description) LIKE ?"

	// Count total matching plans
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where(whereClause, searchQuery, searchQuery, searchQuery).Count(&total).Error; err != nil {
		logger.Error("Failed to count search results",
			logger.String("query", query),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Get matching plans with pagination
	if err := r.db.WithContext(ctx).Where(whereClause, searchQuery, searchQuery, searchQuery).
		Order("sort_order ASC, created_at DESC").
		Limit(limit).Offset(offset).Find(&plans).Error; err != nil {
		logger.Error("Failed to search subscription plans",
			logger.String("query", query),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to search subscription plans: %w", err)
	}

	return plans, total, nil
}

// UpdateStatus updates a subscription plan's status
func (r *subscriptionPlanRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	result := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		logger.Error("Failed to update subscription plan status",
			logger.Uint("plan_id", id),
			logger.String("status", status),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to update subscription plan status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription plan not found")
	}

	logger.Debug("Subscription plan status updated successfully",
		logger.Uint("plan_id", id),
		logger.String("new_status", status),
	)
	return nil
}

// UpdateVisibility updates a subscription plan's visibility
func (r *subscriptionPlanRepository) UpdateVisibility(ctx context.Context, id uint, isVisible bool) error {
	result := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).Where("id = ?", id).Update("is_visible", isVisible)
	if result.Error != nil {
		logger.Error("Failed to update subscription plan visibility",
			logger.Uint("plan_id", id),
			logger.String("is_visible", fmt.Sprintf("%t", isVisible)),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to update subscription plan visibility: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription plan not found")
	}

	logger.Debug("Subscription plan visibility updated successfully",
		logger.Uint("plan_id", id),
		logger.String("is_visible", fmt.Sprintf("%t", isVisible)),
	)
	return nil
}

// UpdateSortOrder updates a subscription plan's sort order
func (r *subscriptionPlanRepository) UpdateSortOrder(ctx context.Context, id uint, sortOrder int) error {
	result := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).Where("id = ?", id).Update("sort_order", sortOrder)
	if result.Error != nil {
		logger.Error("Failed to update subscription plan sort order",
			logger.Uint("plan_id", id),
			logger.Int("sort_order", sortOrder),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to update subscription plan sort order: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription plan not found")
	}

	logger.Debug("Subscription plan sort order updated successfully",
		logger.Uint("plan_id", id),
		logger.Int("sort_order", sortOrder),
	)
	return nil
}

// UpdatePopularFlag updates a subscription plan's popular flag
func (r *subscriptionPlanRepository) UpdatePopularFlag(ctx context.Context, id uint, isPopular bool) error {
	result := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).Where("id = ?", id).Update("is_popular", isPopular)
	if result.Error != nil {
		logger.Error("Failed to update subscription plan popular flag",
			logger.Uint("plan_id", id),
			logger.String("is_popular", fmt.Sprintf("%t", isPopular)),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to update subscription plan popular flag: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription plan not found")
	}

	logger.Debug("Subscription plan popular flag updated successfully",
		logger.Uint("plan_id", id),
		logger.String("is_popular", fmt.Sprintf("%t", isPopular)),
	)
	return nil
}

// UpdateRecommendedFlag updates a subscription plan's recommended flag
func (r *subscriptionPlanRepository) UpdateRecommendedFlag(ctx context.Context, id uint, isRecommended bool) error {
	result := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).Where("id = ?", id).Update("is_recommended", isRecommended)
	if result.Error != nil {
		logger.Error("Failed to update subscription plan recommended flag",
			logger.Uint("plan_id", id),
			logger.String("is_recommended", fmt.Sprintf("%t", isRecommended)),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to update subscription plan recommended flag: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription plan not found")
	}

	logger.Debug("Subscription plan recommended flag updated successfully",
		logger.Uint("plan_id", id),
		logger.String("is_recommended", fmt.Sprintf("%t", isRecommended)),
	)
	return nil
}

// BatchUpdateStatus updates status for multiple subscription plans
func (r *subscriptionPlanRepository) BatchUpdateStatus(ctx context.Context, ids []uint, status string) (int, []uint, error) {
	var updatedCount int
	var failedIDs []uint

	for _, id := range ids {
		if err := r.UpdateStatus(ctx, id, status); err != nil {
			failedIDs = append(failedIDs, id)
		} else {
			updatedCount++
		}
	}

	logger.Debug("Batch status update completed",
		logger.Int("updated_count", updatedCount),
		logger.Int("failed_count", len(failedIDs)),
		logger.String("status", status),
	)

	return updatedCount, failedIDs, nil
}

// BatchUpdateVisibility updates visibility for multiple subscription plans
func (r *subscriptionPlanRepository) BatchUpdateVisibility(ctx context.Context, ids []uint, isVisible bool) (int, []uint, error) {
	var updatedCount int
	var failedIDs []uint

	for _, id := range ids {
		if err := r.UpdateVisibility(ctx, id, isVisible); err != nil {
			failedIDs = append(failedIDs, id)
		} else {
			updatedCount++
		}
	}

	logger.Debug("Batch visibility update completed",
		logger.Int("updated_count", updatedCount),
		logger.Int("failed_count", len(failedIDs)),
		logger.String("is_visible", fmt.Sprintf("%t", isVisible)),
	)

	return updatedCount, failedIDs, nil
}

// BatchDelete performs batch soft delete on multiple subscription plans
func (r *subscriptionPlanRepository) BatchDelete(ctx context.Context, ids []uint) (int, []uint, error) {
	var deletedCount int
	var failedIDs []uint

	for _, id := range ids {
		if err := r.Delete(ctx, id); err != nil {
			failedIDs = append(failedIDs, id)
		} else {
			deletedCount++
		}
	}

	logger.Debug("Batch delete completed",
		logger.Int("deleted_count", deletedCount),
		logger.Int("failed_count", len(failedIDs)),
	)

	return deletedCount, failedIDs, nil
}

// CountTotal returns the total count of subscription plans
func (r *subscriptionPlanRepository) CountTotal(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count total subscription plans: %w", err)
	}
	return count, nil
}

// CountByStatus returns the count of subscription plans by status
func (r *subscriptionPlanRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where("status = ?", status).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count subscription plans by status %s: %w", status, err)
	}
	return count, nil
}

// CountVisible returns the count of visible subscription plans
func (r *subscriptionPlanRepository) CountVisible(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where("is_visible = ?", true).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count visible subscription plans: %w", err)
	}
	return count, nil
}

// CountByCurrency returns the count of subscription plans by currency
func (r *subscriptionPlanRepository) CountByCurrency(ctx context.Context, currency string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where("currency = ?", currency).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count subscription plans for currency %s: %w", currency, err)
	}
	return count, nil
}

// CountByBillingCycle returns the count of subscription plans by billing cycle
func (r *subscriptionPlanRepository) CountByBillingCycle(ctx context.Context, billingCycle string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where("billing_cycle = ?", billingCycle).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count subscription plans for billing cycle %s: %w", billingCycle, err)
	}
	return count, nil
}

// ExistsByCode checks if a subscription plan with the given code exists
func (r *subscriptionPlanRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).Where("code = ?", code).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check subscription plan existence by code: %w", err)
	}
	return count > 0, nil
}

// ExistsByID checks if a subscription plan with the given ID exists
func (r *subscriptionPlanRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check subscription plan existence by ID: %w", err)
	}
	return count > 0, nil
}

// GetOrderedPlans gets subscription plans ordered by a specific field
func (r *subscriptionPlanRepository) GetOrderedPlans(ctx context.Context, orderBy string, ascending bool, limit, offset int) ([]*entities.SubscriptionPlan, int64, error) {
	var plans []*entities.SubscriptionPlan
	var total int64

	direction := "DESC"
	if ascending {
		direction = "ASC"
	}

	orderClause := fmt.Sprintf("%s %s", orderBy, direction)

	// Count total plans
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count ordered subscription plans", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count ordered subscription plans: %w", err)
	}

	// Get plans with pagination and ordering
	if err := r.db.WithContext(ctx).Order(orderClause).
		Limit(limit).Offset(offset).Find(&plans).Error; err != nil {
		logger.Error("Failed to list ordered subscription plans",
			logger.String("order_by", orderBy),
			logger.String("ascending", fmt.Sprintf("%t", ascending)),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list ordered subscription plans: %w", err)
	}

	return plans, total, nil
}

// GetPlansSortedByPrice gets subscription plans sorted by price for a specific currency
func (r *subscriptionPlanRepository) GetPlansSortedByPrice(ctx context.Context, currency string, ascending bool, limit, offset int) ([]*entities.SubscriptionPlan, int64, error) {
	var plans []*entities.SubscriptionPlan
	var total int64

	direction := "DESC"
	if ascending {
		direction = "ASC"
	}

	orderClause := fmt.Sprintf("price %s", direction)

	// Count total plans for currency
	if err := r.db.WithContext(ctx).Model(&entities.SubscriptionPlan{}).
		Where("currency = ?", currency).Count(&total).Error; err != nil {
		logger.Error("Failed to count subscription plans sorted by price",
			logger.String("currency", currency),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count subscription plans sorted by price: %w", err)
	}

	// Get plans with pagination and ordering
	if err := r.db.WithContext(ctx).Where("currency = ?", currency).
		Order(orderClause).Limit(limit).Offset(offset).Find(&plans).Error; err != nil {
		logger.Error("Failed to list subscription plans sorted by price",
			logger.String("currency", currency),
			logger.String("ascending", fmt.Sprintf("%t", ascending)),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list subscription plans sorted by price: %w", err)
	}

	return plans, total, nil
}

// === SubscriptionOrderRepository Implementation ===

// BatchDelete performs batch soft delete on multiple subscription orders
func (r *subscriptionOrderRepository) BatchDelete(ctx context.Context, ids []uint) (int, []uint, error) {
	var deletedCount int
	var failedIDs []uint

	for _, id := range ids {
		result := r.db.WithContext(ctx).Delete(&entities.SubscriptionOrder{}, id)
		if result.Error != nil {
			failedIDs = append(failedIDs, id)
		} else if result.RowsAffected > 0 {
			deletedCount++
		}
	}

	logger.Debug("Batch delete completed",
		logger.Int("deleted_count", deletedCount),
		logger.Int("failed_count", len(failedIDs)),
	)

	return deletedCount, failedIDs, nil
}

// Essential SubscriptionOrderRepository methods (simplified for interface compliance)
func (r *subscriptionOrderRepository) Create(ctx context.Context, order *entities.SubscriptionOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *subscriptionOrderRepository) GetByID(ctx context.Context, id uint) (*entities.SubscriptionOrder, error) {
	var order entities.SubscriptionOrder
	err := r.db.WithContext(ctx).First(&order, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("subscription order not found")
	}
	return &order, err
}

func (r *subscriptionOrderRepository) GetByOrderNumber(ctx context.Context, orderNumber string) (*entities.SubscriptionOrder, error) {
	var order entities.SubscriptionOrder
	err := r.db.WithContext(ctx).Where("order_number = ?", orderNumber).First(&order).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("subscription order not found")
	}
	return &order, err
}

func (r *subscriptionOrderRepository) Update(ctx context.Context, order *entities.SubscriptionOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *subscriptionOrderRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entities.SubscriptionOrder{}, id)
	if result.RowsAffected == 0 {
		return fmt.Errorf("subscription order not found")
	}
	return result.Error
}

// Stub implementations for remaining SubscriptionOrderRepository methods
func (r *subscriptionOrderRepository) SoftDelete(ctx context.Context, id uint) error {
	return r.Delete(ctx, id)
}
func (r *subscriptionOrderRepository) Restore(ctx context.Context, id uint) error {
	return fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) HardDelete(ctx context.Context, id uint) error {
	return fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) List(ctx context.Context, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListDeleted(ctx context.Context, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListByUser(ctx context.Context, userID uint, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) GetUserOrderHistory(ctx context.Context, userID uint, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) GetUserActiveOrders(ctx context.Context, userID uint, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListByPlan(ctx context.Context, planID uint, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) GetPlanOrderStats(ctx context.Context, planID uint, since time.Time) (map[string]int64, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListByOrderType(ctx context.Context, orderType string, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListByPaymentStatus(ctx context.Context, paymentStatus string, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListPendingPayments(ctx context.Context, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListFailedPayments(ctx context.Context, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListByTransactionID(ctx context.Context, transactionID string) ([]*entities.SubscriptionOrder, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListByPaymentGateway(ctx context.Context, gateway string, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListByDateRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListRecentOrders(ctx context.Context, since time.Time, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListOrdersForBillingPeriod(ctx context.Context, start, end time.Time, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListByCouponCode(ctx context.Context, couponCode string, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListWithDiscounts(ctx context.Context, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListRefundedOrders(ctx context.Context, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListRefundableOrders(ctx context.Context, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListByInvoiceStatus(ctx context.Context, invoiceStatus string, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListUninvoicedOrders(ctx context.Context, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) UpdatePaymentStatus(ctx context.Context, id uint, paymentStatus string, transactionID string) error {
	return fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) UpdateInvoiceStatus(ctx context.Context, id uint, invoiceStatus string, invoiceNumber string) error {
	return fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) MarkAsPaid(ctx context.Context, id uint, transactionID string, paidAt time.Time) error {
	return fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) MarkAsRefunded(ctx context.Context, id uint, refundAmount float64, refundReason string, refundedAt time.Time) error {
	return fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) SearchByUserEmail(ctx context.Context, email string, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) CountTotal(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) CountByUser(ctx context.Context, userID uint) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) CountByPlan(ctx context.Context, planID uint) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) CountPaidOrders(ctx context.Context, since time.Time) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) CountFailedOrders(ctx context.Context, since time.Time) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) GetTotalRevenue(ctx context.Context, currency string, since time.Time) (float64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) GetRevenueByPlan(ctx context.Context, planID uint, currency string, since time.Time) (float64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) GetRevenueByPeriod(ctx context.Context, currency string, start, end time.Time) (float64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) GetDailyRevenue(ctx context.Context, currency string, days int) (map[string]float64, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) GetMonthlyRevenue(ctx context.Context, currency string, months int) (map[string]float64, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) GetSupportedCurrencies(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) BatchUpdateStatus(ctx context.Context, ids []uint, status string) (int, []uint, error) {
	return 0, nil, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) GetLastOrderNumber(ctx context.Context, prefix string) (string, error) {
	return "", fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ExistsByOrderNumber(ctx context.Context, orderNumber string) (bool, error) {
	return false, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListBySubscription(ctx context.Context, subscriptionID uint, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) GetSubscriptionOrders(ctx context.Context, subscriptionID uint) ([]*entities.SubscriptionOrder, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) ListWithFilters(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *subscriptionOrderRepository) GetOrdersForRenewal(ctx context.Context, beforeDate time.Time, limit int) ([]*entities.SubscriptionOrder, error) {
	return nil, fmt.Errorf("not implemented")
}

// === UserSubscriptionRepository Implementation ===

// AddTrafficUsage adds traffic usage to a user subscription
func (r *userSubscriptionRepository) AddTrafficUsage(ctx context.Context, id uint, additionalBytes int64) error {
	result := r.db.WithContext(ctx).Model(&entities.UserSubscription{}).
		Where("id = ?", id).
		Update("traffic_used", gorm.Expr("traffic_used + ?", additionalBytes))

	if result.Error != nil {
		logger.Error("Failed to add traffic usage",
			logger.Uint("subscription_id", id),
			logger.Int64("additional_bytes", additionalBytes),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to add traffic usage: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user subscription not found")
	}

	logger.Debug("Traffic usage added successfully",
		logger.Uint("subscription_id", id),
		logger.Int64("additional_bytes", additionalBytes),
	)
	return nil
}

// Essential UserSubscriptionRepository methods (simplified for interface compliance)
func (r *userSubscriptionRepository) Create(ctx context.Context, subscription *entities.UserSubscription) error {
	return r.db.WithContext(ctx).Create(subscription).Error
}

func (r *userSubscriptionRepository) GetByID(ctx context.Context, id uint) (*entities.UserSubscription, error) {
	var subscription entities.UserSubscription
	err := r.db.WithContext(ctx).First(&subscription, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("user subscription not found")
	}
	return &subscription, err
}

func (r *userSubscriptionRepository) GetByUUID(ctx context.Context, uuid string) (*entities.UserSubscription, error) {
	var subscription entities.UserSubscription
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&subscription).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("user subscription not found")
	}
	return &subscription, err
}

func (r *userSubscriptionRepository) Update(ctx context.Context, subscription *entities.UserSubscription) error {
	return r.db.WithContext(ctx).Save(subscription).Error
}

func (r *userSubscriptionRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entities.UserSubscription{}, id)
	if result.RowsAffected == 0 {
		return fmt.Errorf("user subscription not found")
	}
	return result.Error
}

// Stub implementations for remaining UserSubscriptionRepository methods
func (r *userSubscriptionRepository) SoftDelete(ctx context.Context, id uint) error {
	return r.Delete(ctx, id)
}
func (r *userSubscriptionRepository) Restore(ctx context.Context, id uint) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) HardDelete(ctx context.Context, id uint) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListByUser(ctx context.Context, userID uint, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) GetActiveByUser(ctx context.Context, userID uint) ([]*entities.UserSubscription, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) GetUserCurrentSubscription(ctx context.Context, userID uint) (*entities.UserSubscription, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) GetUserActiveSubscriptions(ctx context.Context, userID uint) ([]*entities.UserSubscription, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) GetUserExpiredSubscriptions(ctx context.Context, userID uint, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListByPlan(ctx context.Context, planID uint, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) CountByPlan(ctx context.Context, planID uint) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) GetActivePlanSubscriptions(ctx context.Context, planID uint, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListActive(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListExpired(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListCancelled(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListInTrial(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListExpiringBefore(ctx context.Context, beforeDate time.Time, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListForRenewal(ctx context.Context, beforeDate time.Time, limit int) ([]*entities.UserSubscription, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListOverdueRenewals(ctx context.Context, limit int) ([]*entities.UserSubscription, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListTrialsExpiring(ctx context.Context, beforeDate time.Time, limit int) ([]*entities.UserSubscription, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListByDateRange(ctx context.Context, field string, start, end time.Time, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListCreatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListLastUsedBefore(ctx context.Context, before time.Time, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListByTrafficUsage(ctx context.Context, minUsage, maxUsage int64, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListTrafficSuspended(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListNearTrafficLimit(ctx context.Context, thresholdPercent float64, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListForTrafficReset(ctx context.Context, beforeDate time.Time, limit int) ([]*entities.UserSubscription, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListAutoRenewEnabled(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListFailedRenewals(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListPendingCancellations(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListByServerGroupAccess(ctx context.Context, serverGroupID uint, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) GetSubscriptionsWithServerAccess(ctx context.Context, serverGroupID uint) ([]*entities.UserSubscription, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) UpdateLastUsed(ctx context.Context, id uint, lastUsedAt time.Time) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) UpdateTrafficUsage(ctx context.Context, id uint, trafficUsed int64) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ResetTrafficUsage(ctx context.Context, id uint, resetDate time.Time) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) SuspendForTrafficLimit(ctx context.Context, id uint) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) UnsuspendTraffic(ctx context.Context, id uint) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) UpdateNextBillingDate(ctx context.Context, id uint, nextBillingDate time.Time) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) UpdateRenewalAttempts(ctx context.Context, id uint, attempts int) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) MarkRenewalFailed(ctx context.Context, id uint, failedAt time.Time, reason string) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ResetRenewalAttempts(ctx context.Context, id uint) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) CancelSubscription(ctx context.Context, id uint, reason string, cancelAtPeriodEnd bool) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) CancelAtPeriodEnd(ctx context.Context, id uint, reason string) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) UncancelSubscription(ctx context.Context, id uint) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) UpdateServerGroupAccess(ctx context.Context, id uint, serverGroupIDs []uint) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) GrantServerGroupAccess(ctx context.Context, id uint, serverGroupID uint) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) RevokeServerGroupAccess(ctx context.Context, id uint, serverGroupID uint) error {
	return fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) BatchUpdateStatus(ctx context.Context, ids []uint, status string) (int, []uint, error) {
	return 0, nil, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) BatchCancel(ctx context.Context, ids []uint, reason string) (int, []uint, error) {
	return 0, nil, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) BatchResetTraffic(ctx context.Context, ids []uint, resetDate time.Time) (int, []uint, error) {
	return 0, nil, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) BatchDelete(ctx context.Context, ids []uint) (int, []uint, error) {
	return 0, nil, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) List(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListDeleted(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) SearchByUserEmail(ctx context.Context, email string, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) CountTotal(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) CountActiveSubscriptions(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) CountExpiredSubscriptions(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) CountTrialSubscriptions(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) CountByUser(ctx context.Context, userID uint) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) GetSubscriptionStats(ctx context.Context, since time.Time) (map[string]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) GetChurnRate(ctx context.Context, period time.Duration) (float64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) GetRetentionRate(ctx context.Context, period time.Duration) (float64, error) {
	return 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ExistsByUUID(ctx context.Context, uuid string) (bool, error) {
	return false, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	return false, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) UserHasActiveSubscription(ctx context.Context, userID uint) (bool, error) {
	return false, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) UserHasSubscriptionToPlan(ctx context.Context, userID, planID uint) (bool, error) {
	return false, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) ListWithFilters(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.UserSubscription, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) GetSubscriptionsNeedingAttention(ctx context.Context, limit int) ([]*entities.UserSubscription, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *userSubscriptionRepository) GetSubscriptionsForMaintenance(ctx context.Context, maintenanceType string, limit int) ([]*entities.UserSubscription, error) {
	return nil, fmt.Errorf("not implemented")
}
