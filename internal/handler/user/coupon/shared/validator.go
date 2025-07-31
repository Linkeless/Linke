package coupon

import (
	"strconv"

	"linke/internal/logger"
	"linke/internal/response"

	"github.com/gin-gonic/gin"
)

// CouponValidator provides validation utilities for coupon handlers
type CouponValidator struct{}

// NewCouponValidator creates a new coupon validator
func NewCouponValidator() *CouponValidator {
	return &CouponValidator{}
}

// ValidatePaginationParams validates and parses pagination parameters
func (v *CouponValidator) ValidatePaginationParams(c *gin.Context) (limit, offset int, valid bool) {
	// Default values
	limit = 20
	offset = 0

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// Parse offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	return limit, offset, true
}

// GetUserIDFromContext extracts and validates user ID from context
func (v *CouponValidator) GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context")
		response.Unauthorized(c, "User not authenticated")
		return 0, false
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		logger.Error("Invalid user ID type in context")
		response.Unauthorized(c, "Invalid user context")
		return 0, false
	}

	return userIDUint, true
}

// ValidateCouponRequest represents the coupon validation request structure
type ValidateCouponRequest struct {
	CouponCode  string  `json:"coupon_code" binding:"required" example:"SAVE20"`
	PlanID      uint    `json:"plan_id" binding:"required" example:"1"`
	OrderAmount float64 `json:"order_amount" binding:"required,min=0" example:"29.99"`
	Currency    string  `json:"currency" binding:"required,len=3" example:"USD"`
}

// BindAndValidateCouponRequest binds and validates coupon validation request
func (v *CouponValidator) BindAndValidateCouponRequest(c *gin.Context) (*ValidateCouponRequest, bool) {
	var req ValidateCouponRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid validate coupon request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request data", err.Error())
		return nil, false
	}

	return &req, true
}