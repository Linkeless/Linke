package interfaces

import (
	"context"
	"time"
	"linke/internal/domains/coupon/entities"
)

// CouponService defines the interface for coupon operations
type CouponService interface {
	// Coupon CRUD operations
	CreateCoupon(ctx context.Context, creatorID uint64, req *CreateCouponRequest) (*entities.Coupon, error)
	GetCoupon(ctx context.Context, couponID uint64) (*entities.Coupon, error)
	GetCouponByCode(ctx context.Context, code string) (*entities.Coupon, error)
	UpdateCoupon(ctx context.Context, couponID uint64, req *UpdateCouponRequest) (*entities.Coupon, error)
	DeleteCoupon(ctx context.Context, couponID uint64) error

	// Coupon listing and filtering
	GetCoupons(ctx context.Context, req *GetCouponsRequest) ([]*entities.Coupon, int64, error)
	GetPublicCoupons(ctx context.Context, limit int) ([]*entities.Coupon, error)
	GetActiveCoupons(ctx context.Context) ([]*entities.Coupon, error)

	// Coupon validation and usage
	ValidateCoupon(ctx context.Context, req *ValidateCouponRequest) (*ValidateCouponResponse, error)
	UseCoupon(ctx context.Context, couponID, userID uint64, orderAmount float64, orderID *uint64) (*entities.CouponUsage, error)

	// Coupon management
	ActivateCoupon(ctx context.Context, couponID uint64) error
	DeactivateCoupon(ctx context.Context, couponID uint64) error
	ExpireCoupon(ctx context.Context, couponID uint64) error

	// Usage tracking
	GetCouponUsage(ctx context.Context, couponID uint64, limit, offset int) ([]*entities.CouponUsage, int64, error)
	GetUserCouponUsage(ctx context.Context, userID uint64, limit, offset int) ([]*entities.CouponUsage, int64, error)
	
	// Statistics
	GetCouponStatistics(ctx context.Context, couponID uint64) (map[string]interface{}, error)
	GetCouponSystemStatistics(ctx context.Context) (map[string]interface{}, error)
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
	Valid          bool                     `json:"valid" example:"true"`
	Message        string                   `json:"message" example:"Coupon is valid"`
	DiscountAmount float64                  `json:"discount_amount" example:"5.99"`
	FinalAmount    float64                  `json:"final_amount" example:"24.00"`
	Coupon         *entities.CouponResponse `json:"coupon,omitempty"`
}