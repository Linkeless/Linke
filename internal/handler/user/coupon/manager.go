package coupon

import (
	couponoperation "linke/internal/handler/user/coupon/operation"
	couponquery "linke/internal/handler/user/coupon/query"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// UserCouponManager manages all user coupon-related operations with modular structure
type UserCouponManager struct {
	// Sub-modules
	Operation *couponoperation.CouponOperationHandler
	Query     *couponquery.CouponQueryHandler
}

// NewUserCouponManager creates a new user coupon manager
func NewUserCouponManager(couponService *service.CouponService) *UserCouponManager {
	return &UserCouponManager{
		Operation: couponoperation.NewCouponOperationHandler(couponService),
		Query:     couponquery.NewCouponQueryHandler(couponService),
	}
}

// ============= Compatibility Methods =============
// These methods provide backward compatibility with existing code

// ValidateCoupon provides backward compatibility for coupon validation
func (m *UserCouponManager) ValidateCoupon(c *gin.Context) {
	m.Operation.ValidateCoupon(c)
}

// GetPublicCoupons provides backward compatibility for public coupon listing
func (m *UserCouponManager) GetPublicCoupons(c *gin.Context) {
	m.Operation.GetPublicCoupons(c) 
}

// GetMyCouponUsages provides backward compatibility for coupon usage history
func (m *UserCouponManager) GetMyCouponUsages(c *gin.Context) {
	m.Query.GetMyCouponUsages(c)
}