package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"linke/internal/logger"
	"linke/internal/model"

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

// CreateCouponRequest represents the request to create a coupon
type CreateCouponRequest struct {
	Code            string     `json:"code" binding:"required,min=3,max=50" example:"SAVE20"`
	Name            string     `json:"name" binding:"required,min=1,max=100" example:"20% Off All Plans"`
	Description     string     `json:"description,omitempty" binding:"max=1000" example:"Save 20% on any subscription plan"`
	Type            string     `json:"type" binding:"required,oneof=percentage fixed_amount" example:"percentage"`
	Value           float64    `json:"value" binding:"required,min=0" example:"20"`
	MaxUses         int        `json:"max_uses,omitempty" binding:"min=0" example:"100"`
	MaxUsesPerUser  int        `json:"max_uses_per_user,omitempty" binding:"min=1" example:"1"`
	MinOrderAmount  float64    `json:"min_order_amount,omitempty" binding:"min=0" example:"10"`
	Currency        string     `json:"currency,omitempty" binding:"omitempty,len=3" example:"USD"`
	ValidFrom       *time.Time `json:"valid_from,omitempty" example:"2024-01-01T00:00:00Z"`
	ValidUntil      *time.Time `json:"valid_until,omitempty" example:"2024-12-31T23:59:59Z"`
	ApplicablePlans string     `json:"applicable_plans,omitempty" example:"[1,2,3]"`
	IsPublic        *bool      `json:"is_public,omitempty" example:"true"`
}

// UpdateCouponRequest represents the request to update a coupon
type UpdateCouponRequest struct {
	Name            *string    `json:"name,omitempty" binding:"omitempty,min=1,max=100" example:"Updated Coupon Name"`
	Description     *string    `json:"description,omitempty" binding:"omitempty,max=1000" example:"Updated description"`
	Type            *string    `json:"type,omitempty" binding:"omitempty,oneof=percentage fixed_amount" example:"percentage"`
	Value           *float64   `json:"value,omitempty" binding:"omitempty,min=0" example:"25"`
	MaxUses         *int       `json:"max_uses,omitempty" binding:"omitempty,min=0" example:"200"`
	MaxUsesPerUser  *int       `json:"max_uses_per_user,omitempty" binding:"omitempty,min=1" example:"2"`
	MinOrderAmount  *float64   `json:"min_order_amount,omitempty" binding:"omitempty,min=0" example:"15"`
	ValidFrom       *time.Time `json:"valid_from,omitempty" example:"2024-02-01T00:00:00Z"`
	ValidUntil      *time.Time `json:"valid_until,omitempty" example:"2024-11-30T23:59:59Z"`
	ApplicablePlans *string    `json:"applicable_plans,omitempty" example:"[1,2,4]"`
	Status          *string    `json:"status,omitempty" binding:"omitempty,oneof=active inactive expired" example:"active"`
	IsPublic        *bool      `json:"is_public,omitempty" example:"false"`
}

// GetCouponsRequest represents the request to get coupons
type GetCouponsRequest struct {
	Status   string `form:"status,omitempty" binding:"omitempty,oneof=active inactive expired" example:"active"`
	Type     string `form:"type,omitempty" binding:"omitempty,oneof=percentage fixed_amount" example:"percentage"`
	IsPublic *bool  `form:"is_public,omitempty" example:"true"`
	Limit    int    `form:"limit,omitempty" binding:"omitempty,min=1,max=100" example:"10"`
	Offset   int    `form:"offset,omitempty" binding:"omitempty,min=0" example:"0"`
}

// ValidateCouponRequest represents the request to validate a coupon
type ValidateCouponRequest struct {
	Code        string  `json:"code" binding:"required" example:"SAVE20"`
	UserID      uint64  `json:"user_id" binding:"required" example:"1"`
	OrderAmount float64 `json:"order_amount" binding:"required,min=0" example:"29.99"`
	PlanID      uint64  `json:"plan_id" binding:"required" example:"1"`
	Currency    string  `json:"currency" binding:"required,len=3" example:"USD"`
}

// ValidateCouponResponse represents the response of coupon validation
type ValidateCouponResponse struct {
	Valid          bool    `json:"valid" example:"true"`
	Message        string  `json:"message" example:"Coupon is valid"`
	DiscountAmount float64 `json:"discount_amount" example:"5.99"`
	FinalAmount    float64 `json:"final_amount" example:"24.00"`
	Coupon         *model.CouponResponse `json:"coupon,omitempty"`
}

// CreateCoupon creates a new coupon
func (s *CouponService) CreateCoupon(ctx context.Context, creatorID uint64, req *CreateCouponRequest) (*model.Coupon, error) {
	// Validate and normalize code
	code := strings.TrimSpace(strings.ToUpper(req.Code))
	if code == "" {
		return nil, fmt.Errorf("coupon code cannot be empty")
	}

	// Check if code already exists
	var existingCoupon model.Coupon
	if err := s.db.Where("code = ?", code).First(&existingCoupon).Error; err == nil {
		return nil, fmt.Errorf("coupon with code '%s' already exists", code)
	} else if err != gorm.ErrRecordNotFound {
		logger.Error("Failed to check existing coupon", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to check existing coupon: %w", err)
	}

	// Validate percentage values
	if req.Type == model.CouponTypePercentage && req.Value > 100 {
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
	coupon := &model.Coupon{
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
		Status:          model.CouponStatusActive,
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
func (s *CouponService) GetCoupon(ctx context.Context, couponID uint64) (*model.Coupon, error) {
	var coupon model.Coupon
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
func (s *CouponService) GetCouponByCode(ctx context.Context, code string) (*model.Coupon, error) {
	var coupon model.Coupon
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
func (s *CouponService) GetCoupons(ctx context.Context, req *GetCouponsRequest) ([]*model.Coupon, int64, error) {
	query := s.db.WithContext(ctx).Model(&model.Coupon{})

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

	var coupons []*model.Coupon
	if err := query.Find(&coupons).Error; err != nil {
		logger.Error("Failed to get coupons", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to get coupons: %w", err)
	}

	return coupons, totalCount, nil
}

// GetPublicCoupons gets active and public coupons
// NOTE: This method should only be used internally by the system,
// not exposed to public APIs for security reasons. 
// The 'is_public' flag now indicates whether a coupon can be displayed
// in user interfaces after proper authentication and authorization.
func (s *CouponService) GetPublicCoupons(ctx context.Context) ([]*model.Coupon, error) {
	var coupons []*model.Coupon
	if err := s.db.WithContext(ctx).
		Where("status = ? AND is_public = ?", model.CouponStatusActive, true).
		Order("created_at DESC").
		Find(&coupons).Error; err != nil {
		logger.Error("Failed to get public coupons", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get public coupons: %w", err)
	}

	return coupons, nil
}

// UpdateCoupon updates a coupon
func (s *CouponService) UpdateCoupon(ctx context.Context, couponID uint64, req *UpdateCouponRequest) (*model.Coupon, error) {
	// Get existing coupon
	coupon, err := s.GetCoupon(ctx, couponID)
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

	if req.Type != nil {
		updates["type"] = *req.Type
	}

	if req.Value != nil {
		// Validate percentage values
		if (req.Type != nil && *req.Type == model.CouponTypePercentage) || 
		   (req.Type == nil && coupon.Type == model.CouponTypePercentage) {
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
func (s *CouponService) ValidateCoupon(ctx context.Context, req *ValidateCouponRequest) (*ValidateCouponResponse, error) {
	// Get coupon by code
	coupon, err := s.GetCouponByCode(ctx, req.Code)
	if err != nil {
		return &ValidateCouponResponse{
			Valid:   false,
			Message: "Coupon not found",
		}, nil
	}

	// Check if coupon can be used by this user
	canUse, message := coupon.CanBeUsedBy(req.UserID, req.OrderAmount, req.PlanID, s.db)
	if !canUse {
		return &ValidateCouponResponse{
			Valid:   false,
			Message: message,
			Coupon:  coupon.ToPublicResponse(),
		}, nil
	}

	// Check currency match
	if coupon.Currency != req.Currency {
		return &ValidateCouponResponse{
			Valid:   false,
			Message: fmt.Sprintf("Coupon is only valid for %s currency", coupon.Currency),
			Coupon:  coupon.ToPublicResponse(),
		}, nil
	}

	// Calculate discount
	discountAmount := coupon.CalculateDiscount(req.OrderAmount)
	finalAmount := req.OrderAmount - discountAmount

	return &ValidateCouponResponse{
		Valid:          true,
		Message:        "Coupon is valid",
		DiscountAmount: discountAmount,
		FinalAmount:    finalAmount,
		Coupon:         coupon.ToPublicResponse(),
	}, nil
}

// UseCoupon records the usage of a coupon
func (s *CouponService) UseCoupon(ctx context.Context, couponID, userID, orderID uint64, discountAmount, orderAmount float64, currency string) error {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get coupon with row lock
	var coupon model.Coupon
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
	usage := &model.CouponUsage{
		CouponID:            couponID,
		UserID:              userID,
		OrderID:         orderID,
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
func (s *CouponService) GetCouponUsages(ctx context.Context, couponID *uint64, userID *uint64, limit, offset int) ([]*model.CouponUsage, int64, error) {
	query := s.db.WithContext(ctx).Model(&model.CouponUsage{})

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

	var usages []*model.CouponUsage
	if err := query.Find(&usages).Error; err != nil {
		logger.Error("Failed to get coupon usages", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to get coupon usages: %w", err)
	}

	return usages, totalCount, nil
}