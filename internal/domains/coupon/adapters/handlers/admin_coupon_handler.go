package handlers

import (
	"strconv"
	"strings"
	"time"

	"linke/internal/domains/coupon/entities"
	couponInterfaces "linke/internal/domains/coupon/usecases/interfaces"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// AdminCouponHandler handles admin coupon operations
type AdminCouponHandler struct {
	couponService couponInterfaces.CouponService
}

// NewAdminCouponHandler creates a new AdminCouponHandler
func NewAdminCouponHandler(couponService couponInterfaces.CouponService) *AdminCouponHandler {
	return &AdminCouponHandler{
		couponService: couponService,
	}
}

// CreateCouponRequest represents the request body for creating a coupon
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

// UpdateCouponRequest represents the request body for updating a coupon
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

// ToggleStatusRequest represents the request for toggling coupon status
type ToggleStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive" example:"active"`
}

// ExtendExpiryRequest represents the request for extending coupon expiry
type ExtendExpiryRequest struct {
	ExtendDays  int        `json:"extend_days,omitempty" binding:"omitempty,min=1" example:"30"`
	NewExpiry   *time.Time `json:"new_expiry,omitempty" example:"2024-12-31T23:59:59Z"`
	ExtendType  string     `json:"extend_type" binding:"required,oneof=days specific" example:"days"`
}

// BulkCreateCouponRequest represents the request for bulk coupon creation
type BulkCreateCouponRequest struct {
	CodePrefix      string     `json:"code_prefix" binding:"required,min=2,max=20" example:"BULK"`
	Count           int        `json:"count" binding:"required,min=1,max=1000" example:"100"`
	Name            string     `json:"name" binding:"required,min=1,max=100" example:"Bulk Generated Coupons"`
	Description     string     `json:"description,omitempty" binding:"max=1000" example:"Bulk generated discount coupons"`
	Type            string     `json:"type" binding:"required,oneof=percentage fixed_amount" example:"percentage"`
	Value           float64    `json:"value" binding:"required,min=0" example:"20"`
	MaxUses         int        `json:"max_uses,omitempty" binding:"min=0" example:"1"`
	MaxUsesPerUser  int        `json:"max_uses_per_user,omitempty" binding:"min=1" example:"1"`
	MinOrderAmount  float64    `json:"min_order_amount,omitempty" binding:"min=0" example:"10"`
	Currency        string     `json:"currency,omitempty" binding:"omitempty,len=3" example:"USD"`
	ValidFrom       *time.Time `json:"valid_from,omitempty" example:"2024-01-01T00:00:00Z"`
	ValidUntil      *time.Time `json:"valid_until,omitempty" example:"2024-12-31T23:59:59Z"`
	ApplicablePlans string     `json:"applicable_plans,omitempty" example:"[1,2,3]"`
	IsPublic        *bool      `json:"is_public,omitempty" example:"true"`
}

// BulkUpdateRequest represents the request for bulk operations
type BulkUpdateRequest struct {
	IDs    []uint64 `json:"ids" binding:"required,min=1,max=100"`
	Status *string  `json:"status,omitempty" binding:"omitempty,oneof=active inactive expired"`
}

// SearchCouponsRequest represents the search request
type SearchCouponsRequest struct {
	Query           string     `form:"q" binding:"omitempty,min=1,max=100"`
	Status          string     `form:"status,omitempty" binding:"omitempty,oneof=active inactive expired"`
	Type            string     `form:"type,omitempty" binding:"omitempty,oneof=percentage fixed_amount"`
	IsPublic        *bool      `form:"is_public,omitempty"`
	CreatedAfter    *time.Time `form:"created_after,omitempty"`
	CreatedBefore   *time.Time `form:"created_before,omitempty"`
	ExpiresAfter    *time.Time `form:"expires_after,omitempty"`
	ExpiresBefore   *time.Time `form:"expires_before,omitempty"`
	MinValue        *float64   `form:"min_value,omitempty" binding:"omitempty,min=0"`
	MaxValue        *float64   `form:"max_value,omitempty" binding:"omitempty,min=0"`
	MinUsed         *int       `form:"min_used,omitempty" binding:"omitempty,min=0"`
	MaxUsed         *int       `form:"max_used,omitempty" binding:"omitempty,min=0"`
	Page            int        `form:"page,omitempty" binding:"omitempty,min=1" example:"1"`
	Limit           int        `form:"limit,omitempty" binding:"omitempty,min=1,max=100" example:"10"`
}

// CreateCoupon godoc
// @Summary Create new coupon
// @Description Create a new discount coupon (Admin only)
// @Tags Admin-Coupon-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param coupon body CreateCouponRequest true "Coupon creation data"
// @Success 201 {object} response.StandardResponse{data=entities.CouponResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 409 {object} response.ConflictResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/coupons [post]
func (h *AdminCouponHandler) CreateCoupon(c *gin.Context) {
	var createReq CreateCouponRequest
	if err := c.ShouldBindJSON(&createReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Convert request to service request
	serviceReq := &couponInterfaces.CreateCouponRequest{
		Code:            createReq.Code,
		Name:            createReq.Name,
		Description:     createReq.Description,
		Type:            createReq.Type,
		Value:           createReq.Value,
		MaxUses:         createReq.MaxUses,
		MaxUsesPerUser:  createReq.MaxUsesPerUser,
		MinOrderAmount:  createReq.MinOrderAmount,
		Currency:        createReq.Currency,
		ValidFrom:       createReq.ValidFrom,
		ValidUntil:      createReq.ValidUntil,
		ApplicablePlans: createReq.ApplicablePlans,
		IsPublic:        createReq.IsPublic,
	}

	// Set default values
	if serviceReq.Currency == "" {
		serviceReq.Currency = "USD"
	}
	if serviceReq.MaxUsesPerUser == 0 {
		serviceReq.MaxUsesPerUser = 1
	}

	// Get creator ID from context (admin user)
	creatorID := uint64(1) // TODO: Get from auth context

	coupon, err := h.couponService.CreateCoupon(c.Request.Context(), creatorID, serviceReq)
	if err != nil {
		logger.Error("Admin failed to create coupon",
			logger.String("code", createReq.Code),
			logger.String("name", createReq.Name),
			logger.Error2("error", err),
		)

		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			response.Conflict(c, "Coupon with this code already exists")
			return
		}

		response.InternalServerError(c, "Failed to create coupon")
		return
	}

	logger.Info("Admin created new coupon",
		logger.Uint("coupon_id", uint(coupon.ID)),
		logger.String("code", coupon.Code),
		logger.String("admin_action", "create_coupon"),
	)

	response.Created(c, coupon.ToResponse())
}

// ListCoupons godoc
// @Summary List all coupons
// @Description Get paginated list of all coupons with filtering (Admin only)
// @Tags Admin-Coupon-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Coupon status" Enums(active,inactive,expired)
// @Param type query string false "Coupon type" Enums(percentage,fixed_amount)
// @Param is_public query bool false "Public visibility filter"
// @Success 200 {object} response.StandardListResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/coupons [get]
func (h *AdminCouponHandler) ListCoupons(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	couponType := c.Query("type")
	isPublicStr := c.Query("is_public")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var isPublic *bool
	if isPublicStr != "" {
		val, err := strconv.ParseBool(isPublicStr)
		if err == nil {
			isPublic = &val
		}
	}

	offset := (page - 1) * limit

	serviceReq := &couponInterfaces.GetCouponsRequest{
		Status:   status,
		Type:     couponType,
		IsPublic: isPublic,
		Limit:    limit,
		Offset:   offset,
	}

	coupons, total, err := h.couponService.GetCoupons(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Error("Admin failed to list coupons", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to list coupons")
		return
	}

	// Convert to responses
	var couponResponses []*entities.CouponResponse
	for _, coupon := range coupons {
		couponResponses = append(couponResponses, coupon.ToResponse())
	}

	response.SuccessList(c, couponResponses, page, limit, total)
}

// GetCoupon godoc
// @Summary Get coupon information
// @Description Get coupon details by coupon ID (Admin only)
// @Tags Admin-Coupon-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Coupon ID"
// @Success 200 {object} response.StandardResponse{data=entities.CouponResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/coupons/{id} [get]
func (h *AdminCouponHandler) GetCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid coupon ID")
		return
	}

	coupon, err := h.couponService.GetCoupon(c.Request.Context(), id)
	if err != nil {
		logger.Error("Admin failed to get coupon",
			logger.Uint("coupon_id", uint(id)),
			logger.Error2("error", err),
		)
		response.NotFound(c, "Coupon not found")
		return
	}

	response.Success(c, coupon.ToResponse())
}

// UpdateCoupon godoc
// @Summary Update coupon information
// @Description Update coupon details (Admin only)
// @Tags Admin-Coupon-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Coupon ID"
// @Param coupon body UpdateCouponRequest true "Coupon update data"
// @Success 200 {object} response.StandardResponse{data=entities.CouponResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/coupons/{id} [put]
func (h *AdminCouponHandler) UpdateCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid coupon ID")
		return
	}

	var updateReq UpdateCouponRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Convert to service request
	serviceReq := &couponInterfaces.UpdateCouponRequest{
		Name:            updateReq.Name,
		Description:     updateReq.Description,
		Type:            updateReq.Type,
		Value:           updateReq.Value,
		MaxUses:         updateReq.MaxUses,
		MaxUsesPerUser:  updateReq.MaxUsesPerUser,
		MinOrderAmount:  updateReq.MinOrderAmount,
		ValidFrom:       updateReq.ValidFrom,
		ValidUntil:      updateReq.ValidUntil,
		ApplicablePlans: updateReq.ApplicablePlans,
		Status:          updateReq.Status,
		IsPublic:        updateReq.IsPublic,
	}

	coupon, err := h.couponService.UpdateCoupon(c.Request.Context(), id, serviceReq)
	if err != nil {
		logger.Error("Admin failed to update coupon",
			logger.Uint("coupon_id", uint(id)),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to update coupon")
		return
	}

	logger.Info("Admin updated coupon",
		logger.Uint("coupon_id", uint(id)),
		logger.String("admin_action", "update_coupon"),
	)

	response.Success(c, coupon.ToResponse())
}

// DeleteCoupon godoc
// @Summary Delete coupon
// @Description Soft delete a coupon (Admin only)
// @Tags Admin-Coupon-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Coupon ID"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/coupons/{id} [delete]
func (h *AdminCouponHandler) DeleteCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid coupon ID")
		return
	}

	if err := h.couponService.DeleteCoupon(c.Request.Context(), id); err != nil {
		logger.Error("Admin failed to delete coupon",
			logger.Uint("coupon_id", uint(id)),
			logger.Error2("error", err),
		)
		response.NotFound(c, "Coupon not found")
		return
	}

	logger.Info("Admin deleted coupon",
		logger.Uint("coupon_id", uint(id)),
		logger.String("admin_action", "delete_coupon"),
	)

	response.SuccessWithMessage(c, "Coupon deleted successfully", nil)
}

// ToggleCouponStatus godoc
// @Summary Toggle coupon status
// @Description Activate or deactivate a coupon (Admin only)
// @Tags Admin-Coupon-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Coupon ID"
// @Param status body ToggleStatusRequest true "Status data"
// @Success 200 {object} response.StandardResponse{data=entities.CouponResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/coupons/{id}/toggle-status [put]
func (h *AdminCouponHandler) ToggleCouponStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid coupon ID")
		return
	}

	var statusReq ToggleStatusRequest
	if err := c.ShouldBindJSON(&statusReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var coupon *entities.Coupon
	switch statusReq.Status {
	case "active":
		err = h.couponService.ActivateCoupon(c.Request.Context(), id)
	case "inactive":
		err = h.couponService.DeactivateCoupon(c.Request.Context(), id)
	default:
		response.BadRequest(c, "Invalid status")
		return
	}

	if err != nil {
		logger.Error("Admin failed to toggle coupon status",
			logger.Uint("coupon_id", uint(id)),
			logger.String("status", statusReq.Status),
			logger.Error2("error", err),
		)
		response.NotFound(c, "Coupon not found")
		return
	}

	// Get updated coupon
	coupon, err = h.couponService.GetCoupon(c.Request.Context(), id)
	if err != nil {
		response.InternalServerError(c, "Failed to get updated coupon")
		return
	}

	logger.Info("Admin toggled coupon status",
		logger.Uint("coupon_id", uint(id)),
		logger.String("status", statusReq.Status),
		logger.String("admin_action", "toggle_coupon_status"),
	)

	response.Success(c, coupon.ToResponse())
}

// ExtendCouponExpiry godoc
// @Summary Extend coupon expiry
// @Description Extend the expiry date of a coupon (Admin only)
// @Tags Admin-Coupon-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Coupon ID"
// @Param extend body ExtendExpiryRequest true "Expiry extension data"
// @Success 200 {object} response.StandardResponse{data=entities.CouponResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/coupons/{id}/extend [put]
func (h *AdminCouponHandler) ExtendCouponExpiry(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid coupon ID")
		return
	}

	var extendReq ExtendExpiryRequest
	if err := c.ShouldBindJSON(&extendReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Get current coupon
	coupon, err := h.couponService.GetCoupon(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Coupon not found")
		return
	}

	// Calculate new expiry date
	var newExpiry *time.Time
	if extendReq.ExtendType == "days" && extendReq.ExtendDays > 0 {
		if coupon.ValidUntil != nil {
			newDate := coupon.ValidUntil.AddDate(0, 0, extendReq.ExtendDays)
			newExpiry = &newDate
		} else {
			newDate := time.Now().AddDate(0, 0, extendReq.ExtendDays)
			newExpiry = &newDate
		}
	} else if extendReq.ExtendType == "specific" && extendReq.NewExpiry != nil {
		newExpiry = extendReq.NewExpiry
	} else {
		response.BadRequest(c, "Invalid extend type or missing parameters")
		return
	}

	// Update coupon with new expiry
	updateReq := &couponInterfaces.UpdateCouponRequest{
		ValidUntil: newExpiry,
	}

	updatedCoupon, err := h.couponService.UpdateCoupon(c.Request.Context(), id, updateReq)
	if err != nil {
		logger.Error("Admin failed to extend coupon expiry",
			logger.Uint("coupon_id", uint(id)),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to extend coupon expiry")
		return
	}

	logger.Info("Admin extended coupon expiry",
		logger.Uint("coupon_id", uint(id)),
		logger.Time("new_expiry", *newExpiry),
		logger.String("admin_action", "extend_coupon"),
	)

	response.Success(c, updatedCoupon.ToResponse())
}

// GetCouponUsage godoc
// @Summary Get coupon usage details
// @Description Get detailed usage statistics for a specific coupon (Admin only)
// @Tags Admin-Coupon-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Coupon ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.StandardListResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/coupons/{id}/usage [get]
func (h *AdminCouponHandler) GetCouponUsage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid coupon ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	usageRecords, total, err := h.couponService.GetCouponUsage(c.Request.Context(), id, limit, offset)
	if err != nil {
		logger.Error("Admin failed to get coupon usage",
			logger.Uint("coupon_id", uint(id)),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to get coupon usage")
		return
	}

	// Convert to responses
	var usageResponses []*entities.CouponUsageResponse
	for _, usage := range usageRecords {
		usageResponses = append(usageResponses, usage.ToResponse())
	}

	response.SuccessList(c, usageResponses, page, limit, total)
}

// SearchCoupons godoc
// @Summary Search coupons
// @Description Search coupons by various criteria (Admin only)
// @Tags Admin-Coupon-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string false "Search query (code, name, description)"
// @Param status query string false "Coupon status" Enums(active,inactive,expired)
// @Param type query string false "Coupon type" Enums(percentage,fixed_amount)
// @Param is_public query bool false "Public visibility filter"
// @Param created_after query string false "Created after date" format(date-time)
// @Param created_before query string false "Created before date" format(date-time)
// @Param expires_after query string false "Expires after date" format(date-time)
// @Param expires_before query string false "Expires before date" format(date-time)
// @Param min_value query number false "Minimum coupon value"
// @Param max_value query number false "Maximum coupon value"
// @Param min_used query int false "Minimum usage count"
// @Param max_used query int false "Maximum usage count"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.SearchResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/coupons/search [get]
func (h *AdminCouponHandler) SearchCoupons(c *gin.Context) {
	var searchReq SearchCouponsRequest
	if err := c.ShouldBindQuery(&searchReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if searchReq.Query == "" {
		response.BadRequest(c, "Search query is required")
		return
	}

	if searchReq.Page == 0 {
		searchReq.Page = 1
	}
	if searchReq.Limit == 0 || searchReq.Limit > 100 {
		searchReq.Limit = 10
	}

	offset := (searchReq.Page - 1) * searchReq.Limit

	// Build service request
	serviceReq := &couponInterfaces.GetCouponsRequest{
		Status:   searchReq.Status,
		Type:     searchReq.Type,
		IsPublic: searchReq.IsPublic,
		Limit:    searchReq.Limit,
		Offset:   offset,
	}

	coupons, _, err := h.couponService.GetCoupons(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Error("Admin failed to search coupons",
			logger.String("query", searchReq.Query),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to search coupons")
		return
	}

	// Filter results by search query (simple text search)
	var filteredCoupons []*entities.Coupon
	query := strings.ToLower(searchReq.Query)
	for _, coupon := range coupons {
		if strings.Contains(strings.ToLower(coupon.Code), query) ||
			strings.Contains(strings.ToLower(coupon.Name), query) ||
			strings.Contains(strings.ToLower(coupon.Description), query) {
			filteredCoupons = append(filteredCoupons, coupon)
		}
	}

	// Convert to responses
	var couponResponses []*entities.CouponResponse
	for _, coupon := range filteredCoupons {
		couponResponses = append(couponResponses, coupon.ToResponse())
	}

	response.SuccessListWithExtra(c, "Search completed", couponResponses, searchReq.Page, searchReq.Limit, int64(len(filteredCoupons)), gin.H{
		"query": searchReq.Query,
	})
}

// GetCouponStatistics godoc
// @Summary Get coupon statistics
// @Description Get overall coupon system statistics (Admin only)
// @Tags Admin-Coupon-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/coupons/statistics [get]
func (h *AdminCouponHandler) GetCouponStatistics(c *gin.Context) {
	stats, err := h.couponService.GetCouponSystemStatistics(c.Request.Context())
	if err != nil {
		logger.Error("Admin failed to get coupon statistics", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get coupon statistics")
		return
	}

	response.Success(c, stats)
}

// GetCouponAnalytics godoc
// @Summary Get coupon analytics
// @Description Get detailed analytics for coupon performance (Admin only)
// @Tags Admin-Coupon-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period query string false "Analytics period" Enums(7d,30d,90d,1y) default(30d)
// @Success 200 {object} response.StandardResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/coupons/analytics [get]
func (h *AdminCouponHandler) GetCouponAnalytics(c *gin.Context) {
	period := c.DefaultQuery("period", "30d")
	
	// This would be implemented in the service layer
	analytics := gin.H{
		"period": period,
		"total_coupons": 0,
		"active_coupons": 0,
		"total_redemptions": 0,
		"total_savings": 0,
		"average_discount": 0,
		"top_performing_coupons": []gin.H{},
		"redemption_trends": []gin.H{},
		"fraud_indicators": gin.H{
			"suspicious_patterns": 0,
			"blocked_attempts": 0,
		},
	}

	logger.Info("Admin requested coupon analytics",
		logger.String("period", period),
		logger.String("admin_action", "get_coupon_analytics"),
	)

	response.Success(c, analytics)
}

// BulkCreateCoupons godoc
// @Summary Bulk create coupons
// @Description Create multiple coupons with auto-generated codes (Admin only)
// @Tags Admin-Coupon-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bulk body BulkCreateCouponRequest true "Bulk creation data"
// @Success 201 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/coupons/bulk/create [post]
func (h *AdminCouponHandler) BulkCreateCoupons(c *gin.Context) {
	var bulkReq BulkCreateCouponRequest
	if err := c.ShouldBindJSON(&bulkReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	creatorID := uint64(1) // TODO: Get from auth context
	createdCoupons := make([]*entities.Coupon, 0, bulkReq.Count)
	failedCodes := make([]string, 0)

	// Set defaults
	if bulkReq.Currency == "" {
		bulkReq.Currency = "USD"
	}
	if bulkReq.MaxUsesPerUser == 0 {
		bulkReq.MaxUsesPerUser = 1
	}
	if bulkReq.MaxUses == 0 {
		bulkReq.MaxUses = 1
	}

	for i := 0; i < bulkReq.Count; i++ {
		// Generate unique code
		code := h.generateCouponCode(bulkReq.CodePrefix, i+1)

		serviceReq := &couponInterfaces.CreateCouponRequest{
			Code:            code,
			Name:            bulkReq.Name,
			Description:     bulkReq.Description,
			Type:            bulkReq.Type,
			Value:           bulkReq.Value,
			MaxUses:         bulkReq.MaxUses,
			MaxUsesPerUser:  bulkReq.MaxUsesPerUser,
			MinOrderAmount:  bulkReq.MinOrderAmount,
			Currency:        bulkReq.Currency,
			ValidFrom:       bulkReq.ValidFrom,
			ValidUntil:      bulkReq.ValidUntil,
			ApplicablePlans: bulkReq.ApplicablePlans,
			IsPublic:        bulkReq.IsPublic,
		}

		coupon, err := h.couponService.CreateCoupon(c.Request.Context(), creatorID, serviceReq)
		if err != nil {
			logger.Error("Failed to create bulk coupon",
				logger.String("code", code),
				logger.Error2("error", err),
			)
			failedCodes = append(failedCodes, code)
			continue
		}

		createdCoupons = append(createdCoupons, coupon)
	}

	logger.Info("Admin created bulk coupons",
		logger.Int("requested_count", bulkReq.Count),
		logger.Int("created_count", len(createdCoupons)),
		logger.Int("failed_count", len(failedCodes)),
		logger.String("admin_action", "bulk_create_coupons"),
	)

	response.SuccessWithMessage(c, "Bulk coupon creation completed", gin.H{
		"requested_count": bulkReq.Count,
		"created_count":   len(createdCoupons),
		"failed_count":    len(failedCodes),
		"failed_codes":    failedCodes,
		"created_coupons": func() []*entities.CouponResponse {
			var responses []*entities.CouponResponse
			for _, coupon := range createdCoupons {
				responses = append(responses, coupon.ToResponse())
			}
			return responses
		}(),
	})
}

// BulkUpdateCoupons godoc
// @Summary Bulk update coupons
// @Description Update multiple coupons at once (Admin only)
// @Tags Admin-Coupon-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bulk body BulkUpdateRequest true "Bulk update data"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/coupons/bulk/update [post]
func (h *AdminCouponHandler) BulkUpdateCoupons(c *gin.Context) {
	var bulkReq BulkUpdateRequest
	if err := c.ShouldBindJSON(&bulkReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	successCount := 0
	failedIDs := make([]uint64, 0)

	serviceReq := &couponInterfaces.UpdateCouponRequest{
		Status: bulkReq.Status,
	}

	for _, id := range bulkReq.IDs {
		_, err := h.couponService.UpdateCoupon(c.Request.Context(), id, serviceReq)
		if err != nil {
			logger.Error("Failed to update coupon in bulk operation",
				logger.Uint("coupon_id", uint(id)),
				logger.Error2("error", err),
			)
			failedIDs = append(failedIDs, id)
			continue
		}
		successCount++
	}

	logger.Info("Admin executed bulk coupon update",
		logger.Int("requested_count", len(bulkReq.IDs)),
		logger.Int("success_count", successCount),
		logger.Int("failed_count", len(failedIDs)),
		logger.String("admin_action", "bulk_update_coupons"),
	)

	response.SuccessWithMessage(c, "Bulk coupon update completed", gin.H{
		"requested_count": len(bulkReq.IDs),
		"success_count":   successCount,
		"failed_count":    len(failedIDs),
		"failed_ids":      failedIDs,
	})
}

// BulkDeactivateCoupons godoc
// @Summary Bulk deactivate coupons
// @Description Deactivate multiple coupons at once (Admin only)
// @Tags Admin-Coupon-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bulk body BulkUpdateRequest true "Bulk deactivation data"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/coupons/bulk/deactivate [post]
func (h *AdminCouponHandler) BulkDeactivateCoupons(c *gin.Context) {
	var bulkReq BulkUpdateRequest
	if err := c.ShouldBindJSON(&bulkReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	successCount := 0
	failedIDs := make([]uint64, 0)

	for _, id := range bulkReq.IDs {
		err := h.couponService.DeactivateCoupon(c.Request.Context(), id)
		if err != nil {
			logger.Error("Failed to deactivate coupon in bulk operation",
				logger.Uint("coupon_id", uint(id)),
				logger.Error2("error", err),
			)
			failedIDs = append(failedIDs, id)
			continue
		}
		successCount++
	}

	logger.Info("Admin executed bulk coupon deactivation",
		logger.Int("requested_count", len(bulkReq.IDs)),
		logger.Int("success_count", successCount),
		logger.Int("failed_count", len(failedIDs)),
		logger.String("admin_action", "bulk_deactivate_coupons"),
	)

	response.SuccessWithMessage(c, "Bulk coupon deactivation completed", gin.H{
		"requested_count": len(bulkReq.IDs),
		"success_count":   successCount,
		"failed_count":    len(failedIDs),
		"failed_ids":      failedIDs,
	})
}

// Helper function to generate coupon codes
func (h *AdminCouponHandler) generateCouponCode(prefix string, sequence int) string {
	return strings.ToUpper(prefix) + strconv.Itoa(sequence)
}