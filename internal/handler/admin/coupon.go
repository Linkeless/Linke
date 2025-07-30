package admin

import (
	"strconv"

	"linke/internal/logger"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminCouponHandler handles admin coupon operations
type AdminCouponHandler struct {
	couponService *service.CouponService
}

// NewAdminCouponHandler creates a new admin coupon handler
func NewAdminCouponHandler(couponService *service.CouponService) *AdminCouponHandler {
	return &AdminCouponHandler{
		couponService: couponService,
	}
}

// CreateCoupon godoc
// @Summary [Admin] Create coupon
// @Description Create a new coupon with specified parameters
// @Tags Admin - Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param create_request body service.CreateCouponRequest true "Coupon creation data"
// @Success 201 {object} response.StandardResponse{data=model.CouponResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/coupons [post]
func (h *AdminCouponHandler) CreateCoupon(c *gin.Context) {
	// Get current user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context")
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req service.CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid create coupon request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Create coupon
	coupon, err := h.couponService.CreateCoupon(c.Request.Context(), uint64(userID.(uint)), &req)
	if err != nil {
		logger.Error("Failed to create coupon", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to create coupon", err.Error())
		return
	}

	logger.Info("Coupon created successfully by admin", 
		logger.Any("coupon_id", coupon.ID),
		logger.String("coupon_code", coupon.Code),
		logger.Any("admin_id", userID.(uint)))

	response.CreatedWithMessage(c, "Coupon created successfully", coupon.ToResponse())
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
func (h *AdminCouponHandler) GetCoupons(c *gin.Context) {
	var req service.GetCouponsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("Invalid get coupons request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	// Set default limit if not provided
	if req.Limit == 0 {
		req.Limit = 20
	}

	// Get coupons
	coupons, totalCount, err := h.couponService.GetCoupons(c.Request.Context(), &req)
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

// GetCoupon godoc
// @Summary [Admin] Get coupon details
// @Description Get detailed information about a specific coupon
// @Tags Admin - Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Coupon ID"
// @Success 200 {object} response.StandardResponse{data=model.CouponResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/coupons/{id} [get]
func (h *AdminCouponHandler) GetCoupon(c *gin.Context) {
	// Parse coupon ID
	couponID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		logger.Error("Invalid coupon ID", logger.Error2("error", err))
		response.BadRequest(c, "Invalid coupon ID", err.Error())
		return
	}

	// Get coupon
	coupon, err := h.couponService.GetCoupon(c.Request.Context(), couponID)
	if err != nil {
		logger.Error("Failed to get coupon", logger.Error2("error", err), logger.Any("coupon_id", uint(couponID)))
		if err.Error() == "coupon not found" {
			response.NotFound(c, "Coupon not found")
		} else {
			response.InternalServerError(c, "Failed to get coupon", err.Error())
		}
		return
	}

	response.OK(c, "Coupon retrieved successfully", coupon.ToResponse())
}

// UpdateCoupon godoc
// @Summary [Admin] Update coupon
// @Description Update an existing coupon's information
// @Tags Admin - Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Coupon ID"
// @Param update_request body service.UpdateCouponRequest true "Coupon update data"
// @Success 200 {object} response.StandardResponse{data=model.CouponResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/coupons/{id} [put]
func (h *AdminCouponHandler) UpdateCoupon(c *gin.Context) {
	// Parse coupon ID
	couponID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		logger.Error("Invalid coupon ID", logger.Error2("error", err))
		response.BadRequest(c, "Invalid coupon ID", err.Error())
		return
	}

	var req service.UpdateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid update coupon request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Update coupon
	coupon, err := h.couponService.UpdateCoupon(c.Request.Context(), couponID, &req)
	if err != nil {
		logger.Error("Failed to update coupon", logger.Error2("error", err), logger.Any("coupon_id", uint(couponID)))
		if err.Error() == "coupon not found" {
			response.NotFound(c, "Coupon not found")
		} else {
			response.InternalServerError(c, "Failed to update coupon", err.Error())
		}
		return
	}

	// Get current user ID for logging
	userID, _ := c.Get("user_id")
	logger.Info("Coupon updated successfully by admin", 
		logger.Any("coupon_id", coupon.ID),
		logger.Uint("admin_id", userID.(uint)))

	response.OK(c, "Coupon updated successfully", coupon.ToResponse())
}

// DeleteCoupon godoc
// @Summary [Admin] Delete coupon
// @Description Soft delete a coupon (it will no longer be usable)
// @Tags Admin - Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Coupon ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/coupons/{id} [delete]
func (h *AdminCouponHandler) DeleteCoupon(c *gin.Context) {
	// Parse coupon ID
	couponID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		logger.Error("Invalid coupon ID", logger.Error2("error", err))
		response.BadRequest(c, "Invalid coupon ID", err.Error())
		return
	}

	// Delete coupon
	err = h.couponService.DeleteCoupon(c.Request.Context(), couponID)
	if err != nil {
		logger.Error("Failed to delete coupon", logger.Error2("error", err), logger.Any("coupon_id", uint(couponID)))
		if err.Error() == "coupon not found" {
			response.NotFound(c, "Coupon not found")
		} else {
			response.InternalServerError(c, "Failed to delete coupon", err.Error())
		}
		return
	}

	// Get current user ID for logging
	userID, _ := c.Get("user_id")
	logger.Info("Coupon deleted successfully by admin", 
		logger.Any("coupon_id", uint(couponID)),
		logger.Uint("admin_id", userID.(uint)))

	response.OK(c, "Coupon deleted successfully", nil)
}

// GetCouponUsages godoc
// @Summary [Admin] Get coupon usage history
// @Description Get a paginated list of coupon usage records
// @Tags Admin - Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Coupon ID"
// @Param limit query int false "Number of items per page (1-100)" minimum(1) maximum(100)
// @Param offset query int false "Number of items to skip" minimum(0)
// @Success 200 {object} response.PaginatedResponse{data=[]model.CouponUsageResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/coupons/{id}/usages [get]
func (h *AdminCouponHandler) GetCouponUsages(c *gin.Context) {
	// Parse coupon ID
	couponID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		logger.Error("Invalid coupon ID", logger.Error2("error", err))
		response.BadRequest(c, "Invalid coupon ID", err.Error())
		return
	}

	// Parse query parameters
	limit := 20 // default
	offset := 0 // default

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Check if coupon exists
	_, err = h.couponService.GetCoupon(c.Request.Context(), couponID)
	if err != nil {
		logger.Error("Failed to check coupon existence", logger.Error2("error", err), logger.Uint("coupon_id", uint(couponID)))
		if err.Error() == "coupon not found" {
			response.NotFound(c, "Coupon not found")
		} else {
			response.InternalServerError(c, "Failed to check coupon", err.Error())
		}
		return
	}

	// Get coupon usages
	usages, totalCount, err := h.couponService.GetCouponUsages(c.Request.Context(), &couponID, nil, limit, offset)
	if err != nil {
		logger.Error("Failed to get coupon usages", logger.Error2("error", err), logger.Uint("coupon_id", uint(couponID)))
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

// GetCouponByCode godoc
// @Summary [Admin] Get coupon by code
// @Description Get coupon information by coupon code
// @Tags Admin - Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Coupon Code"
// @Success 200 {object} response.StandardResponse{data=model.CouponResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/coupons/code/{code} [get]
func (h *AdminCouponHandler) GetCouponByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "Coupon code is required")
		return
	}

	// Get coupon by code
	coupon, err := h.couponService.GetCouponByCode(c.Request.Context(), code)
	if err != nil {
		logger.Error("Failed to get coupon by code", logger.Error2("error", err), logger.String("code", code))
		if err.Error() == "coupon not found" {
			response.NotFound(c, "Coupon not found")
		} else {
			response.InternalServerError(c, "Failed to get coupon", err.Error())
		}
		return
	}

	response.OK(c, "Coupon retrieved successfully", coupon.ToResponse())
}