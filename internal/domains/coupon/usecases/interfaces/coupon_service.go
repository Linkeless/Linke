package interfaces

import (
	"context"

	"linke/internal/domains/coupon/dto"
	"linke/internal/domains/coupon/entities"
)

// CouponService defines coupon-specific operations
type CouponService interface {
	// Coupon-specific operations
	GetCouponByCode(ctx context.Context, code string) (*entities.Coupon, error)
	GetPublicCoupons(ctx context.Context, limit int) ([]*entities.Coupon, error)
	GetActiveCoupons(ctx context.Context) ([]*entities.Coupon, error)

	// Coupon validation and usage (domain-specific business logic)
	ValidateCoupon(ctx context.Context, req *dto.ValidateCouponRequest) (*dto.ValidateCouponResponse, error)
	UseCoupon(ctx context.Context, couponID, userID uint64, orderAmount float64, orderID *uint64) (*entities.CouponUsage, error)

	// Coupon management (extends generic status management)
	ActivateCoupon(ctx context.Context, couponID uint64) error
	DeactivateCoupon(ctx context.Context, couponID uint64) error
	ExpireCoupon(ctx context.Context, couponID uint64) error

	// Usage tracking
	GetCouponUsage(ctx context.Context, couponID uint64, limit, offset int) ([]*entities.CouponUsage, int64, error)
	GetUserCouponUsage(ctx context.Context, userID uint64, limit, offset int) ([]*entities.CouponUsage, int64, error)

	// Legacy method support for backward compatibility
	CreateCoupon(ctx context.Context, creatorID uint64, req *dto.CreateCouponRequest) (*entities.Coupon, error)
	GetCoupon(ctx context.Context, couponID uint64) (*entities.Coupon, error)
	UpdateCoupon(ctx context.Context, couponID uint64, req *dto.UpdateCouponRequest) (*entities.Coupon, error)
	DeleteCoupon(ctx context.Context, couponID uint64) error
	GetCoupons(ctx context.Context, req *dto.GetCouponsRequest) ([]*entities.Coupon, int64, error)
	GetCouponStatistics(ctx context.Context, couponID uint64) (map[string]any, error)
	GetCouponSystemStatistics(ctx context.Context) (map[string]any, error)
}

