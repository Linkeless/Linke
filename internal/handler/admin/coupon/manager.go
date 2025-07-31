package coupon

import (
	"linke/internal/handler/admin/coupon/management"
	"linke/internal/handler/admin/coupon/query"
	"linke/internal/handler/admin/coupon/statistics"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminCouponManager manages all coupon-related admin handlers
type AdminCouponManager struct {
	// Sub-handlers for different coupon management aspects
	Management *management.CouponCRUDHandler
	List       *query.CouponListHandler
	Search     *query.CouponSearchHandler
	Usage      *statistics.CouponUsageHandler
}

// NewAdminCouponManager creates a new admin coupon manager with all sub-handlers
func NewAdminCouponManager(couponService *service.CouponService) *AdminCouponManager {
	return &AdminCouponManager{
		Management: management.NewCouponCRUDHandler(couponService),
		List:       query.NewCouponListHandler(couponService),
		Search:     query.NewCouponSearchHandler(couponService),
		Usage:      statistics.NewCouponUsageHandler(couponService),
	}
}

// Legacy compatibility layer - maintains the same interface as the original AdminCouponHandler
// This allows existing code to continue working without changes while using the modular structure internally

// CreateCoupon delegates to the management module
func (m *AdminCouponManager) CreateCoupon(c *gin.Context) {
	m.Management.CreateCoupon(c)
}

// GetCoupon delegates to the management module
func (m *AdminCouponManager) GetCoupon(c *gin.Context) {
	m.Management.GetCoupon(c)
}

// UpdateCoupon delegates to the management module
func (m *AdminCouponManager) UpdateCoupon(c *gin.Context) {
	m.Management.UpdateCoupon(c)
}

// DeleteCoupon delegates to the management module
func (m *AdminCouponManager) DeleteCoupon(c *gin.Context) {
	m.Management.DeleteCoupon(c)
}

// GetCoupons delegates to the list module
func (m *AdminCouponManager) GetCoupons(c *gin.Context) {
	m.List.GetCoupons(c)
}

// GetCouponByCode delegates to the search module
func (m *AdminCouponManager) GetCouponByCode(c *gin.Context) {
	m.Search.GetCouponByCode(c)
}

// GetCouponUsages delegates to the usage module
func (m *AdminCouponManager) GetCouponUsages(c *gin.Context) {
	m.Usage.GetCouponUsages(c)
}