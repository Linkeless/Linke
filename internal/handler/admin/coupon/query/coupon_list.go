package query

import (
	"linke/internal/handler/admin/coupon/shared"
	"linke/internal/logger"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// CouponListHandler handles coupon listing and filtering operations
type CouponListHandler struct {
	*shared.BaseHandler
}

// NewCouponListHandler creates a new coupon list handler
func NewCouponListHandler(couponService *service.CouponService) *CouponListHandler {
	return &CouponListHandler{
		BaseHandler: shared.NewBaseHandler(couponService),
	}
}

// GetCoupons godoc
// @Summary [Admin] List all coupons
// @Description Get a paginated list of all coupons with filtering options
// @Tags Admin - Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status" Enums(active,inactive,expired)
// @Param type query string false "Filter by type" Enums(percentage,fixed_amount)
// @Param is_public query bool false "Filter by public visibility"
// @Param limit query int false "Number of items per page (1-100)" minimum(1) maximum(100)
// @Param offset query int false "Number of items to skip" minimum(0)
// @Success 200 {object} response.PaginatedResponse{data=[]model.CouponResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/coupons [get]
func (h *CouponListHandler) GetCoupons(c *gin.Context) {
	// Validate pagination parameters
	limit, offset := h.Validator.ValidatePaginationParams(c)
	
	// Validate and extract filter parameters
	filters, err := h.Validator.ValidateFilterParams(c)
	if err != nil {
		response.BadRequest(c, "Invalid filter parameters", err.Error())
		return
	}

	// Build request from query parameters and filters
	var req service.GetCouponsRequest
	req.Limit = limit
	req.Offset = offset
	
	if status, ok := filters["status"].(string); ok {
		req.Status = status
	}
	if couponType, ok := filters["type"].(string); ok {
		req.Type = couponType
	}
	if isPublic, ok := filters["is_public"].(bool); ok {
		req.IsPublic = &isPublic
	}

	// Get coupons
	coupons, totalCount, err := h.CouponService.GetCoupons(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to get coupons", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get coupons", err.Error())
		return
	}

	// Convert to response format
	couponResponses := make([]*model.CouponResponse, len(coupons))
	for i, coupon := range coupons {
		couponResponses[i] = coupon.ToResponse()
	}

	response.OKPaginated(c, "Coupons retrieved successfully", couponResponses, totalCount, req.Limit, req.Offset)
}