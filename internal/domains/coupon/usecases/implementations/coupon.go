package implementations

import (
	"context"
	"fmt"
	"strings"

	"linke/internal/domains/coupon/entities"
	"linke/internal/domains/coupon/usecases/interfaces"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

type CouponService struct {
	db *gorm.DB
}

func NewCouponService(db *gorm.DB) *CouponService {
	return &CouponService{
		db: db,
	}
}

// CreateCoupon creates a new coupon
func (s *CouponService) CreateCoupon(ctx context.Context, creatorID uint64, req *interfaces.CreateCouponRequest) (*entities.Coupon, error) {
	// Validate and normalize code
	code := strings.TrimSpace(strings.ToUpper(req.Code))
	if code == "" {
		return nil, fmt.Errorf("coupon code cannot be empty")
	}

	// Check if code already exists
	var existingCoupon entities.Coupon
	if err := s.db.Where("code = ?", code).First(&existingCoupon).Error; err == nil {
		return nil, fmt.Errorf("coupon with code '%s' already exists", code)
	} else if err != gorm.ErrRecordNotFound {
		logger.Error("Failed to check existing coupon", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to check existing coupon: %w", err)
	}

	// Validate percentage values
	if req.Type == entities.CouponTypePercentage && req.Value > 100 {
		return nil, fmt.Errorf("percentage discount cannot exceed 100%%")
	}

	// Set defaults
	currency := "USD"
	if req.Currency != "" {
		currency = strings.ToUpper(req.Currency)
	}

	maxUsesPerUser := 1
	if req.MaxUsesPerUser > 0 {
		maxUsesPerUser = req.MaxUsesPerUser
	}

	isPublic := false
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	// Create the coupon
	coupon := &entities.Coupon{
		Code:            code,
		Name:            req.Name,
		Description:     req.Description,
		Type:            req.Type,
		Value:           req.Value,
		MaxUses:         req.MaxUses,
		MaxUsesPerUser:  maxUsesPerUser,
		MinOrderAmount:  req.MinOrderAmount,
		Currency:        currency,
		ValidFrom:       req.ValidFrom,
		ValidUntil:      req.ValidUntil,
		ApplicablePlans: req.ApplicablePlans,
		Status:          entities.CouponStatusActive,
		IsPublic:        isPublic,
		CreatedBy:       creatorID,
	}

	if err := s.db.WithContext(ctx).Create(coupon).Error; err != nil {
		logger.Error("Failed to create coupon", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to create coupon: %w", err)
	}

	logger.Info("Coupon created successfully",
		logger.Any("coupon_id", coupon.ID),
		logger.String("coupon_code", coupon.Code),
		logger.Any("creator_id", creatorID))

	return coupon, nil
}

// GetCoupon gets a coupon by ID
func (s *CouponService) GetCoupon(ctx context.Context, couponID uint64) (*entities.Coupon, error) {
	var coupon entities.Coupon
	if err := s.db.WithContext(ctx).First(&coupon, couponID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("coupon not found")
		}
		logger.Error("Failed to get coupon", logger.Error2("error", err), logger.Any("coupon_id", couponID))
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}

	return &coupon, nil
}

// GetCouponByCode gets a coupon by code
func (s *CouponService) GetCouponByCode(ctx context.Context, code string) (*entities.Coupon, error) {
	var coupon entities.Coupon
	if err := s.db.WithContext(ctx).Where("code = ?", strings.ToUpper(strings.TrimSpace(code))).First(&coupon).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("coupon not found")
		}
		logger.Error("Failed to get coupon by code", logger.Error2("error", err), logger.String("code", code))
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}

	return &coupon, nil
}

// GetCoupons gets coupons with filtering and pagination
func (s *CouponService) GetCoupons(ctx context.Context, req *interfaces.GetCouponsRequest) ([]*entities.Coupon, int64, error) {
	query := s.db.WithContext(ctx).Model(&entities.Coupon{})

	// Apply filters
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}

	if req.IsPublic != nil {
		query = query.Where("is_public = ?", *req.IsPublic)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		logger.Error("Failed to count coupons", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count coupons: %w", err)
	}

	// Apply pagination and ordering
	query = query.Order("created_at DESC")

	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	var coupons []*entities.Coupon
	if err := query.Find(&coupons).Error; err != nil {
		logger.Error("Failed to get coupons", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to get coupons: %w", err)
	}

	return coupons, totalCount, nil
}

// GetPublicCoupons gets active and public coupons with limit
// NOTE: This method should only be used internally by the system,
// not exposed to public APIs for security reasons.
// The 'is_public' flag now indicates whether a coupon can be displayed
// in user interfaces after proper authentication and authorization.
func (s *CouponService) GetPublicCoupons(ctx context.Context, limit int) ([]*entities.Coupon, error) {
	query := s.db.WithContext(ctx).
		Where("status = ? AND is_public = ?", entities.CouponStatusActive, true).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	var coupons []*entities.Coupon
	if err := query.Find(&coupons).Error; err != nil {
		logger.Error("Failed to get public coupons", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get public coupons: %w", err)
	}

	return coupons, nil
}

// UpdateCoupon updates a coupon
func (s *CouponService) UpdateCoupon(ctx context.Context, couponID uint64, req *interfaces.UpdateCouponRequest) (*entities.Coupon, error) {
	// Get existing coupon
	coupon, err := s.GetCoupon(ctx, couponID)
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

	if req.Type != nil {
		updates["type"] = *req.Type
	}

	if req.Value != nil {
		// Validate percentage values
		if (req.Type != nil && *req.Type == entities.CouponTypePercentage) ||
			(req.Type == nil && coupon.Type == entities.CouponTypePercentage) {
			if *req.Value > 100 {
				return nil, fmt.Errorf("percentage discount cannot exceed 100%%")
			}
		}
		updates["value"] = *req.Value
	}

	if req.MaxUses != nil {
		updates["max_uses"] = *req.MaxUses
	}

	if req.MaxUsesPerUser != nil {
		updates["max_uses_per_user"] = *req.MaxUsesPerUser
	}

	if req.MinOrderAmount != nil {
		updates["min_order_amount"] = *req.MinOrderAmount
	}

	if req.ValidFrom != nil {
		updates["valid_from"] = req.ValidFrom
	}

	if req.ValidUntil != nil {
		updates["valid_until"] = req.ValidUntil
	}

	if req.ApplicablePlans != nil {
		updates["applicable_plans"] = *req.ApplicablePlans
	}

	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}

	// Update the coupon
	if err := s.db.WithContext(ctx).Model(coupon).Updates(updates).Error; err != nil {
		logger.Error("Failed to update coupon", logger.Error2("error", err), logger.Any("coupon_id", couponID))
		return nil, fmt.Errorf("failed to update coupon: %w", err)
	}

	// Reload the coupon
	if err := s.db.WithContext(ctx).First(coupon, couponID).Error; err != nil {
		logger.Error("Failed to reload updated coupon", logger.Error2("error", err), logger.Any("coupon_id", couponID))
		return nil, fmt.Errorf("failed to reload updated coupon: %w", err)
	}

	logger.Info("Coupon updated successfully", logger.Any("coupon_id", coupon.ID))

	return coupon, nil
}

// DeleteCoupon soft deletes a coupon
func (s *CouponService) DeleteCoupon(ctx context.Context, couponID uint64) error {
	// Check if coupon exists
	coupon, err := s.GetCoupon(ctx, couponID)
	if err != nil {
		return err
	}

	// Soft delete the coupon
	if err := s.db.WithContext(ctx).Delete(coupon).Error; err != nil {
		logger.Error("Failed to delete coupon", logger.Error2("error", err), logger.Any("coupon_id", couponID))
		return fmt.Errorf("failed to delete coupon: %w", err)
	}

	logger.Info("Coupon deleted successfully", logger.Any("coupon_id", couponID))

	return nil
}

// ValidateCoupon validates a coupon for a specific user and order
func (s *CouponService) ValidateCoupon(ctx context.Context, req *interfaces.ValidateCouponRequest) (*interfaces.ValidateCouponResponse, error) {
	// Get coupon by code
	coupon, err := s.GetCouponByCode(ctx, req.Code)
	if err != nil {
		return &interfaces.ValidateCouponResponse{
			Valid:   false,
			Message: "Coupon not found",
		}, nil
	}

	// Check if coupon can be used by this user
	canUse, message := coupon.CanBeUsedBy(req.UserID, req.OrderAmount, req.PlanID, s.db)
	if !canUse {
		return &interfaces.ValidateCouponResponse{
			Valid:   false,
			Message: message,
			Coupon:  coupon.ToPublicResponse(),
		}, nil
	}

	// Check currency match
	if coupon.Currency != req.Currency {
		return &interfaces.ValidateCouponResponse{
			Valid:   false,
			Message: fmt.Sprintf("Coupon is only valid for %s currency", coupon.Currency),
			Coupon:  coupon.ToPublicResponse(),
		}, nil
	}

	// Calculate discount
	discountAmount := coupon.CalculateDiscount(req.OrderAmount)
	finalAmount := req.OrderAmount - discountAmount

	return &interfaces.ValidateCouponResponse{
		Valid:          true,
		Message:        "Coupon is valid",
		DiscountAmount: discountAmount,
		FinalAmount:    finalAmount,
		Coupon:         coupon.ToPublicResponse(),
	}, nil
}

// UseCoupon records the usage of a coupon (interface compatible version)
func (s *CouponService) UseCoupon(ctx context.Context, couponID, userID uint64, orderAmount float64, orderID *uint64) (*entities.CouponUsage, error) {
	// Get coupon to calculate discount
	coupon, err := s.GetCoupon(ctx, couponID)
	if err != nil {
		return nil, err
	}

	discountAmount := coupon.CalculateDiscount(orderAmount)

	var actualOrderID uint64
	if orderID != nil {
		actualOrderID = *orderID
	}

	if err := s.useCouponInternal(ctx, couponID, userID, actualOrderID, discountAmount, orderAmount, coupon.Currency); err != nil {
		return nil, err
	}

	// Return the created usage record
	var usage entities.CouponUsage
	if err := s.db.WithContext(ctx).
		Where("coupon_id = ? AND user_id = ? AND subscription_order_id = ?", couponID, userID, actualOrderID).
		Order("created_at DESC").
		First(&usage).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve coupon usage record: %w", err)
	}

	return &usage, nil
}

// useCouponInternal is the internal implementation that was originally named UseCoupon
func (s *CouponService) useCouponInternal(ctx context.Context, couponID, userID, orderID uint64, discountAmount, orderAmount float64, currency string) error {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get coupon with row lock
	var coupon entities.Coupon
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&coupon, couponID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get coupon for usage: %w", err)
	}

	// Validate coupon can still be used
	if !coupon.IsValid() {
		tx.Rollback()
		return fmt.Errorf("coupon is no longer valid")
	}

	// Check usage limits
	if coupon.MaxUses > 0 && coupon.UsedCount >= coupon.MaxUses {
		tx.Rollback()
		return fmt.Errorf("coupon usage limit exceeded")
	}

	// Create usage record
	usage := &entities.CouponUsage{
		CouponID:            couponID,
		UserID:              userID,
		SubscriptionOrderID: orderID,
		DiscountAmount:      discountAmount,
		OrderAmount:         orderAmount,
		Currency:            currency,
	}

	if err := tx.Create(usage).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to create coupon usage record", logger.Error2("error", err))
		return fmt.Errorf("failed to create coupon usage record: %w", err)
	}

	// Increment used count
	if err := tx.Model(&coupon).Update("used_count", coupon.UsedCount+1).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to update coupon used count", logger.Error2("error", err))
		return fmt.Errorf("failed to update coupon used count: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit coupon usage: %w", err)
	}

	logger.Info("Coupon used successfully",
		logger.Any("coupon_id", couponID),
		logger.Any("user_id", userID),
		logger.Any("order_id", orderID),
		logger.Any("discount_amount", discountAmount))

	return nil
}

// GetCouponUsages gets coupon usage records with filtering and pagination
func (s *CouponService) GetCouponUsages(ctx context.Context, couponID *uint64, userID *uint64, limit, offset int) ([]*entities.CouponUsage, int64, error) {
	query := s.db.WithContext(ctx).Model(&entities.CouponUsage{})

	// Apply filters
	if couponID != nil {
		query = query.Where("coupon_id = ?", *couponID)
	}

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		logger.Error("Failed to count coupon usages", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count coupon usages: %w", err)
	}

	// Apply pagination and ordering
	query = query.Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	var usages []*entities.CouponUsage
	if err := query.Find(&usages).Error; err != nil {
		logger.Error("Failed to get coupon usages", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to get coupon usages: %w", err)
	}

	return usages, totalCount, nil
}

// GetActiveCoupons gets all active coupons
func (s *CouponService) GetActiveCoupons(ctx context.Context) ([]*entities.Coupon, error) {
	var coupons []*entities.Coupon
	if err := s.db.WithContext(ctx).
		Where("status = ?", entities.CouponStatusActive).
		Order("created_at DESC").
		Find(&coupons).Error; err != nil {
		logger.Error("Failed to get active coupons", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get active coupons: %w", err)
	}

	return coupons, nil
}

// ActivateCoupon activates a coupon
func (s *CouponService) ActivateCoupon(ctx context.Context, couponID uint64) error {
	if err := s.db.WithContext(ctx).Model(&entities.Coupon{}).
		Where("id = ?", couponID).
		Update("status", entities.CouponStatusActive).Error; err != nil {
		logger.Error("Failed to activate coupon", logger.Error2("error", err), logger.Any("coupon_id", couponID))
		return fmt.Errorf("failed to activate coupon: %w", err)
	}

	logger.Info("Coupon activated successfully", logger.Any("coupon_id", couponID))
	return nil
}

// DeactivateCoupon deactivates a coupon
func (s *CouponService) DeactivateCoupon(ctx context.Context, couponID uint64) error {
	if err := s.db.WithContext(ctx).Model(&entities.Coupon{}).
		Where("id = ?", couponID).
		Update("status", entities.CouponStatusInactive).Error; err != nil {
		logger.Error("Failed to deactivate coupon", logger.Error2("error", err), logger.Any("coupon_id", couponID))
		return fmt.Errorf("failed to deactivate coupon: %w", err)
	}

	logger.Info("Coupon deactivated successfully", logger.Any("coupon_id", couponID))
	return nil
}

// ExpireCoupon expires a coupon
func (s *CouponService) ExpireCoupon(ctx context.Context, couponID uint64) error {
	if err := s.db.WithContext(ctx).Model(&entities.Coupon{}).
		Where("id = ?", couponID).
		Update("status", entities.CouponStatusExpired).Error; err != nil {
		logger.Error("Failed to expire coupon", logger.Error2("error", err), logger.Any("coupon_id", couponID))
		return fmt.Errorf("failed to expire coupon: %w", err)
	}

	logger.Info("Coupon expired successfully", logger.Any("coupon_id", couponID))
	return nil
}

// GetCouponUsage gets coupon usage records for a specific coupon
func (s *CouponService) GetCouponUsage(ctx context.Context, couponID uint64, limit, offset int) ([]*entities.CouponUsage, int64, error) {
	return s.GetCouponUsages(ctx, &couponID, nil, limit, offset)
}

// GetUserCouponUsage gets coupon usage records for a specific user
func (s *CouponService) GetUserCouponUsage(ctx context.Context, userID uint64, limit, offset int) ([]*entities.CouponUsage, int64, error) {
	return s.GetCouponUsages(ctx, nil, &userID, limit, offset)
}

// GetCouponStatistics gets statistics for a specific coupon
func (s *CouponService) GetCouponStatistics(ctx context.Context, couponID uint64) (map[string]any, error) {
	coupon, err := s.GetCoupon(ctx, couponID)
	if err != nil {
		return nil, err
	}

	// Count total usages
	var totalUsages int64
	if err := s.db.WithContext(ctx).Model(&entities.CouponUsage{}).
		Where("coupon_id = ?", couponID).
		Count(&totalUsages).Error; err != nil {
		return nil, fmt.Errorf("failed to count coupon usages: %w", err)
	}

	// Calculate total discount amount
	var totalDiscountAmount float64
	if err := s.db.WithContext(ctx).Model(&entities.CouponUsage{}).
		Where("coupon_id = ?", couponID).
		Select("COALESCE(SUM(discount_amount), 0)").
		Scan(&totalDiscountAmount).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate total discount amount: %w", err)
	}

	stats := map[string]any{
		"coupon_id":             coupon.ID,
		"coupon_code":           coupon.Code,
		"status":                coupon.Status,
		"max_uses":              coupon.MaxUses,
		"used_count":            coupon.UsedCount,
		"total_usages":          totalUsages,
		"total_discount_amount": totalDiscountAmount,
		"remaining_uses":        coupon.MaxUses - coupon.UsedCount,
	}

	if coupon.MaxUses > 0 {
		stats["usage_percentage"] = float64(coupon.UsedCount) / float64(coupon.MaxUses) * 100
	}

	return stats, nil
}

// GetCouponSystemStatistics gets overall system statistics for coupons
func (s *CouponService) GetCouponSystemStatistics(ctx context.Context) (map[string]any, error) {
	stats := make(map[string]any)

	// Count coupons by status
	statuses := []string{entities.CouponStatusActive, entities.CouponStatusInactive, entities.CouponStatusExpired}
	for _, status := range statuses {
		var count int64
		if err := s.db.WithContext(ctx).Model(&entities.Coupon{}).
			Where("status = ?", status).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("failed to count coupons with status %s: %w", status, err)
		}
		stats[status+"_coupons"] = count
	}

	// Total coupons
	var totalCoupons int64
	if err := s.db.WithContext(ctx).Model(&entities.Coupon{}).Count(&totalCoupons).Error; err != nil {
		return nil, fmt.Errorf("failed to count total coupons: %w", err)
	}
	stats["total_coupons"] = totalCoupons

	// Total usages
	var totalUsages int64
	if err := s.db.WithContext(ctx).Model(&entities.CouponUsage{}).Count(&totalUsages).Error; err != nil {
		return nil, fmt.Errorf("failed to count total coupon usages: %w", err)
	}
	stats["total_usages"] = totalUsages

	// Total discount amount
	var totalDiscountAmount float64
	if err := s.db.WithContext(ctx).Model(&entities.CouponUsage{}).
		Select("COALESCE(SUM(discount_amount), 0)").
		Scan(&totalDiscountAmount).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate total discount amount: %w", err)
	}
	stats["total_discount_amount"] = totalDiscountAmount

	return stats, nil
}
