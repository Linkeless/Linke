package coupon

import (
	"linke/internal/service"
)

// BaseHandler provides common dependencies for all coupon handlers
type BaseCouponHandler struct {
	CouponService *service.CouponService
}

// NewBaseCouponHandler creates a new base coupon handler
func NewBaseCouponHandler(couponService *service.CouponService) *BaseCouponHandler {
	return &BaseCouponHandler{
		CouponService: couponService,
	}
}