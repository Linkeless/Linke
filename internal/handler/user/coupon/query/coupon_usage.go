package coupon

import (
	couponshared "linke/internal/handler/user/coupon/shared"
	"linke/internal/logger"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// CouponQueryHandler handles coupon query-related requests
type CouponQueryHandler struct {
	*couponshared.BaseCouponHandler
	validator *couponshared.CouponValidator
}

// NewCouponQueryHandler creates a new coupon query handler
func NewCouponQueryHandler(couponService *service.CouponService) *CouponQueryHandler {
	return &CouponQueryHandler{
		BaseCouponHandler: couponshared.NewBaseCouponHandler(couponService),
		validator:         couponshared.NewCouponValidator(),
	}
}

// GetMyCouponUsages godoc
// @Summary [User] Get my coupon usage history
// @Description Get a paginated list of current user's coupon usage records
// @Tags User - Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of items per page (1-100)" minimum(1) maximum(100)
// @Param offset query int false "Number of items to skip" minimum(0)
// @Success 200 {object} response.PaginatedResponse{data=[]model.CouponUsageResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /user/coupons/usages [get]
func (h *CouponQueryHandler) GetMyCouponUsages(c *gin.Context) {
	// Get current user ID from context
	userID, valid := h.validator.GetUserIDFromContext(c)
	if !valid {
		return
	}

	// Validate pagination parameters
	limit, offset, valid := h.validator.ValidatePaginationParams(c)
	if !valid {
		return
	}

	// Get user's coupon usages
	userID64 := uint64(userID)
	usages, totalCount, err := h.BaseCouponHandler.CouponService.GetCouponUsages(c.Request.Context(), nil, &userID64, limit, offset)
	if err != nil {
		logger.Error("Failed to get user coupon usages", logger.Error2("error", err), logger.Uint("user_id", userID))
		response.InternalServerError(c, "Failed to get coupon usages", err.Error())
		return
	}

	// Convert to response format
	usageResponses := make([]*model.CouponUsageResponse, len(usages))
	for i, usage := range usages {
		usageResponses[i] = usage.ToResponse()
	}

	response.OKPaginated(c, "Coupon usages retrieved successfully", usageResponses, totalCount, limit, offset)
}